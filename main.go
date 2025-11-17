package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JamzMarks/tcc_orquestrador/services"
	"github.com/JamzMarks/tcc_orquestrador/services/core"
	iothub "github.com/JamzMarks/tcc_orquestrador/services/iot"
	"github.com/JamzMarks/tcc_orquestrador/services/rabbit"
	"github.com/JamzMarks/tcc_orquestrador/types"
)

type Operation struct {
	MinorPackCicle int32
	MajorPackCicle int32
}

type OnServices struct {
	Iot   *iothub.Service
	Mq    *rabbit.RabbitService
	Graph *services.OrquestradorService
}

func main() {
	svcs := Connection()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	packChan := make(chan types.Pack, 100)

	numWorkers := 10
	for i := range numWorkers {
		go worker(ctx, i, packChan, svcs)
	}

	for {
		packs, err := svcs.Graph.FetchPacks(ctx)
		if err != nil {
			log.Printf("Erro ao buscar packs: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("📦 Total de packs encontrados: %d", len(packs))

		for _, pack := range packs {
			packChan <- pack
		}

		time.Sleep(60 * time.Second)
	}
}

func Connection() *OnServices {
	cfg := services.LoadConfig()

	type result[T any] struct {
		svc T
		err error
	}

	graphCh := make(chan result[*services.OrquestradorService])
	iotCh := make(chan result[*iothub.Service])
	mqCh := make(chan result[*rabbit.RabbitService])

	//  GRAPH
	go func() {
		for {
			graph := services.NewGraphService(cfg.GraphDBURI, cfg.GraphDBUser, cfg.GraphDBPass)
			if graph != nil {
				if err := graph.TestConnection(); err == nil {
					graphCh <- result[*services.OrquestradorService]{svc: graph, err: nil}
					return
				}
			}
			log.Println("[Retry] Neo4j offline, tentando novamente em 5s...")
			time.Sleep(5 * time.Second)
		}
	}()

	//  IOT
	go func() {
		for {
			iot, err := iothub.NewSAS(cfg.IotConnectionString, 86400)
			if err == nil {
				if err := iot.TestConnection(); err == nil {
					iotCh <- result[*iothub.Service]{svc: iot, err: nil}
					return
				}
			}
			log.Println("[Retry] IoT Hub offline, tentando novamente em 5s...")
			time.Sleep(5 * time.Second)
		}
	}()

	//  RABBITMQ
	go func() {
		for {
			mq, err := rabbit.NewRabbitService(cfg.RabbitURL)
			if err == nil {
				mqCh <- result[*rabbit.RabbitService]{svc: mq, err: nil}

				queues := []string{"orquestrador", "logs", "system_events"}
				mq.ConsumeMultiple(queues, func(queue string, body []byte) error {
					log.Printf("Fila [%s] recebeu: %s", queue, string(body))
					return nil
				})

				return
			}
			log.Println("[Retry] RabbitMQ offline, tentando novamente em 5s...")
			time.Sleep(5 * time.Second)
		}
	}()

	//  AGUARDA TODOS
	var graph *services.OrquestradorService
	var iot *iothub.Service
	var mq *rabbit.RabbitService

	for i := 0; i < 3; i++ {
		select {
		case r := <-graphCh:
			graph = r.svc
			log.Println("✔ Neo4j conectado.")

		case r := <-iotCh:
			iot = r.svc
			log.Println("✔ IoT Hub conectado.")

		case r := <-mqCh:
			mq = r.svc
			log.Println("✔ RabbitMQ conectado.")
		}
	}

	return &OnServices{
		Iot:   iot,
		Mq:    mq,
		Graph: graph,
	}
}

func worker(ctx context.Context, id int, packChan <-chan types.Pack, svcs *OnServices) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("[Worker %d] Finalizando...", id)
			return

		case pack, ok := <-packChan:
			if !ok {
				log.Printf("[Worker %d] Canal fechado. Encerrando.", id)
				return
			}

			processPack(ctx, pack, svcs.Iot, svcs.Mq)
		}
	}
}

func processPack(ctx context.Context, p types.Pack, iot *iothub.Service, mq *rabbit.RabbitService) {

	blocos := core.DistribuirCycleComBlocos(p)

	var (
		startFloat float64
		totalFloat float64
		totalInt   int32 = int32(p.Cycle)
	)

	for i, bloco := range blocos {

		durFloat := bloco.Duration

		startInt := int32(startFloat + 0.5)
		durInt := int32(durFloat + 0.5)

		if i == len(blocos)-1 {
			diff := totalInt - (startInt + durInt)
			if diff != 0 {
				durInt += diff
			}
		}

		for _, deviceID := range bloco.DeviceIDs {

			payload := types.CycleCommand{
				GreenStart:    startInt,
				GreenDuration: durInt,
				CycleTotal:    totalInt,
			}

			msg, _ := json.Marshal(payload)
			topic := fmt.Sprintf("devices/%s/messages/devicebound/", deviceID)

			fmt.Printf(
				"[IoT] → %s | Start=%d Dur=%d Total=%d\n",
				deviceID, payload.GreenStart, payload.GreenDuration, payload.CycleTotal,
			)

			if err := iot.Publish(topic, msg); err != nil {
				fmt.Printf("[ERRO IoT] %s: %v\n", deviceID, err)
			}

			listener := types.ListenerCommand{
				CycleCommand: payload,
				DeviceId:     deviceID,
			}

			msgRabbit, _ := json.Marshal(listener)

			if err := mq.Publish("device_events", msgRabbit); err != nil {
				fmt.Printf("[ERRO Rabbit] %s: %v\n", deviceID, err)
			}

			fmt.Printf("[IoT+Rabbit] → %s | Start=%d Dur=%d Total=%d\n",
				deviceID, payload.GreenStart, payload.GreenDuration, payload.CycleTotal)
		}

		startFloat += durFloat
		totalFloat += durFloat
	}
}
