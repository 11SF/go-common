package kafka

import (
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"hash"

	"github.com/Shopify/sarama"
	"github.com/xdg-go/scram"
)

const tag = "kafka"

// SASLMechanism supported values.
const (
	SASLPlain       = "PLAIN"
	SASLSCRAMSHA256 = "SCRAM-SHA-256"
	SASLSCRAMSHA512 = "SCRAM-SHA-512"
)

// Config holds connection settings shared by producer and consumer.
type Config struct {
	Brokers       []string `env:"KAFKA_BROKERS"        envSeparator:","`
	SASLEnable    bool     `env:"KAFKA_SASL_ENABLE"`
	SASLMechanism string   `env:"KAFKA_SASL_MECHANISM"` // PLAIN | SCRAM-SHA-256 | SCRAM-SHA-512
	SASLUser      string   `env:"KAFKA_SASL_USER"`
	SASLPassword  string   `env:"KAFKA_SASL_PASSWORD"`
	TLSEnable     bool     `env:"KAFKA_TLS_ENABLE"`
}

func (c Config) newSaramaConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_8_0_0

	if c.TLSEnable {
		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	if c.SASLEnable {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.Handshake = true
		cfg.Net.SASL.User = c.SASLUser
		cfg.Net.SASL.Password = c.SASLPassword

		switch c.SASLMechanism {
		case SASLSCRAMSHA256:
			cfg.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
			cfg.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &scramClient{HashGeneratorFcn: func() hash.Hash { return sha256.New() }}
			}
		case SASLSCRAMSHA512:
			cfg.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
			cfg.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &scramClient{HashGeneratorFcn: func() hash.Hash { return sha512.New() }}
			}
		default: // PLAIN
			cfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		}
	}

	return cfg
}

// scramClient implements sarama.SCRAMClient using xdg/scram.
type scramClient struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

func (x *scramClient) Begin(userName, password, authzID string) error {
	var err error
	x.Client, err = x.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	x.ClientConversation = x.Client.NewConversation()
	return nil
}

func (x *scramClient) Step(challenge string) (string, error) {
	return x.ClientConversation.Step(challenge)
}

func (x *scramClient) Done() bool {
	return x.ClientConversation.Done()
}
