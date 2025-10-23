package iothub

import "fmt"

// Publish envia uma mensagem ao IoT Hub
func (s *Service) Publish(topic string, payload []byte) error {
	token := s.client.Publish(topic, 1, false, payload)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}
	fmt.Printf("[IoT] Enviado para %s: %s\n", topic, string(payload))
	return nil
}
