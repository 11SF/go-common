package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/11SF/go-common/logger"
	"github.com/Shopify/sarama"
)

// ProducerConfig holds settings for the Kafka producer.
type ProducerConfig struct {
	Config

	// RequiredAcks controls durability. Default: WaitForAll (strongest guarantee).
	RequiredAcks sarama.RequiredAcks
	// Idempotent enables exactly-once delivery. Requires RequiredAcks = WaitForAll and MaxRetry > 0.
	Idempotent bool
	// Compression codec. Default: Snappy (fast + good ratio).
	Compression sarama.CompressionCodec
	// MaxRetry is the number of retries on transient send errors. Default: 3.
	MaxRetry int
	// RetryBackoff between retries. Default: 250ms.
	RetryBackoff time.Duration
}

func (c *ProducerConfig) withDefaults() {
	if c.RequiredAcks == 0 {
		c.RequiredAcks = sarama.WaitForAll
	}
	if c.Compression == sarama.CompressionNone {
		c.Compression = sarama.CompressionSnappy
	}
	if c.MaxRetry == 0 {
		c.MaxRetry = 3
	}
	if c.RetryBackoff == 0 {
		c.RetryBackoff = 250 * time.Millisecond
	}
}

// Producer is a synchronous Kafka producer.
type Producer struct {
	client sarama.SyncProducer
}

// NewProducer creates a new synchronous Kafka producer and verifies connectivity.
func NewProducer(ctx context.Context, cfg ProducerConfig) (*Producer, error) {
	cfg.withDefaults()

	sc := cfg.Config.newSaramaConfig()
	sc.Producer.RequiredAcks = cfg.RequiredAcks
	sc.Producer.Compression = cfg.Compression
	sc.Producer.Retry.Max = cfg.MaxRetry
	sc.Producer.Retry.Backoff = cfg.RetryBackoff
	sc.Producer.Return.Successes = true
	sc.Producer.Return.Errors = true

	if cfg.Idempotent {
		sc.Producer.Idempotent = true
		sc.Net.MaxOpenRequests = 1
	}

	logger.Info(ctx, "connecting kafka producer",
		slog.Any("brokers", cfg.Brokers),
		logger.LogAttrTag(tag),
	)

	client, err := sarama.NewSyncProducer(cfg.Brokers, sc)
	if err != nil {
		logger.Error(ctx, "failed to create kafka producer",
			logger.LogAttrError(err),
			logger.LogAttrTag(tag),
		)
		return nil, err
	}

	logger.Info(ctx, "kafka producer connected",
		slog.Any("brokers", cfg.Brokers),
		logger.LogAttrTag(tag),
	)

	return &Producer{client: client}, nil
}

// Send publishes a raw byte message to the given topic.
// Returns the assigned partition and offset.
func (p *Producer) Send(ctx context.Context, topic string, key, value []byte) (partition int32, offset int64, err error) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(value),
	}
	if len(key) > 0 {
		msg.Key = sarama.ByteEncoder(key)
	}

	partition, offset, err = p.client.SendMessage(msg)
	if err != nil {
		logger.Error(ctx, "failed to send kafka message",
			logger.LogAttrError(err),
			slog.String("topic", topic),
			logger.LogAttrTag(tag),
		)
		return 0, 0, err
	}

	logger.Debug(ctx, "kafka message sent",
		slog.String("topic", topic),
		slog.Int("partition", int(partition)),
		slog.Int64("offset", offset),
		logger.LogAttrTag(tag),
	)

	return partition, offset, nil
}

// SendJSON marshals v as JSON and publishes it to the given topic.
func (p *Producer) SendJSON(ctx context.Context, topic string, key string, v any) error {
	payload, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, _, err = p.Send(ctx, topic, []byte(key), payload)
	return err
}

// Close shuts down the producer and flushes any pending messages.
func (p *Producer) Close() error {
	return p.client.Close()
}
