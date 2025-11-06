package types

import "time"

type OrchestrationStatus string

const (
	StatusPending OrchestrationStatus = "PENDING"
	StatusSent    OrchestrationStatus = "SENT"
	StatusAcked   OrchestrationStatus = "ACKED"
	StatusApplied OrchestrationStatus = "APPLIED"
	StatusRetry   OrchestrationStatus = "RETRY"
	StatusFailed  OrchestrationStatus = "FAILED"
)

type DesiredPayload struct {
	ActivationTime int `json:"activation_time"`
}

type OrchestrationMessage struct {
	OrchestrationID string              `json:"orchestrationId"`
	PackID          int64               `json:"packId"`
	DeviceID        string              `json:"deviceId"`
	DesiredPayload  DesiredPayload      `json:"desiredPayload"`
	Status          OrchestrationStatus `json:"status"`
	SentAt          time.Time           `json:"sentAt"`
	Attempts        int                 `json:"attempts"`
	LastError       *string             `json:"lastError,omitempty"`
	Version         int                 `json:"version"`
}
