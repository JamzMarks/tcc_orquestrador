package main

import (
	"log"
	"time"

	"github.com/JamzMarks/tcc_orquestrador/internal/services"
	"github.com/JamzMarks/tcc_orquestrador/internal/services/rabbit"
)

// RabbitConnector mantém um ponteiro compartilhado do serviço RabbitMQ
type RabbitConnector struct {
	Service   *rabbit.RabbitService
	IsRunning bool
}

// Start inicia a tentativa de conexão em background
func (rc *RabbitConnector) Start(cfg *services.Config) {
	rc.IsRunning = true
	go func() {
		for rc.IsRunning {
			if rc.Service == nil {
				log.Printf("🔌 Tentando conectar ao RabbitMQ em %s ...", cfg.RabbitURL)
				svc, err := rabbit.NewRabbitService(cfg.RabbitURL, cfg.QueueName)
				if err != nil {
					log.Printf("⚠️ Falha ao conectar ao RabbitMQ: %v", err)
					log.Println("⏳ Tentando novamente em 15 segundos...")
					time.Sleep(15 * time.Second)
					continue
				}
				rc.Service = svc
				log.Printf("✅ Conectado ao RabbitMQ em %s (fila: %s)", cfg.RabbitURL, cfg.QueueName)
			}

			// Verifica se a conexão ainda está viva
			if rc.Service != nil && rc.Service.ConnClosed() {
				log.Printf("⚠️ Conexão com RabbitMQ perdida! Tentando reconectar...")
				rc.Service.Close()
				rc.Service = nil
			}

			time.Sleep(10 * time.Second)
		}
	}()
}

// Stop encerra a goroutine e fecha o serviço
func (rc *RabbitConnector) Stop() {
	rc.IsRunning = false
	if rc.Service != nil {
		rc.Service.Close()
	}
}

// Helper para saber se pode publicar
func (rc *RabbitConnector) IsReady() bool {
	return rc.Service != nil
}
