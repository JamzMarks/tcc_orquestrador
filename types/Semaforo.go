package types

type Semaforo struct {
	ID            string
	DeviceId      string
	WayId         string
	Priority      float64
	GreenStart    *float64
	GreenDuration *float64
}

type DistribuicaoSemaforo struct {
	SemaforoID string
	Tempo      float64
}
