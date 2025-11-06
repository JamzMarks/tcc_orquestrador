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

// Cria instância do serviço Neo4j
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

// Fecha a conexão com Neo4j
func (s *OrquestradorService) Close() {
	if s.driver != nil {
		s.driver.Close(context.Background())
	}
}

// Testa a conexão com uma query simples
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

// FetchWays retorna todas as OSMWays do Neo4j
func (s *OrquestradorService) FetchWays(ctx context.Context) ([]*types.OSMWay, error) {
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

		id, _ := record.Get("id")
		name, _ := record.Get("name")
		priority, _ := record.Get("priority")
		updatedAt, _ := record.Get("updated_at")

		way := &types.OSMWay{
			ID:        fmt.Sprintf("%v", id),
			Name:      fmt.Sprintf("%v", name),
			Priority:  toFloat(priority),
			UpdatedAt: toTime(updatedAt),
		}

		ways = append(ways, way)
	}

	if err = result.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar Ways: %w", err)
	}

	return ways, nil
}

// Helper para converter para float64
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

// Helper para converter para time.Time
func toTime(v any) time.Time {
	if v == nil {
		return time.Time{}
	}
	switch val := v.(type) {
	case time.Time:
		return val
	case string:
		t, _ := time.Parse(time.RFC3339, val)
		return t
	default:
		return time.Time{}
	}
}
