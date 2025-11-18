package types

import "time"

type CycleCommand struct {
	GreenStart    int32 `json:"green_start"`
	GreenDuration int32 `json:"green_duration"`
	CycleTotal    int32 `json:"cycle_total"`
}

type ListenerCommand struct {
	CycleCommand CycleCommand `json:"command"`
	DeviceId     string       `json:"deviceId"`
}

type TelemetryError struct {
	DeviceID  string    `json:"deviceId"`
	Module    string    `json:"module"`    // ex: "orchestrator"
	Event     string    `json:"event"`     // ex: "signal.update"
	ErrorMsg  string    `json:"error"`     // ex: "IoT publish failed"
	Payload   any       `json:"payload"`   // mensagem original enviada
	Timestamp time.Time `json:"timestamp"` // ISO-8601
}
