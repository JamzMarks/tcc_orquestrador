package types

type SubPack struct {
	ID            string
	Name          *string
	Semaforos     []Semaforo
	GreenStart    float64
	GreenDuration float64
}

type Pack struct {
	ID        string
	Cycle     float64
	Name      string
	Semaforos []Semaforo
	SubPacks  []SubPack
}
