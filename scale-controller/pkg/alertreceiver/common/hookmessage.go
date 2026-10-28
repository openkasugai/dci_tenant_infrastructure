package common

// AlertManagerHookMessage defines the structure of the message received from Prometheus AlertManager webhook.
// See https://grafana.com/docs/grafana/latest/alerting/configure-notifications/manage-contact-points/integrations/webhook-notifier/
type AlertManagerHookMessage struct {
	Version           string              `json:"version"`
	GroupKey          string              `json:"groupKey"`
	TruncatedAlerts   int                 `json:"truncatedAlerts"`
	Status            string              `json:"status"`
	Receiver          string              `json:"receiver"`
	GroupLabels       map[string]string   `json:"groupLabels"`
	CommonLabels      map[string]string   `json:"commonLabels"`
	CommonAnnotations map[string]string   `json:"commonAnnotations"`
	ExternalURL       string              `json:"externalURL"`
	Alerts            []AlertManagerAlert `json:"alerts"`
}

// AlertManagerAlert defines the structure of the alert received from Prometheus AlertManager webhook.
// See https://grafana.com/docs/grafana/latest/alerting/configure-notifications/manage-contact-points/integrations/webhook-notifier/
type AlertManagerAlert struct {
	Status       string            `json:"status,omitempty"` // firing or resolved
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt,omitempty"`
	EndsAt       string            `json:"endsAt,omitempty"`
	GeneratorURL string            `json:"generatorURL,omitempty"`
	Fingerprint  string            `json:"fingerprint,omitempty"`
}

const (
	// AlertManagerAlertStatusFiring represents the "firing" status of an alert.
	AlertManagerAlertStatusFiring = "firing"

	// AlertManagerAlertStatusResolved represents the "resolved" status of an alert.
	AlertManagerAlertStatusResolved = "resolved"

	// AlertManagerAlertLabelAlertNameKey is the key for the alert name label in an alert.
	AlertManagerAlertLabelAlertNameKey = "alertname"
)
