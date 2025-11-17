package types

type CycleCommand struct {
	GreenStart    int32 `json:"green_start"`
	GreenDuration int32 `json:"green_duration"`
	CycleTotal    int32 `json:"cycle_total"`
}

type ListenerCommand struct {
	CycleCommand CycleCommand `json:"command"`
	DeviceId     string       `json:"deviceId"`
}
