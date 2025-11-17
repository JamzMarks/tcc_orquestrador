package types

type SubPack struct {
	ID        string
	Semaforos []Semaforo
}

type Pack struct {
	ID        string
	Cycle     float64
	Semaforos []Semaforo
	SubPacks  []SubPack
}
