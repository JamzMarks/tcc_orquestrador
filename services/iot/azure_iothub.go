package iothub

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Service struct {
	client   mqtt.Client
	clientID string
	host     string
}

// NewSASFromConnectionString cria um Service conectado ao IoT Hub usando a Connection String completa
func NewSAS(connStr string, duration time.Duration) (*Service, error) {
	parts := strings.Split(connStr, ";")
	var host, keyName, key string
	for _, part := range parts {
		if strings.HasPrefix(part, "HostName=") {
			host = strings.TrimPrefix(part, "HostName=")
		} else if strings.HasPrefix(part, "SharedAccessKeyName=") {
			keyName = strings.TrimPrefix(part, "SharedAccessKeyName=")
		} else if strings.HasPrefix(part, "SharedAccessKey=") {
			key = strings.TrimPrefix(part, "SharedAccessKey=")
		}
	}

	if host == "" || keyName == "" || key == "" {
		return nil, fmt.Errorf("connection string inválida")
	}
	fmt.Printf(host, keyName, key)
	// Gera SAS Token válido pelo tempo duration
	expiry := time.Now().Add(duration)
	sasToken := buildSASToken(host, key, expiry)

	clientID := "orquestrador"
	username := fmt.Sprintf("%s/?api-version=2021-04-12", host)

	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tls://%s:8883", host)).
		SetClientID(clientID).
		SetUsername(username).
		SetPassword(sasToken).
		SetKeepAlive(60 * time.Second).
		SetPingTimeout(10 * time.Second)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("falha ao conectar ao IoT Hub: %w", token.Error())
	}

	fmt.Println("[IoT] Conectado ao Hub:", host)
	return &Service{
		client:   client,
		clientID: clientID,
		host:     host,
	}, nil
}

// buildSASToken gera o SAS Token usado no MQTT
func buildSASToken(resourceURI, key string, expiry time.Time) string {
	encodedURI := url.QueryEscape(resourceURI)
	exp := strconv.FormatInt(expiry.Unix(), 10)
	stringToSign := encodedURI + "\n" + exp

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(stringToSign))
	signature := url.QueryEscape(base64.StdEncoding.EncodeToString(h.Sum(nil)))

	return fmt.Sprintf("SharedAccessSignature sr=%s&sig=%s&se=%s", encodedURI, signature, exp)
}

// Publish envia uma mensagem ao IoT Hub
func (s *Service) Publish(topic string, payload []byte) error {
	token := s.client.Publish(topic, 1, false, payload)
	token.Wait()
	if err := token.Error(); err != nil {
		return fmt.Errorf("erro ao publicar: %w", err)
	}
	fmt.Printf("[IoT] → Enviado para %s: %s\n", topic, payload)
	return nil
}

// Subscribe inscreve o dispositivo em um tópico
func (s *Service) Subscribe(topic string, handler func(msg string)) error {
	callback := func(_ mqtt.Client, msg mqtt.Message) {
		fmt.Printf("[IoT] ← Recebido de %s: %s\n", msg.Topic(), msg.Payload())
		handler(string(msg.Payload()))
	}
	token := s.client.Subscribe(topic, 1, callback)
	token.Wait()
	if err := token.Error(); err != nil {
		return fmt.Errorf("erro ao fazer subscribe: %w", err)
	}
	fmt.Printf("[IoT] Inscrito no tópico: %s\n", topic)
	return nil
}

// Close encerra a sessão MQTT
func (s *Service) Close() {
	fmt.Println("[IoT] Conexão encerrada.")
	s.client.Disconnect(250)
}

func (s *Service) TestConnection() error {
	topic := fmt.Sprintf("devices/%s/messages/events/", s.clientID)
	payload := []byte("teste de conexão")

	if err := s.Publish(topic, payload); err != nil {
		return fmt.Errorf("falha no publish de teste: %w", err)
	}

	fmt.Println("[IoT] Teste de conexão OK")
	return nil
}
