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

func (s *OrquestradorService) TestConnection() {
	ctx := context.Background()
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	result, err := session.Run(ctx, "RETURN 'Conexão OK!' AS message", nil)
	if err != nil {
		log.Fatalf("Erro ao executar query: %v", err)
	}

	if result.Next(ctx) {
		fmt.Println(result.Record().Values[0])
	} else if err = result.Err(); err != nil {
		log.Fatalf("Erro ao ler resultado: %v", err)
	}
}

func (s *OrquestradorService) FetchPacks(ctx context.Context) ([]types.Pack, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	query := `
	MATCH (p:Pack)
	OPTIONAL MATCH (p)-[:HAS_SEMAFORO]->(s:Semaforo)-[:CONTROLS_TRAFFIC_ON]->(w:OSMWay)
	OPTIONAL MATCH (p)-[:HAS_SUBPACK]->(sp:SubPack)-[:HAS_SEMAFORO]->(ss:Semaforo)-[:CONTROLS_TRAFFIC_ON]->(sw:OSMWay)
	RETURN p.id AS packId,
	       p.cycle AS cycle,
	       collect(DISTINCT {id: s.id, wayIds: collect(DISTINCT elementId(w)), priority: sum(w.priority)}) AS semaforos,
	       collect(DISTINCT {id: sp.id, semaforos: collect(DISTINCT {id: ss.id, wayIds: collect(DISTINCT elementId(sw)), priority: sum(sw.priority)})}) AS subpacks
	`

	result, err := session.Run(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	var packs []types.Pack

	for result.Next(ctx) {
		record := result.Record()

		packID, _ := record.Get("packId")
		cycle, _ := record.Get("cycle")

		pack := types.Pack{
			ID:    packID.(string),
			Cycle: utils.ToFloat(cycle),
		}

		// Semáforos diretos
		if semaforosRaw, ok := record.Get("semaforos"); ok && semaforosRaw != nil {
			for _, s := range semaforosRaw.([]interface{}) {
				sMap := s.(map[string]interface{})
				wayIds := []string{}
				for _, w := range sMap["wayIds"].([]interface{}) {
					wayIds = append(wayIds, w.(string))
				}
				pack.Semaforos = append(pack.Semaforos, types.Semaforo{
					ID:       sMap["id"].(string),
					WayId:    wayIds,
					Priority: utils.ToFloat(sMap["priority"]),
				})
			}
		}

		// Subpacks
		if subpacksRaw, ok := record.Get("subpacks"); ok && subpacksRaw != nil {
			for _, sp := range subpacksRaw.([]interface{}) {
				spMap := sp.(map[string]interface{})
				subpack := types.SubPack{
					ID: spMap["id"].(string),
				}
				for _, ss := range spMap["semaforos"].([]interface{}) {
					ssMap := ss.(map[string]interface{})
					wayIds := []string{}
					for _, w := range ssMap["wayIds"].([]interface{}) {
						wayIds = append(wayIds, w.(string))
					}
					subpack.Semaforos = append(subpack.Semaforos, types.Semaforo{
						ID:       ssMap["id"].(string),
						WayId:    wayIds,
						Priority: utils.ToFloat(ssMap["priority"]),
					})
				}
				pack.SubPacks = append(pack.SubPacks, subpack)
			}
		}

		packs = append(packs, pack)
	}

	if err = result.Err(); err != nil {
		return nil, err
	}

	return packs, nil
}
