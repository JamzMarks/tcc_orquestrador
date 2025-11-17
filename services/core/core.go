package core

import "github.com/JamzMarks/tcc_orquestrador/types"

type Bloco struct {
	DeviceIDs []string
	Duration  float64
}

func DistribuirCycleComBlocos(p types.Pack) []Bloco {
	var blocos []Bloco
	var totalPriority float64

	type entidade struct {
		DeviceIDs []string
		Priority  float64
	}

	var entidades []entidade

	// 1. Semáforos
	for _, s := range p.Semaforos {
		entidades = append(entidades, entidade{
			DeviceIDs: []string{s.ID},
			Priority:  s.Priority,
		})
		totalPriority += s.Priority
	}

	// 2. Subpacks
	for _, sp := range p.SubPacks {
		if len(sp.Semaforos) == 0 {
			continue
		}

		var ids []string
		for _, s := range sp.Semaforos {
			ids = append(ids, s.ID)
		}

		entidades = append(entidades, entidade{
			DeviceIDs: ids,
			Priority:  sp.Semaforos[0].Priority,
		})

		totalPriority += sp.Semaforos[0].Priority
	}

	if totalPriority == 0 {
		return blocos
	}

	// 3. Calcula blocos proporcionais
	for _, e := range entidades {
		duration := (e.Priority / totalPriority) * p.Cycle

		blocos = append(blocos, Bloco{
			DeviceIDs: e.DeviceIDs,
			Duration:  duration,
		})
	}

	return blocos
}
