package rabbit

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitService struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

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

// Publish envia uma mensagem para a fila
func (r *RabbitService) Publish(body []byte) error {
	return r.channel.Publish(
		"",      // exchange (vazio = padrão)
		r.queue, // routing key = nome da fila
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

// Consume inicia o consumo de mensagens em background
func (r *RabbitService) Consume(handler func(body []byte) error) error {
	msgs, err := r.channel.Consume(
		r.queue,
		"",    // consumer tag
		false, // autoAck -> melhor deixar false
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				log.Printf("Erro ao processar mensagem: %v", err)
				msg.Nack(false, true) // requeue
				continue
			}
			msg.Ack(false)
		}
	}()

	return nil
}

// Close fecha tudo
func (r *RabbitService) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}
