package types

import "time"

type Semaforo struct {
	ID         string    `json:"id"`
	DeviceID   string    `json:"deviceId"`
	CreatedAt  time.Time `json:"createdAt"`
	IP         string    `json:"ip"`
	MacAddress string    `json:"macAddress"`
	IsActive   bool      `json:"isActive"`
	PackID     *int64    `json:"packId,omitempty"`
	SubPackID  *int64    `json:"subPackId,omitempty"`
	DeviceKey  string    `json:"deviceKey"`
}

type SubPack struct {
	ID        string     `json:"id"`
	PackID    int64      `json:"packId"`
	Semaforos []Semaforo `json:"semaforos"`
}

type Pack struct {
	ID        string     `json:"id"`
	Cycle     int64      `json:"cycle"`
	Name      string     `json:"name"`
	SubPacks  []SubPack  `json:"subPacks"`
	Semaforos []Semaforo `json:"semaforos"`
}
