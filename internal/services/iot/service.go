package iothub

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Service struct {
	client   mqtt.Client
	deviceID string
}

// New cria uma nova instância conectada ao Azure IoT Hub
func New(host, deviceID, username, password string) (*Service, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tls://%s:8883", host)).
		SetClientID(deviceID).
		SetUsername(username).
		SetPassword(password).
		SetKeepAlive(60 * time.Second).
		SetPingTimeout(10 * time.Second)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &Service{client: client, deviceID: deviceID}, nil
}

// Close encerra a conexão com o IoT Hub
func (s *Service) Close() {
	s.client.Disconnect(250)
}
