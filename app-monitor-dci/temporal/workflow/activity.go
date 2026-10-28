package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"temporal/shared"

	examplecomv1 "temporal/api/v1"

	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	resourcev1beta1 "k8s.io/api/resource/v1beta1"
	resourcev1beta2 "k8s.io/api/resource/v1beta2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PrometheusResponse represents the JSON response structure from Prometheus API
// Contains status, result type, and metric data with values
type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []any             `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

const maxResponseSize int64 = 10 * 1024 * 1024

// indirection variables for easier testing
var (
	checkPodResourceClaimFunc     = checkPodResourceClaim
	checkPodGPUFromPrometheusFunc = checkPodGPUFromPrometheus
	inClusterConfigFunc           = rest.InClusterConfig
	newForConfigFunc              = func(cfg *rest.Config) (kubernetes.Interface, error) {
		return kubernetes.NewForConfig(cfg)
	}

	httpGetFunc = http.Get
	httpDoFunc  = func(req *http.Request) (*http.Response, error) {
		return http.DefaultClient.Do(req)
	}
	jsonMarshalFunc = json.Marshal
)

var ctrlSetControllerReferenceFunc = ctrl.SetControllerReference

// VectorSignalActivity is an Activity function that receives a signal from Vector->Argo Events, fetches metric values from Prometheus, passes the podUpdateStatus struct data holding the latest values in the workflow to ArgoEvents,
// and also updates the podUpdateStatus struct data.
func VectorSignalActivity(ctx context.Context, k8sClient client.Client, podState shared.PodUpdateStatus, sch *runtime.Scheme) (*shared.PodUpdateStatus, error) {

	klog.Info("vector signal activity started")

	mapUsages := make(map[string]string)

	// Get Prometheus URL from environment variable or use default cluster service
	prometheusURL := os.Getenv("PROMETHEUS_URL")
	if prometheusURL == "" {
		klog.Warning("PROMETHEUS_URL environment variable is not set. Prometheus client will use the default URL: http://prometheus-kube-prometheus-prometheus.prometheus.svc.cluster.local:9090.")
		prometheusURL = "http://prometheus-kube-prometheus-prometheus.prometheus.svc.cluster.local:9090"
	}

	// Get cpuUtilRange URL from environment
	cpuUtilRange := os.Getenv("CPU_UTIL_RANGE")
	if cpuUtilRange == "" {
		klog.Warning("CPU_UTIL_RANGE environment variable is not set. Using default value: 5m.")
		cpuUtilRange = "5m"
	}

	// Prometheus query endpoint path
	queryPath := "/api/v1/query"

	// Build CPU utilization query: calculates rate of CPU usage over 5 minutes
	// Excludes idle containers and groups by namespace, pod, and container
	cpuUtilQuery := fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{container!="", container!="POD", pod="%s", namespace="%s"}[%s]))by(namespace, pod)/sum(kube_pod_container_resource_limits{resource="cpu", unit="core", pod="%s", namespace="%s"})by(namespace, pod)`, podState.Metadata.Name, podState.Metadata.Namespace, cpuUtilRange, podState.Metadata.Name, podState.Metadata.Namespace)

	// Use net/url.Values for automatic URL encoding
	params := url.Values{}
	params.Add("query", cpuUtilQuery)

	// Build the full URL for the CPU utilization query
	cpuTotalUrl := fmt.Sprintf("%s%s?%s", prometheusURL, queryPath, params.Encode())

	klog.Infof("cpuTotalUrl:%s", cpuTotalUrl)
	klog.Infof("Sending request to Prometheus for CPU metrics: namespace=%s, pod=%s", podState.Metadata.Namespace, podState.Metadata.Name)
	// Make HTTP GET request to Prometheus for CPU utilization
	cpuResp, err := httpGetFunc(cpuTotalUrl)
	if err != nil {
		klog.Warningf("failed to query cpu usage: %v", err)
		mapUsages["cpuUtilization"] = "N/A"
	} else {
		klog.Infof("Received response from Prometheus for CPU metrics: namespace=%s, pod=%s", podState.Metadata.Namespace, podState.Metadata.Name)
		defer func() { _ = cpuResp.Body.Close() }()
		//debug
		klog.Infof("CPU response status for %s: %s (Code: %d)", cpuTotalUrl, cpuResp.Status, cpuResp.StatusCode)
		klog.Infof("CPU response headers: %+v", cpuResp.Header)

		cpuBody, err := io.ReadAll(io.LimitReader(cpuResp.Body, maxResponseSize))
		if err != nil {
			klog.Warningf("failed to read CPU response body: %v", err)
			mapUsages["cpuUtilization"] = "N/A"
		} else if int64(len(cpuBody)) >= maxResponseSize {
			klog.Warningf("CPU response body too large (>= 10MB), truncated")
			mapUsages["cpuUtilization"] = "N/A"
		} else {
			cpuValue, err := ExactPrometheusValue(cpuBody)
			if err != nil {
				klog.Warningf("failed to extract CPU value: %v", err)
				cpuValue = "N/A"
			}
			mapUsages["cpuUtilization"] = cpuValue
		}
	}

	// Check if the pod has GPU resources allocated by querying Kubernetes metrics
	hasGPU, err := CheckPodGPURequest(prometheusURL, podState.Metadata.Namespace, podState.Metadata.Name)
	if err != nil {
		if strings.Contains(err.Error(), "prometheus returned error status") {
			klog.Warningf("Failed to check GPU allocation (temporary issue): %v, namespace=%s, pod=%s", err, podState.Metadata.Namespace, podState.Metadata.Name)
		} else {
			klog.Warningf("Failed to check GPU allocation (invalid response): %v, namespace=%s, pod=%s", err, podState.Metadata.Namespace, podState.Metadata.Name)
		}
		// Set all GPU metrics to N/A when check fails
		mapUsages["gpuTotalUtilization"] = "N/A"
		mapUsages["gpuSMUtilization"] = "N/A"
		mapUsages["gpuGraphicUtilization"] = "N/A"
		mapUsages["gpuMemoryUtilization"] = "N/A"

	}
	if !hasGPU {
		klog.Info("GPU not allocated")
		// Set all GPU metrics to N/A when no GPU is allocated
		mapUsages["gpuTotalUtilization"] = "N/A"
		mapUsages["gpuSMUtilization"] = "N/A"
		mapUsages["gpuGraphicUtilization"] = "N/A"
		mapUsages["gpuMemoryUtilization"] = "N/A"

	}

	if hasGPU {
		// DCGM_FI_DEV_GPU_UTIL: Overall GPU utilization
		gpuTotalQuery := fmt.Sprintf(`sum(DCGM_FI_DEV_GPU_UTIL{exported_container!="", exported_pod="%s", exported_namespace="%s"}) by (exported_namespace, exported_pod)`, podState.Metadata.Name, podState.Metadata.Namespace)

		// Use net/url.Values for automatic URL encoding
		gpu_total_params := url.Values{}
		gpu_total_params.Add("query", gpuTotalQuery)
		// Build the full URL for the GPU total utilization query
		gpuTotalUrl := fmt.Sprintf("%s%s?%s", prometheusURL, queryPath, gpu_total_params.Encode())
		klog.Infof("Sending request to Prometheus for GPU Total metrics: namespace=%s, pod=%s", podState.Metadata.Namespace, podState.Metadata.Name)
		// Make HTTP GET request to Prometheus for GPU total utilization
		gpuTotalResp, err := httpGetFunc(gpuTotalUrl)

		if err != nil {
			klog.Warningf("failed to query gpu total usage: %v", err)
			mapUsages["gpuTotalUtilization"] = "N/A"
		} else {
			klog.Infof("Received response from Prometheus for GPU Total metrics: namespace=%s, pod=%s", podState.Metadata.Namespace, podState.Metadata.Name)
			gpuTotalBody, err := io.ReadAll(io.LimitReader(gpuTotalResp.Body, maxResponseSize))
			if err != nil {
				klog.Warningf("failed to read GPU Total response body: %v", err)
				mapUsages["gpuTotalUtilization"] = "N/A"
			} else if int64(len(gpuTotalBody)) >= maxResponseSize {
				klog.Warningf("GPU Total response body too large (>= 10MB), truncated")
				mapUsages["gpuTotalUtilization"] = "N/A"
			} else {
				gpuTotalValue, err := ExactPrometheusValue(gpuTotalBody)
				if err != nil {
					klog.Warningf("failed to extract GPU Total value: %v. If a single GPU is allocated, this is abnormal. If a MIG GPU is allocated, this is normal.", err)
					gpuTotalValue = "N/A"
				}
				mapUsages["gpuTotalUtilization"] = gpuTotalValue
			}
		}
		// DCGM_FI_PROF_SM_ACTIVE: GPU Streaming Multiprocessor utilization
		gpuSMQuery := fmt.Sprintf(`sum(DCGM_FI_PROF_SM_ACTIVE{exported_container!="", exported_pod="%s", exported_namespace="%s"}) by (exported_namespace, exported_pod)`, podState.Metadata.Name, podState.Metadata.Namespace)

		// Use net/url.Values for automatic URL encoding
		gpu_sm_params := url.Values{}
		gpu_sm_params.Add("query", gpuSMQuery)
		// Build the full URL for the GPU SM utilization query
		gpuSMUrl := fmt.Sprintf("%s%s?%s", prometheusURL, queryPath, gpu_sm_params.Encode())
		klog.Infof("Sending request to Prometheus for GPU SM metrics: namespace=%s, pod=%s", podState.Metadata.Namespace, podState.Metadata.Name)
		// Make HTTP GET request to Prometheus for GPU SM utilization
		gpuSMResp, err := httpGetFunc(gpuSMUrl)

		if err != nil {
			klog.Warningf("failed to query gpu sm usage: %v", err)
			mapUsages["gpuSMUtilization"] = "N/A"
		} else {
			klog.Infof("Received response from Prometheus for GPU SM metrics: namespace=%s, pod=%s", podState.Metadata.Namespace, podState.Metadata.Name)
			defer func() { _ = gpuSMResp.Body.Close() }()
			gpuSMBody, err := io.ReadAll(io.LimitReader(gpuSMResp.Body, maxResponseSize))
			if err != nil {
				klog.Warningf("failed to read GPU SM response body: %v", err)
				mapUsages["gpuSMUtilization"] = "N/A"
			} else if int64(len(gpuSMBody)) >= maxResponseSize {
				klog.Warningf("GPU SM response body too large (>= 10MB), truncated")
				mapUsages["gpuSMUtilization"] = "N/A"
			} else {
				gpuSMValue, err := ExactPrometheusValue(gpuSMBody)
				if err != nil {
					klog.Warningf("failed to extract GPU SM usage value: %v", err)
					gpuSMValue = "N/A"
				}
				mapUsages["gpuSMUtilization"] = gpuSMValue

			}
		}
		// DCGM_FI_PROF_GR_ENGINE_ACTIVE: GPU Graphics Engine utilization
		gpuGRQuery := fmt.Sprintf(`sum(DCGM_FI_PROF_GR_ENGINE_ACTIVE{exported_container!="", exported_pod="%s", exported_namespace="%s"}) by (exported_namespace, exported_pod)`, podState.Metadata.Name, podState.Metadata.Namespace)

		// Use net/url.Values for automatic URL encoding
		gpu_gr_params := url.Values{}
		gpu_gr_params.Add("query", gpuGRQuery)
		// Build the full URL for the GPU Graphics Engine utilization query
		gpuGRUrl := fmt.Sprintf("%s%s?%s", prometheusURL, queryPath, gpu_gr_params.Encode())
		// Make HTTP GET request to Prometheus for GPU Graphics Engine utilization
		klog.Infof("Sending request to Prometheus for GPU Graphic metrics: namespace=%s, pod=%s", podState.Metadata.Namespace, podState.Metadata.Name)
		gpuGRResp, err := httpGetFunc(gpuGRUrl)

		if err != nil {
			klog.Warningf("failed to query gpu graphic usage: %v", err)
			mapUsages["gpuGraphicUtilization"] = "N/A"
		} else {
			klog.Infof("Received response from Prometheus for GPU Graphic metrics: namespace=%s, pod=%s", podState.Metadata.Namespace, podState.Metadata.Name)
			defer func() { _ = gpuGRResp.Body.Close() }()
			gpuGRBody, err := io.ReadAll(io.LimitReader(gpuGRResp.Body, maxResponseSize))
			if err != nil {
				klog.Warningf("failed to read GPU Graphic response body: %v", err)
				mapUsages["gpuGraphicUtilization"] = "N/A"
			} else if int64(len(gpuGRBody)) >= maxResponseSize {
				klog.Warningf("GPU Graphic response body too large (>= 10MB), truncated")
				mapUsages["gpuGraphicUtilization"] = "N/A"
			} else {
				gpuGRValue, err := ExactPrometheusValue(gpuGRBody)
				if err != nil {
					klog.Warningf("failed to extract GPU Graphic usage value: %v", err)
					gpuGRValue = "N/A"
				}
				mapUsages["gpuGraphicUtilization"] = gpuGRValue
			}
		}
		// DCGM_FI_DEV_FB_USED / DCGM_FI_DEV_FB_TOTAL: GPU Memory utilization
		gpuMemQuery := fmt.Sprintf(`sum(DCGM_FI_DEV_FB_USED{exported_container!="", exported_pod="%s", exported_namespace="%s"}) by (exported_namespace, exported_pod)/sum(DCGM_FI_DEV_FB_TOTAL{exported_container!="", exported_pod="%s", exported_namespace="%s"}) by (exported_namespace, exported_pod)`, podState.Metadata.Name, podState.Metadata.Namespace, podState.Metadata.Name, podState.Metadata.Namespace)

		// Use net/url.Values for automatic URL encoding
		gpu_mem_params := url.Values{}
		gpu_mem_params.Add("query", gpuMemQuery)
		// Build the full URL for the GPU Memory utilization query
		gpuMemUrl := fmt.Sprintf("%s%s?%s", prometheusURL, queryPath, gpu_mem_params.Encode())
		klog.Infof("Sending request to Prometheus for GPU Memory metrics: namespace=%s, pod=%s", podState.Metadata.Namespace, podState.Metadata.Name)

		// Make HTTP GET request to Prometheus for GPU Memory utilization
		gpuMemResp, err := httpGetFunc(gpuMemUrl)

		if err != nil {
			klog.Warningf("failed to query gpu memory usage: %v", err)
			mapUsages["gpuMemoryUtilization"] = "N/A"
		} else {
			klog.Infof("Received response from Prometheus for GPU Memory metrics: namespace=%s, pod=%s", podState.Metadata.Namespace, podState.Metadata.Name)
			defer func() { _ = gpuMemResp.Body.Close() }()
			gpuMemBody, err := io.ReadAll(io.LimitReader(gpuMemResp.Body, maxResponseSize))
			if err != nil {
				klog.Warningf("failed to read GPU Memory response body: %v", err)
				mapUsages["gpuMemoryUtilization"] = "N/A"
			} else if int64(len(gpuMemBody)) >= maxResponseSize {
				klog.Warningf("GPU Memory response body too large (>= 10MB), truncated")
				mapUsages["gpuMemoryUtilization"] = "N/A"
			} else {
				gpuMemValue, err := ExactPrometheusValue(gpuMemBody)
				if err != nil {
					klog.Warningf("failed to extract GPU Memory usage value: %v", err)
					gpuMemValue = "N/A"
				}
				mapUsages["gpuMemoryUtilization"] = gpuMemValue
			}
		}
	}
	// Marshal the updated Status.AppStatus.Usages to JSON
	podState.Status.AppStatus.Usages, err = jsonMarshalFunc(mapUsages)
	if err != nil {
		klog.Errorf("failed to marshal Usages %s/%s: %v", podState.Metadata.Namespace, podState.Metadata.Name, err)
		return nil, fmt.Errorf("failed to marshal Usages: %w", err)
	}

	// Directly create or update the DCIAppStateCR
	err = PodStateCrateOrPatch(ctx, k8sClient, podState, sch)
	if err != nil {
		klog.Errorf("failed to create or patch podStatus %s/%s: %v", podState.Metadata.Namespace, podState.Metadata.Name, err)
		return nil, fmt.Errorf("failed to create or patch podStatus: %w", err)
	}

	klog.Infof("Vector signal activity: Successfully updated the DCIAppStateCR: namespace=%s, pod=%s", podState.Metadata.Namespace, podState.Metadata.Name)
	// Return the updated podState
	return &podState, nil
}

// AlertManagerSignalActivity is an Activity function that receives a signal from Alertmanager->Argo Events, fetches metric values from Prometheus, passes the podUpdateStatus struct data holding the latest values in the workflow to ArgoEvents,
// and also updates the podUpdateStatus struct data.
func AlertManagerSignalActivity(ctx context.Context, k8sClient client.Client, podState shared.PodUpdateStatus, sch *runtime.Scheme, alertNamespace string, alertPod string) (*shared.PodUpdateStatus, error) {
	klog.Info("alertmanager signal activity started")

	mapUsages := make(map[string]string)

	// Get Prometheus URL from environment variable or use default cluster service
	prometheusURL := os.Getenv("PROMETHEUS_URL")
	if prometheusURL == "" {
		klog.Warning("PROMETHEUS_URL environment variable is not set. Prometheus client will use the default URL: http://prometheus-kube-prometheus-prometheus.prometheus.svc.cluster.local:9090.")
		prometheusURL = "http://prometheus-kube-prometheus-prometheus.prometheus.svc.cluster.local:9090"
	}

	// Get cpuUtilRange URL from environment
	cpuUtilRange := os.Getenv("CPU_UTIL_RANGE")
	if cpuUtilRange == "" {
		klog.Warning("CPU_UTIL_RANGE environment variable is not set. Using default value: 5m.")
		cpuUtilRange = "5m"
	}

	// Prometheus query endpoint path
	queryPath := "/api/v1/query"

	// Build the full URL for the CPU utilization query
	cpuUtilQuery := fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{container!="", container!="POD", pod="%s", namespace="%s"}[%s]))by(namespace, pod)/sum(kube_pod_container_resource_limits{resource="cpu", unit="core", pod="%s", namespace="%s"})by(namespace, pod)`, podState.Metadata.Name, podState.Metadata.Namespace, cpuUtilRange, podState.Metadata.Name, podState.Metadata.Namespace)

	// Use net/url.Values for automatic URL encoding
	params := url.Values{}
	params.Add("query", cpuUtilQuery)

	// Build the full URL for the CPU utilization query
	cpuTotalUrl := fmt.Sprintf("%s%s?%s", prometheusURL, queryPath, params.Encode())

	klog.Infof("cpuTotalUrl:%s", cpuTotalUrl)
	klog.Infof("Sending request to Prometheus for CPU metrics: namespace=%s, pod=%s", alertNamespace, alertPod)
	// Make HTTP GET request to Prometheus for CPU utilization
	cpuResp, err := httpGetFunc(cpuTotalUrl)

	if err != nil {
		klog.Warningf("failed to query cpu usage: %v", err)
		mapUsages["cpuUtilization"] = "N/A"
	} else {
		klog.Infof("Received response from Prometheus for CPU metrics: namespace=%s, pod=%s", alertNamespace, alertPod)
		defer func() { _ = cpuResp.Body.Close() }()
		cpuBody, err := io.ReadAll(io.LimitReader(cpuResp.Body, maxResponseSize))
		if err != nil {
			klog.Warningf("failed to read CPU response body: %v", err)
			mapUsages["cpuUtilization"] = "N/A"
		} else if int64(len(cpuBody)) >= maxResponseSize {
			klog.Warningf("CPU response body too large (>= 10MB), truncated")
			mapUsages["cpuUtilization"] = "N/A"
		} else {
			cpuValue, err := ExactPrometheusValue(cpuBody)
			if err != nil {
				klog.Warningf("failed to extract CPU value: %v", err)
				cpuValue = "N/A"
			}
			mapUsages["cpuUtilization"] = cpuValue
		}
	}
	// Check if the pod has GPU resources allocated by querying Kubernetes metrics
	hasGPU, err := CheckPodGPURequest(prometheusURL, alertNamespace, alertPod)
	if err != nil {
		if strings.Contains(err.Error(), "prometheus returned error status") {
			klog.Warningf("Failed to check GPU allocation (temporary issue): %v, namespace=%s, pod=%s", err, alertNamespace, alertPod)
		} else {
			klog.Warningf("Failed to check GPU allocation (invalid response): %v, namespace=%s, pod=%s", err, alertNamespace, alertPod)
		}
		// Set all GPU metrics to N/A when check fails
		mapUsages["gpuTotalUtilization"] = "N/A"
		mapUsages["gpuSMUtilization"] = "N/A"
		mapUsages["gpuGraphicUtilization"] = "N/A"
		mapUsages["gpuMemoryUtilization"] = "N/A"

	}
	if !hasGPU {
		klog.Infof("GPU not allocated")
		// Set all GPU metrics to N/A when no GPU is allocated
		mapUsages["gpuTotalUtilization"] = "N/A"
		mapUsages["gpuSMUtilization"] = "N/A"
		mapUsages["gpuGraphicUtilization"] = "N/A"
		mapUsages["gpuMemoryUtilization"] = "N/A"

	}

	if hasGPU {
		gpuTotalQuery := fmt.Sprintf(`sum(DCGM_FI_DEV_GPU_UTIL{exported_container!="", exported_pod="%s", exported_namespace="%s"}) by (exported_namespace, exported_pod)`, alertPod, alertNamespace)

		// Use net/url.Values for automatic URL encoding
		gpu_total_params := url.Values{}
		gpu_total_params.Add("query", gpuTotalQuery)
		// Build the full URL for the GPU total utilization query
		gpuTotalUrl := fmt.Sprintf("%s%s?%s", prometheusURL, queryPath, gpu_total_params.Encode())
		klog.Infof("Sending request to Prometheus for GPU total metrics: namespace=%s, pod=%s", alertNamespace, alertPod)

		// Make HTTP GET request to Prometheus for GPU total utilization
		gpuTotalResp, err := httpGetFunc(gpuTotalUrl)
		if err != nil {
			klog.Warningf("failed to query gpu total usage: %v", err)
			mapUsages["gpuTotalUtilization"] = "N/A"
		} else {
			klog.Infof("Received response from Prometheus for GPU total metrics: namespace=%s, pod=%s", alertNamespace, alertPod)
			defer func() { _ = gpuTotalResp.Body.Close() }()
			gpuTotalBody, err := io.ReadAll(io.LimitReader(gpuTotalResp.Body, maxResponseSize))
			if err != nil {
				klog.Warningf("failed to read GPU total response body: %v", err)
				mapUsages["gpuTotalUtilization"] = "N/A"
			} else if int64(len(gpuTotalBody)) >= maxResponseSize {
				klog.Warningf("GPU total response body too large (>= 10MB), truncated")
				mapUsages["gpuTotalUtilization"] = "N/A"
			} else {
				gpuTotalValue, err := ExactPrometheusValue(gpuTotalBody)
				if err != nil {
					klog.Warningf("failed to extract GPU Total value: %v. If a single GPU is allocated, this is abnormal. If a MIG GPU is allocated, this is normal.", err)
					gpuTotalValue = "N/A"
				}
				mapUsages["gpuTotalUtilization"] = gpuTotalValue
			}
		}
		// DCGM_FI_PROF_SM_ACTIVE: GPU Streaming Multiprocessor utilization
		gpuSMQuery := fmt.Sprintf(`sum(DCGM_FI_PROF_SM_ACTIVE{exported_container!="", exported_pod="%s", exported_namespace="%s"}) by (exported_namespace, exported_pod)`, alertPod, alertNamespace)

		// Use net/url.Values for automatic URL encoding
		gpu_sm_params := url.Values{}
		gpu_sm_params.Add("query", gpuSMQuery)
		// Build the full URL for the GPU SM utilization query
		gpuSMUrl := fmt.Sprintf("%s%s?%s", prometheusURL, queryPath, gpu_sm_params.Encode())
		klog.Infof("Sending request to Prometheus for GPU SM metrics: namespace=%s, pod=%s", alertNamespace, alertPod)
		// Make HTTP GET request to Prometheus for GPU SM utilization
		gpuSMResp, err := httpGetFunc(gpuSMUrl)

		if err != nil {
			klog.Warningf("failed to query gpu sm usage: %v", err)
			mapUsages["gpuSMUtilization"] = "N/A"
		} else {
			klog.Infof("Received response from Prometheus for GPU SM metrics: namespace=%s, pod=%s", alertNamespace, alertPod)
			defer func() { _ = gpuSMResp.Body.Close() }()
			gpuSMBody, err := io.ReadAll(io.LimitReader(gpuSMResp.Body, maxResponseSize))
			if err != nil {
				klog.Warningf("failed to read GPU SM response body: %v", err)
				mapUsages["gpuSMUtilization"] = "N/A"
			} else if int64(len(gpuSMBody)) >= maxResponseSize {
				klog.Warningf("GPU SM response body too large (>= 10MB), truncated")
				mapUsages["gpuSMUtilization"] = "N/A"
			} else {
				gpuSMValue, err := ExactPrometheusValue(gpuSMBody)
				if err != nil {
					klog.Warningf("failed to extract GPU SM value: %v", err)
					gpuSMValue = "N/A"
				}
				mapUsages["gpuSMUtilization"] = gpuSMValue
			}
		}
		// DCGM_FI_PROF_GR_ENGINE_ACTIVE: GPU Graphics Engine utilization
		gpuGRQuery := fmt.Sprintf(`sum(DCGM_FI_PROF_GR_ENGINE_ACTIVE{exported_container!="", exported_pod="%s", exported_namespace="%s"}) by (exported_namespace, exported_pod)`, alertPod, alertNamespace)
		// Use net/url.Values for automatic URL encoding
		gpu_gr_params := url.Values{}
		gpu_gr_params.Add("query", gpuGRQuery)
		// Build the full URL for the GPU Graphics Engine utilization query
		gpuGRUrl := fmt.Sprintf("%s%s?%s", prometheusURL, queryPath, gpu_gr_params.Encode())
		klog.Infof("Sending request to Prometheus for GPU Graphic metrics: namespace=%s, pod=%s", alertNamespace, alertPod)
		// Make HTTP GET request to Prometheus for GPU Graphics Engine utilization
		gpuGRResp, err := httpGetFunc(gpuGRUrl)

		if err != nil {
			klog.Warningf("failed to query gpu graphic usage: %v", err)
			mapUsages["gpuGraphicUtilization"] = "N/A"
		} else {
			klog.Infof("Received response from Prometheus for GPU Graphic metrics: namespace=%s, pod=%s", alertNamespace, alertPod)
			defer func() { _ = gpuGRResp.Body.Close() }()
			gpuGRBody, err := io.ReadAll(io.LimitReader(gpuGRResp.Body, maxResponseSize))
			if err != nil {
				klog.Warningf("failed to read GPU Graphic response body: %v", err)
				mapUsages["gpuGraphicUtilization"] = "N/A"
			} else if int64(len(gpuGRBody)) >= maxResponseSize {
				klog.Warningf("GPU Graphic response body too large (>= 10MB), truncated")
				mapUsages["gpuGraphicUtilization"] = "N/A"
			} else {
				gpuGRValue, err := ExactPrometheusValue(gpuGRBody)
				if err != nil {
					klog.Warningf("failed to extract GPU Graphic value: %v", err)
					gpuGRValue = "N/A"
				}
				mapUsages["gpuGraphicUtilization"] = gpuGRValue
			}
		}
		// DCGM_FI_DEV_FB_USED / DCGM_FI_DEV_FB_TOTAL: GPU Memory utilization
		gpuMemQuery := fmt.Sprintf(`sum(DCGM_FI_DEV_FB_USED{exported_container!="", exported_pod="%s", exported_namespace="%s"}) by (exported_namespace, exported_pod)/sum(DCGM_FI_DEV_FB_TOTAL{exported_container!="", exported_pod="%s", exported_namespace="%s"}) by (exported_namespace, exported_pod)`, alertPod, alertNamespace, alertPod, alertNamespace)
		// Use net/url.Values for automatic URL encoding
		gpu_mem_params := url.Values{}
		gpu_mem_params.Add("query", gpuMemQuery)
		// Build the full URL for the GPU Memory utilization query
		gpuMemUrl := fmt.Sprintf("%s%s?%s", prometheusURL, queryPath, gpu_mem_params.Encode())
		klog.Infof("Sending request to Prometheus for GPU Memory metrics: namespace=%s, pod=%s", alertNamespace, alertPod)

		// Make HTTP GET request to Prometheus for GPU Memory utilization
		gpuMemResp, err := httpGetFunc(gpuMemUrl)

		if err != nil {
			klog.Warningf("failed to query gpu memory usage: %v", err)
			mapUsages["gpuMemoryUtilization"] = "N/A"

		} else {
			klog.Infof("Received response from Prometheus for GPU Memory metrics: namespace=%s, pod=%s", alertNamespace, alertPod)
			defer func() { _ = gpuMemResp.Body.Close() }()
			gpuMemBody, err := io.ReadAll(io.LimitReader(gpuMemResp.Body, maxResponseSize))
			if err != nil {
				klog.Warningf("failed to read GPU Memory response body: %v", err)
				mapUsages["gpuMemoryUtilization"] = "N/A"
			} else if int64(len(gpuMemBody)) >= maxResponseSize {
				klog.Warningf("GPU Memory response body too large (>= 10MB), truncated")
				mapUsages["gpuMemoryUtilization"] = "N/A"
			} else {
				gpuMemValue, err := ExactPrometheusValue(gpuMemBody)
				if err != nil {
					klog.Warningf("failed to extract GPU Memory value: %v", err)
					gpuMemValue = "N/A"
				}
				mapUsages["gpuMemoryUtilization"] = gpuMemValue
			}
		}
	}
	// Marshal the updated Status.AppStatus.Usages to JSON
	podState.Status.AppStatus.Usages, err = jsonMarshalFunc(mapUsages)
	if err != nil {
		klog.Errorf("failed to marshal Usages %s/%s: %v", podState.Metadata.Namespace, podState.Metadata.Name, err)
		return nil, fmt.Errorf("failed to marshal Usages: %w", err)
	}

	// Directly create or update the DCIAppStateCR
	err = PodStateCrateOrPatch(ctx, k8sClient, podState, sch)
	if err != nil {
		klog.Errorf("failed to create or patch podStatus %s/%s: %v", podState.Metadata.Namespace, podState.Metadata.Name, err)
		return nil, fmt.Errorf("failed to create or patch podStatus: %w", err)
	}

	klog.Infof("Alertmanager signal activity: Successfully updated the DCIAppStateCR: namespace=%s, pod=%s", podState.Metadata.Namespace, podState.Metadata.Name)
	// Return the updated podState
	return &podState, nil
}

// Directly create or update the DCIAppStateCR
func PodStateCrateOrPatch(ctx context.Context, k8sClient client.Client, podState shared.PodUpdateStatus, sch *runtime.Scheme) error {
	var err error
	var rawAppStatusBytes []byte

	klog.Info("PodStateCR Crate Or Patch started")

	for range 1 {
		rawAppStatusBytes, err = jsonMarshalFunc(podState.Status.AppStatus)
		if err != nil {
			klog.Errorf("failed to marshal AppStatus. %s/%s: %v", podState.Metadata.Namespace, podState.Metadata.Name, err)
			break
		}

		pstCRData := &examplecomv1.DCIAppState{}

		// Fetch DCIAppState resources
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: podState.Metadata.Namespace,
			Name:      podState.Metadata.Name}, pstCRData)
		if err != nil {
			if !errors.IsNotFound(err) {
				klog.Errorf("unable to fetch DCIAppStateCR. %s/%s: %v", podState.Metadata.Namespace, podState.Metadata.Name, err)
				break
			}

			pstPodCRData := &corev1.Pod{}
			// Fetch Pod resources
			err = k8sClient.Get(ctx, client.ObjectKey{
				Namespace: podState.Metadata.Namespace,
				Name:      podState.Metadata.Name}, pstPodCRData)
			if err != nil {
				klog.Errorf("unable to fetch DCIAppStateCR. %s/%s: %v", podState.Metadata.Namespace, podState.Metadata.Name, err)
				break
			}
			pstCRData.Namespace = podState.Metadata.Namespace
			pstCRData.Name = podState.Metadata.Name
			pstCRData.Spec.Type = podState.Spec.Type
			pstCRData.Status.AppStatus.Raw = rawAppStatusBytes

			err = ctrlSetControllerReferenceFunc(
				pstPodCRData,
				pstCRData,
				sch)
			if err != nil {
				klog.Errorf("failed to SetControllerReference. %s/%s: %v", podState.Metadata.Namespace, podState.Metadata.Name, err)
			}

			err = k8sClient.Create(ctx, pstCRData)
			if err != nil {
				klog.Errorf("failed to create DCIAppStateCR. %s/%s: %v", podState.Metadata.Namespace, podState.Metadata.Name, err)
				break
			}
			klog.Infof("create DCIAppStateCR successfully. %s/%s", podState.Metadata.Namespace, podState.Metadata.Name)
		} else {
			pstBaseCRData := client.MergeFrom(pstCRData.DeepCopy())
			pstCRData.Status.AppStatus.Raw = rawAppStatusBytes

			err = k8sClient.Patch(ctx, pstCRData, pstBaseCRData)
			if err != nil {
				klog.Errorf("failed to patch DCIAppStateCR. %s/%s: %v", podState.Metadata.Namespace, podState.Metadata.Name, err)
				break
			}
			klog.Infof("patch DCIAppStateCR successfully. %s/%s", podState.Metadata.Namespace, podState.Metadata.Name)
		}
	}
	return err
}

// Function to extract metrics values from Prometheus response data
func ExactPrometheusValue(body []byte) (string, error) {
	var resp PrometheusResponse
	// Unmarshal the JSON response body into the 'resp' struct.
	// If unmarshaling fails, return an error.
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("failed to unmarshal: %w", err)
	}

	// Check if the response status indicates success.
	// If not, return an error with the reported status.
	if resp.Status != "success" {
		return "", fmt.Errorf("query failed, status: %s", resp.Status)
	}

	if len(resp.Data.Result) == 0 || len(resp.Data.Result[0].Value) < 2 {
		return "", fmt.Errorf("result empty or value missing")
	}

	strVal, ok := resp.Data.Result[0].Value[1].(string)
	if !ok {
		return "", fmt.Errorf("value is not string")
	}

	return strVal, nil
}

// CheckPodGPURequest checks if a GPU is allocated to a Pod by querying Kubernetes metrics
// Returns true if GPU, MIG resources are found in resource requests or limits
func CheckPodGPURequest(promURL, namespace, podName string) (bool, error) {
	// Try to check via Kubernetes API first (ResourceClaim)
	hasGPU, err := checkPodResourceClaimFunc(namespace, podName)
	if err != nil {
		return false, fmt.Errorf("failed to check ResourceClaim: %w", err)
	}

	if hasGPU {
		return true, nil
	}

	// If no ResourceClaim found, check traditional resource requests as fallback
	return checkPodGPUFromPrometheusFunc(promURL, namespace, podName)

}

// checkPodResourceClaim checks if a Pod has GPU allocated via ResourceClaim (DRA)
// Supports v1beta1, v1beta2, and v1 (GA) APIs based on RESOURCE_API_VERSION environment variable
func checkPodResourceClaim(namespace, podName string) (bool, error) {
	// Create in-cluster config
	config, err := inClusterConfigFunc()
	if err != nil {
		return false, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	// Create clientset
	clientset, err := newForConfigFunc(config)
	if err != nil {
		return false, fmt.Errorf("failed to create clientset: %w", err)
	}

	ctx := context.Background()

	// Get Pod information
	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get pod: %w", err)
	}

	// Check if Pod has ResourceClaims
	if len(pod.Spec.ResourceClaims) == 0 {
		klog.Infof("Pod %s/%s has no ResourceClaims", namespace, podName)
		return false, nil
	}

	// Get API version from environment variable (default: v1)
	apiVersion := os.Getenv("RESOURCE_API_VERSION")
	if apiVersion == "" {
		klog.Warning("RESOURCE_API_VERSION environment variable is not set. Defaulting to v1.")
		apiVersion = "v1"
	}
	if apiVersion != "v1beta1" && apiVersion != "v1beta2" && apiVersion != "v1" {
		klog.Warningf("Unsupported RESOURCE_API_VERSION: %s", apiVersion)
		return false, fmt.Errorf("unsupported RESOURCE_API_VERSION: %s", apiVersion)
	}
	klog.Infof("Using ResourceClaim API version: %s", apiVersion)

	// Iterate through all ResourceClaims referenced by the Pod
	for _, claimRef := range pod.Spec.ResourceClaims {
		claimName := claimRef.Name

		// Resolve the actual ResourceClaim name from pod status
		var actualClaimName string
		if pod.Status.ResourceClaimStatuses != nil {
			for _, status := range pod.Status.ResourceClaimStatuses {
				if status.Name == claimName {
					if status.ResourceClaimName != nil {
						actualClaimName = *status.ResourceClaimName
						klog.Infof("Resolved ResourceClaim name from status: %s -> %s", claimName, actualClaimName)
					}
					break
				}
			}
		}

		// Skip if ResourceClaim name cannot be resolved
		if actualClaimName == "" {
			klog.Infof("ResourceClaim %s has no resolved name in status", claimName)
			continue
		}

		// Get ResourceClaim using the appropriate API version
		var hasGPU bool

		switch apiVersion {
		case "v1beta2":
			claim, err := clientset.ResourceV1beta2().ResourceClaims(namespace).Get(ctx, actualClaimName, metav1.GetOptions{})
			if err != nil {
				klog.Warningf("Failed to get ResourceClaim %s/%s (v1beta2): %v", namespace, actualClaimName, err)
				continue
			}
			hasGPU = checkResourceClaimAllocationV1beta2(claim)

		case "v1beta1":
			claim, err := clientset.ResourceV1beta1().ResourceClaims(namespace).Get(ctx, actualClaimName, metav1.GetOptions{})
			if err != nil {
				klog.Warningf("Failed to get ResourceClaim %s/%s (v1beta1): %v", namespace, actualClaimName, err)
				continue
			}
			hasGPU = checkResourceClaimAllocationV1beta1(claim)

		default: // v1
			claim, err := clientset.ResourceV1().ResourceClaims(namespace).Get(ctx, actualClaimName, metav1.GetOptions{})
			if err != nil {
				klog.Warningf("Failed to get ResourceClaim %s/%s (v1): %v", namespace, actualClaimName, err)
				continue
			}
			hasGPU = checkResourceClaimAllocationV1(claim)
		}

		klog.Infof("Checked ResourceClaim: %s/%s (API: %s)", namespace, actualClaimName, apiVersion)

		if hasGPU {
			return true, nil
		}
	}

	return false, nil
}

// checkResourceClaimAllocationV1beta1 checks v1beta1 ResourceClaim for gpu.nvidia.com driver
func checkResourceClaimAllocationV1beta1(claim *resourcev1beta1.ResourceClaim) bool {
	if claim.Status.Allocation != nil {
		klog.Infof("ResourceClaim %s is allocated", claim.Name)

		if claim.Status.Allocation.Devices.Results != nil {
			for _, result := range claim.Status.Allocation.Devices.Results {
				klog.Infof("ResourceClaim %s has driver: %s", claim.Name, result.Driver)
				if result.Driver == "gpu.nvidia.com" {
					klog.Infof("GPU ResourceClaim found with gpu.nvidia.com driver: %s", claim.Name)
					return true
				}
			}
		}
	}
	return false
}

// checkResourceClaimAllocationV1beta2 checks v1beta2 ResourceClaim for gpu.nvidia.com driver
func checkResourceClaimAllocationV1beta2(claim *resourcev1beta2.ResourceClaim) bool {
	if claim.Status.Allocation != nil {
		klog.Infof("ResourceClaim %s is allocated", claim.Name)

		if claim.Status.Allocation.Devices.Results != nil {
			for _, result := range claim.Status.Allocation.Devices.Results {
				klog.Infof("ResourceClaim %s has driver: %s", claim.Name, result.Driver)
				if result.Driver == "gpu.nvidia.com" {
					klog.Infof("GPU ResourceClaim found with gpu.nvidia.com driver: %s", claim.Name)
					return true
				}
			}
		}
	}
	return false
}

// checkResourceClaimAllocationV1 checks v1 (GA) ResourceClaim for gpu.nvidia.com driver
func checkResourceClaimAllocationV1(claim *resourcev1.ResourceClaim) bool {
	if claim.Status.Allocation != nil {
		klog.Infof("ResourceClaim %s is allocated", claim.Name)

		if claim.Status.Allocation.Devices.Results != nil {
			for _, result := range claim.Status.Allocation.Devices.Results {
				klog.Infof("ResourceClaim %s has driver: %s", claim.Name, result.Driver)
				if result.Driver == "gpu.nvidia.com" {
					klog.Infof("GPU ResourceClaim found with gpu.nvidia.com driver: %s", claim.Name)
					return true
				}
			}
		}
	}
	return false
}

func checkPodGPUFromPrometheus(promURL, namespace, podName string) (bool, error) {
	queries := []string{
		fmt.Sprintf(`kube_pod_container_resource_requests{namespace="%s", pod="%s"}`, namespace, podName),
		fmt.Sprintf(`kube_pod_container_resource_limits{namespace="%s", pod="%s"}`, namespace, podName),
	}

	queryPath := "/api/v1/query"

	for _, query := range queries {

		// Use net/url.Values for automatic URL encoding
		params := url.Values{}
		params.Add("query", query)
		// Build the URL
		url := fmt.Sprintf("%s%s?%s", promURL, queryPath, params.Encode())

		// Make HTTP GET request to Prometheus
		resp, err := httpGetFunc(url)
		if err != nil {
			return false, fmt.Errorf("failed to connect to prometheus for pod %s/%s [query: %s]: %w", namespace, podName, query, err)
		}
		defer func() { _ = resp.Body.Close() }()
		// Read the response body
		body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
		if err != nil {
			return false, fmt.Errorf("failed to read prometheus response for pod %s/%s [query: %s]: %w", namespace, podName, query, err)
		} else if int64(len(body)) >= maxResponseSize {
			return false, fmt.Errorf("prometheus response body too large (>= 10MB), truncated for pod %s/%s [query: %s]", namespace, podName, query)
		}
		// Unmarshal the response JSON into PrometheusResponse struct
		var pr PrometheusResponse
		klog.Infof("prometheus response: %s", body)
		if err := json.Unmarshal(body, &pr); err != nil {
			return false, fmt.Errorf("prometheus returned error status '%s' for pod %s/%s [query: %s]", pr.Status, namespace, podName, query)
		}
		// Check if the response status is success
		if pr.Status != "success" {
			return false, fmt.Errorf("prometheus query failed: %s", pr.Status)
		}
		// Iterate through the results to check for GPU, MIG, or RDMA resources
		for _, result := range pr.Data.Result {
			resource := result.Metric["resource"]
			if strings.Contains(resource, "gpu") || strings.Contains(resource, "mig") {
				return true, nil
			}
		}
	}

	return false, nil
}
