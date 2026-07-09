package iot

import (
	"os"
	"strconv"
	"time"

	iotauth "github.com/functionfly/functionfly/internal/auth/iot"
)

type IoTConfig struct {
	MQTTPort       int
	MQTTTLS        bool
	MQTTCertFile   string
	MQTTKeyFile    string
	UseEmbedded    bool
	ExternalBroker string

	COAPPort  int
	COAPSTLS  bool

	NATSUrl string

	Auth *iotauth.DeviceAuthConfig

	EventRetention time.Duration
	CommandTimeout time.Duration
}

func DefaultConfig() *IoTConfig {
	return &IoTConfig{
		MQTTPort:       1883,
		MQTTTLS:        false,
		UseEmbedded:    true,
		ExternalBroker: "",
		COAPPort:       5683,
		COAPSTLS:       false,
		NATSUrl:        "nats://localhost:4222",
		Auth:           iotauth.DefaultDeviceAuthConfig(),
		EventRetention: 24 * time.Hour,
		CommandTimeout: 30 * time.Second,
	}
}

func ConfigFromEnv() *IoTConfig {
	cfg := DefaultConfig()

	if port := os.Getenv("IOT_MQTT_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.MQTTPort = p
		}
	}
	if v := os.Getenv("IOT_MQTT_TLS"); v == "true" || v == "1" {
		cfg.MQTTTLS = true
	}
	cfg.MQTTCertFile = os.Getenv("IOT_MQTT_CERT_FILE")
	cfg.MQTTKeyFile = os.Getenv("IOT_MQTT_KEY_FILE")
	if v := os.Getenv("IOT_USE_EMBEDDED_BROKER"); v == "false" || v == "0" {
		cfg.UseEmbedded = false
	}
	cfg.ExternalBroker = os.Getenv("IOT_EXTERNAL_BROKER")

	if port := os.Getenv("IOT_COAP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.COAPPort = p
		}
	}
	if v := os.Getenv("IOT_COAP_DTLS"); v == "true" || v == "1" {
		cfg.COAPSTLS = true
	}

	if url := os.Getenv("NATS_URL"); url != "" {
		cfg.NATSUrl = url
	}

	return cfg
}
