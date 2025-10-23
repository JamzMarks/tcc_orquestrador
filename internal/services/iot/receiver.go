package iothub

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func (s *Service) Subscribe(topic string, handler func(msg string)) error {
	callback := func(client mqtt.Client, message mqtt.Message) {
		fmt.Printf("[IoT] Recebido de %s: %s\n", message.Topic(), string(message.Payload()))
		handler(string(message.Payload()))
	}
	token := s.client.Subscribe(topic, 1, callback)
	token.Wait()
	return token.Error()
}
