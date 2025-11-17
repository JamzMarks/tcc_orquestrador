package main

import (
	"github.com/JamzMarks/tcc_orquestrador/services"
	iothub "github.com/JamzMarks/tcc_orquestrador/services/iot"
	"github.com/JamzMarks/tcc_orquestrador/services/rabbit"
	"github.com/JamzMarks/tcc_orquestrador/types"
)

func main() {
	cfg := services.LoadConfig()

	// Neo4j
	svc := services.NewGraphService(cfg.GraphDBURI, cfg.GraphDBUser, cfg.GraphDBPass)
	svc.TestConnection()

	// Azure IoT Hub
	iothub, err := iothub.NewSAS(cfg.IotConnectionString, 86400)
	if err != nil {
		panic(err)
	}
	defer iothub.Close()
	iothub.TestConnection()
	// RabbitMq
	rbmq, err := rabbit.NewRabbitService(cfg.RabbitURL, cfg.QueueName)
	if err != nil {
		panic(err)
	}
	rbmq.Publish([]byte(`{"acao":"verde"}`))
}

func DistribuirCycle(p types.Pack) []types.DistribuicaoSemaforo {
	var resultado []types.DistribuicaoSemaforo

	entidades := len(p.Semaforos) + len(p.SubPacks)
	if entidades == 0 {
		return resultado
	}

	// Tempo para cada entidade (semafaro OU subpack)
	tempoPorEntidade := p.Cycle / float64(entidades)

	// 1. Semáforos diretos do pack
	for _, s := range p.Semaforos {
		resultado = append(resultado, types.DistribuicaoSemaforo{
			SemaforoID: s.ID,
			Tempo:      tempoPorEntidade,
		})
	}

	// 2. Subpacks — todos recebem o mesmo tempo
	for _, sp := range p.SubPacks {
		for _, s := range sp.Semaforos {
			resultado = append(resultado, types.DistribuicaoSemaforo{
				SemaforoID: s.ID,
				Tempo:      tempoPorEntidade,
			})
		}
	}

	return resultado
}
