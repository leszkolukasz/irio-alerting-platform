package dto

type PubSubPayload struct {
	IncidentID   string                 `json:"incident_id,omitempty"`
	ServiceID    string                 `json:"monitored_service_id,omitempty"`
	OnCallerData map[string]interface{} `json:"oncaller_data,omitempty"`
	Timestamp    string                 `json:"timestamp,omitempty"`
}
