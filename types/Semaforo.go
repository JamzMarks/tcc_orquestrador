package types

type Semaforo struct {
	ID       string
	WayId    []string
	Priority float64
}

type DistribuicaoSemaforo struct {
	SemaforoID string
	Tempo      float64
}
