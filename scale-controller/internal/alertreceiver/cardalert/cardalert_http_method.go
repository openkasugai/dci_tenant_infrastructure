package cardalert

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	client "sigs.k8s.io/controller-runtime/pkg/client"

	scalev1 "scale-controller/api/v1alpha1"
	"scale-controller/pkg/alertreceiver/common"
)

// Keys for annotations and labels in AlertManager webhook messages
const (
	alertManagerAlertCardAlertName    = "GPU ECC Double-Bit Error Detected"
	alertManagerAlertLabelNodeNameKey = "HostName"
	alertManagerAlertLabelGPUModelKey = "modelName"
	alertManagerAlertLabelGPUUUIDKey  = "UUID"
	alertManagerAlertLabelMIGUUIDKey  = "dra_mig_device_uuid"
)

// Annotation keys and values for CardAlert resources created by alert receiver for debug
const (
	cardAlertReceiverCreationAnnotationKey   = "webhook.alert.receiver/created-by"
	cardAlertReceiverCreationAnnotationValue = "card-alert-receiver"
)

// cardAlertHTTPMethod defines a handler for processing AlertManager webhook messages related to CardAlert.
type cardAlertHTTPMethod struct {
	client             client.Client
	logger             logr.Logger
	cardAlertNamespace string
}

// Ensure cardAlertHTTPMethod implements the AlertsHTTPMethod interface.
var _ common.HandlerHTTPMethod = &cardAlertHTTPMethod{}

// newCardAlertHandler is a function to create Kubernetes client for workload cluster from in-cluster configuration.
func newCardAlertHandler(
	options *CardAlertReceiverNewOptions,
) (*cardAlertHTTPMethod, error) {
	if options == nil || options.AlertReceiverNewOptions.Logger.GetSink() == nil {
		return nil, errors.New("expected options or Logger in options is not nil, but got nil")
	}

	// Get CardAlert namespace from options and validate it
	cardAlertNamespace := options.CardAlertNamespace
	if cardAlertNamespace == "" {
		return nil, errors.New("expected CardAlertNamespace in options to be not empty, but got empty")
	}

	// Use scheme with scalev1 (for CardAlert) and corev1 (for Namespace)
	scheme := runtime.NewScheme()
	if err := scalev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add scale-controller scheme: %w", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add corev1 scheme: %w", err)
	}

	// Get in-cluster config
	kubeconfig, err := InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	// Create in-cluster client
	clientInCluster, err := NewClient(kubeconfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create client for In-Cluster: %w", err)
	}

	return &cardAlertHTTPMethod{
		client:             clientInCluster,
		logger:             options.AlertReceiverNewOptions.Logger,
		cardAlertNamespace: cardAlertNamespace,
	}, nil
}

// Get is a function to handle GET requests to the /alerts endpoint.
// Alert Manager expects an 200 OK response, otherwise send_resolved will never work
func (h *cardAlertHTTPMethod) Get(
	w http.ResponseWriter, _ *http.Request,
) {
	h.logger.Info("Received GET request")

	enc := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := enc.Encode("OK"); err != nil {
		h.logger.Error(err, "failed to encode messages for GET response")
	}
}

// Post is a function to handle POST requests to the /alerts endpoint.
func (h *cardAlertHTTPMethod) Post(
	w http.ResponseWriter, req *http.Request,
) {
	h.logger.Info("Received POST request")

	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()

	message := &common.AlertManagerHookMessage{}
	if err := dec.Decode(message); err != nil {
		h.logger.Error(err, "failed to decode message")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Process the received alert message for CardAlert
	if err := h.createCardAlertResource(req.Context(), message); err != nil {
		h.logger.Error(err, "failed to create CardAlert resource")
		http.Error(w, "error happened for creating CardAlert resource", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// createCardAlertResource is a function to create CardAlert resources
// based on the received AlertManager webhook message for CardAlert.
func (h *cardAlertHTTPMethod) createCardAlertResource(
	ctx context.Context, message *common.AlertManagerHookMessage,
) error {
	// Parse the received alert message and extract scaling information
	scalingInfo := h.extractCardAlertInformation(message)
	if len(scalingInfo) == 0 {
		h.logger.Info("No valid alert found in the received message, no CardAlert resource will be created")
		return nil
	}
	h.logger.Info("Valid alerts found in the received message, start to create CardAlert resources")

	// Create CardAlert resources based on the extracted scaling information
	var err error
	for nodeName, deviceInfo := range scalingInfo {
		// Create CardAlert resource
		ca := &scalev1.CardAlert{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "card-alert-receiver-",
				Namespace:    h.cardAlertNamespace,
				Annotations: map[string]string{
					cardAlertReceiverCreationAnnotationKey: cardAlertReceiverCreationAnnotationValue,
				},
			},
			Spec: scalev1.CardAlertSpec{
				NodeName: nodeName,
				Cards:    deviceInfo,
				Source:   scalev1.WebhookReceiverOnTenant,
				Reason:   alertManagerAlertCardAlertName,
			},
		}
		if createErr := h.client.Create(ctx, ca); createErr != nil {
			createErr = fmt.Errorf("failed to create CardAlert deviceInfo['%s']='%v': %w", nodeName, deviceInfo, createErr)
			err = fmt.Errorf("%w: %w", createErr, err)
			// Continue to create other CardAlert resources with appending errors, although an error occurs
			continue
		}
		h.logger.Info("Created CardAlert resource",
			"namespace", ca.Namespace, "name", ca.Name, "nodeName", nodeName, "deviceInfo", deviceInfo)
	}
	return err
}

// extractCardAlertInformation is a function to parse the received Prometheus AlertManager webhook message
// and extract scaling information from alerts in webhook message.
func (h *cardAlertHTTPMethod) extractCardAlertInformation(
	message *common.AlertManagerHookMessage,
) map[string][]scalev1.CardInfo {
	// scalingInfo is a map with key of node name and value of slice of CardInfo for the node
	scalingInfo := make(map[string][]scalev1.CardInfo)
	for idx := range message.Alerts {
		alertName := message.Alerts[idx].Labels[common.AlertManagerAlertLabelAlertNameKey]
		alertStatus := message.Alerts[idx].Status

		if alertName != alertManagerAlertCardAlertName || alertStatus == common.AlertManagerAlertStatusResolved {
			// Skip non-CardAlert alerts or resolved alerts
			continue
		}

		nodeName, isFoundNodeName := message.Alerts[idx].Labels[alertManagerAlertLabelNodeNameKey]
		model, isFoundModel := message.Alerts[idx].Labels[alertManagerAlertLabelGPUModelKey]
		uuid, isFoundUUID := message.Alerts[idx].Labels[alertManagerAlertLabelGPUUUIDKey]
		migUUID, isFoundMIGUUID := message.Alerts[idx].Labels[alertManagerAlertLabelMIGUUIDKey]

		if !isFoundNodeName || !isFoundModel || !isFoundUUID {
			// Found invalid alert without required labels and this alert will be ignored in creating CardAlert resource,
			// but continue to process other alerts
			h.logger.Error(nil, "expected labels are missing in alert",
				"labels", message.Alerts[idx].Labels,
				"nodeNameKey", alertManagerAlertLabelNodeNameKey,
				"modelKey", alertManagerAlertLabelGPUModelKey,
				"uuidKey", alertManagerAlertLabelGPUUUIDKey,
			)
			continue
		}

		var si *scalev1.CardInfo
		for i := range scalingInfo[nodeName] {
			si = &scalingInfo[nodeName][i]
			if si.Model == model && si.GPUUUID == uuid {
				// If the same card info already exists for the node, add MIG UUID to the existing card info if MIG UUID is found,
				// and skip adding new card info
				if isFoundMIGUUID && !slices.Contains(si.MIGUUIDs, migUUID) {
					si.MIGUUIDs = append(si.MIGUUIDs, migUUID)
				}
				break
			}
			si = nil
		}
		if si == nil {
			// If the same card info does not exist for the node, create new card info and add to scalingInfo
			additionalInfo := scalev1.CardInfo{
				Model:   model,
				GPUUUID: uuid,
			}
			if isFoundMIGUUID {
				additionalInfo.MIGUUIDs = append(additionalInfo.MIGUUIDs, migUUID)
			}
			scalingInfo[nodeName] = append(scalingInfo[nodeName], additionalInfo)
		}
	}
	return scalingInfo
}
