package rabbit

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitService struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

// NewRabbitService cria conexão e garante que a fila existe
func NewRabbitService(amqpURL, queueName string) (*RabbitService, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitService{
		conn:    conn,
		channel: ch,
		queue:   queueName,
	}, nil
}

// Publish envia uma mensagem à fila sem esperar resposta (fire-and-forget)
func (r *RabbitService) Publish(ctx context.Context, body []byte) error {
	if r.channel == nil {
		return amqp.ErrClosed
	}

	// Contexto com timeout (opcional)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := r.channel.PublishWithContext(ctx,
		"",      // exchange vazio → envia direto à fila
		r.queue, // routing key (nome da fila)
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // garante persistência se a fila for durável
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		log.Printf("❌ Erro ao publicar mensagem no RabbitMQ: %v", err)
		return err
	}

	log.Printf("📤 Mensagem publicada na fila '%s' (%d bytes)", r.queue, len(body))
	return nil
}

func (r *RabbitService) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

// ConnClosed verifica se a conexão foi encerrada
func (r *RabbitService) ConnClosed() bool {
	return r.conn == nil || r.conn.IsClosed()
}
