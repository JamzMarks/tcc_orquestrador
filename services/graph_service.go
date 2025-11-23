package services

import (
	"context"
	"fmt"
	"log"

	// "time"
	"github.com/JamzMarks/tcc_orquestrador/types"
	"github.com/JamzMarks/tcc_orquestrador/utils"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type OrquestradorService struct {
	uri      string
	username string
	password string
	driver   neo4j.DriverWithContext
}

func NewGraphService(uri, username, password string) *OrquestradorService {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		log.Fatalf("Erro ao criar driver Neo4j: %v", err)
	}
	return &OrquestradorService{
		uri:      uri,
		username: username,
		password: password,
		driver:   driver,
	}
}

func (s *OrquestradorService) Close() {
	if s.driver != nil {
		s.driver.Close(context.Background())
	}
}

func (s *OrquestradorService) TestConnection() error {
	ctx := context.Background()
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	result, err := session.Run(ctx, "RETURN 'Conexão OK!' AS message", nil)
	if err != nil {
		return fmt.Errorf("erro ao executar query: %v", err)
	}

	if result.Next(ctx) {
		fmt.Println(result.Record().Values[0])
	} else if err = result.Err(); err != nil {
		return fmt.Errorf("erro ao ler resultado: %v", err)

	}
	return nil

}

func (s *OrquestradorService) FetchPacks(ctx context.Context) ([]types.Pack, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	query := `
	MATCH (p:Pack)

		OPTIONAL MATCH (p)-[:HAS_SEMAFORO]->(s:Semaforo)
		OPTIONAL MATCH (s)-[:CONTROLS_TRAFFIC_ON]->(w:OSMWay)
		OPTIONAL MATCH (p)-[:HAS_SUBPACK]->(sp:SubPack)
		OPTIONAL MATCH (sp)-[:HAS_SEMAFORO]->(ss:Semaforo)
		OPTIONAL MATCH (ss)-[:CONTROLS_TRAFFIC_ON]->(w2:OSMWay)

		WITH p,

			// Semáforos diretos do Pack
			collect(DISTINCT {
				id: elementId(s),
				deviceId: s.deviceId,
				wayId: elementId(w),
				priority: w.priority,
				green_start: s.green_start,
				green_duration: s.green_duration
			}) AS packSemaforos,

			collect(DISTINCT sp) AS subpackNodes,

			collect(DISTINCT {
				spId: elementId(sp),
				ssId: elementId(ss),
				deviceId: ss.deviceId,
				wayId: elementId(w2),
				priority: w2.priority,
				green_start: ss.green_start,
				green_duration: ss.green_duration
			}) AS subPairs

		WITH p, packSemaforos,
			[spNode IN subpackNodes |
				{
				id: elementId(spNode),
				name: spNode.name,
				green_start: spNode.green_start,
				green_duration: spNode.green_duration,
				semaforos: [
					pair IN subPairs
					WHERE pair.spId = elementId(spNode)
					| {
						id: pair.ssId,
						deviceId: pair.deviceId,
						wayId: pair.wayId,
						priority: pair.priority

					}
				]
				}
			] AS subpacks

		RETURN {
			id: elementId(p),
			name: p.name,
			configs: { cicle: p.cicle },
			semaforos: packSemaforos,
			subPacks: subpacks
		} AS pack;

	`
	log.Println("Antes de session.Run🔥")
	result, err := session.Run(ctx, query, nil)
	if err != nil {
		log.Println("Erro na query🔥")
		return nil, err
	}
	log.Println("DEPOIS de session.Run🔥")
	var packs []types.Pack

	for result.Next(ctx) {
		record := result.Record()

		packRaw, _ := record.Get("pack")
		packMap := packRaw.(map[string]interface{})

		pack := types.Pack{
			ID:    packMap["id"].(string),
			Name:  packMap["name"].(string),
			Cycle: utils.ToFloat(packMap["configs"].(map[string]interface{})["cicle"]),
		}

		if sems, ok := packMap["semaforos"].([]interface{}); ok {

			for _, s := range sems {
				sm := s.(map[string]interface{})

				sem := types.Semaforo{
					ID:            sm["id"].(string),
					DeviceId:      sm["deviceId"].(string),
					WayId:         sm["wayId"].(string),
					Priority:      utils.ToFloat(sm["priority"]),
					GreenStart:    utils.ToFloatPtr(sm["green_start"]),
					GreenDuration: utils.ToFloatPtr(sm["green_duration"]),
				}

				pack.Semaforos = append(pack.Semaforos, sem)
			}
		}

		if subs, ok := packMap["subPacks"].([]interface{}); ok {
			for _, sp := range subs {
				spm := sp.(map[string]interface{})

				subpack := types.SubPack{
					ID:            spm["id"].(string),
					Name:          utils.PtrString(spm["name"]),
					GreenStart:    utils.ToFloat(spm["green_start"]),
					GreenDuration: utils.ToFloat(spm["green_duration"]),
				}

				// Semáforos do SubPack
				if ssList, ok := spm["semaforos"].([]interface{}); ok {
					for _, ss := range ssList {
						ssm := ss.(map[string]interface{})

						sem := types.Semaforo{
							ID:       ssm["id"].(string),
							Priority: utils.ToFloat(ssm["priority"]),
							WayId:    ssm["wayId"].(string),
							DeviceId: ssm["deviceId"].(string),
						}

						subpack.Semaforos = append(subpack.Semaforos, sem)
					}
				}

				pack.SubPacks = append(pack.SubPacks, subpack)
			}
		}
		fmt.Println("Finalizando")
		packs = append(packs, pack)
	}

	if err = result.Err(); err != nil {
		return nil, err
	}

	return packs, nil
}
