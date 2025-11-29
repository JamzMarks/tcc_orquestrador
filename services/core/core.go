package core

import (
	"sort"

	"github.com/JamzMarks/tcc_orquestrador/types"
)

type Bloco struct {
	DeviceIDs []string
	Duration  float64
}
type entidade struct {
	DeviceIDs  []string
	Priority   float64
	GreenStart *float64
}

func DistribuirCycleComBlocos(p types.Pack) []Bloco {
	var blocos []Bloco
	var totalPriority float64

	var entidades []entidade

	// 1. Semáforos
	for _, s := range p.Semaforos {
		entidades = append(entidades, entidade{
			DeviceIDs:  []string{s.DeviceId},
			Priority:   s.Priority,
			GreenStart: s.GreenStart,
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
			ids = append(ids, s.DeviceId)
		}

		entidades = append(entidades, entidade{
			DeviceIDs:  ids,
			Priority:   sp.Semaforos[0].Priority,
			GreenStart: sp.Semaforos[0].GreenStart,
		})

		totalPriority += sp.Semaforos[0].Priority
	}

	if totalPriority == 0 {
		return blocos
	}
	sort.Slice(entidades, func(i, j int) bool {
		gi := entidades[i].GreenStart
		gj := entidades[j].GreenStart

		// Casos com nil
		if gi == nil && gj == nil {
			return false // mantém ordem
		}
		if gi == nil {
			return false // nil vem depois
		}
		if gj == nil {
			return true // nil vem depois
		}

		// Comparação por valor
		return *gi < *gj
	})

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
