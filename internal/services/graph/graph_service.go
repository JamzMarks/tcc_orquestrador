package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JamzMarks/tcc_orquestrador/types"
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

func (s *Service) FetchWays(ctx context.Context) ([]*types.OSMWay, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	query := `
		MATCH (w:OSMWay)
		RETURN elementId(w) AS id, 
		       w.name AS name,
		       w.priority AS priority,
		       w.updated_at AS updated_at
	`

	result, err := session.Run(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar Ways: %w", err)
	}

	var ways []*types.OSMWay

	for result.Next(ctx) {
		record := result.Record()

		way := &types.OSMWay{
			ID:        record.GetByIndex(0).(string),
			Name:      record.GetByIndex(1).(string),
			Priority:  toFloat(record.GetByIndex(2)),
			UpdatedAt: toTime(record.GetByIndex(6)),
		}
		ways = append(ways, way)
	}

	if err = result.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar Ways: %w", err)
	}

	return ways, nil
}

func toFloat(v any) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	default:
		return 0
	}
}

func toTime(v any) time.Time {
	t, _ := v.(time.Time)
	return t
}
