package kafka

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/11SF/go-common/logger"
	"github.com/Shopify/sarama"
)

const (
	defaultMaxRetry     = 3
	defaultRetryBackoff = 500 * time.Millisecond
)

// ConsumerConfig holds settings for a Kafka consumer group.
type ConsumerConfig struct {
	Config

	// Group is the Kafka consumer group ID.
	Group string `env:"KAFKA_GROUP"`
	// Topics to subscribe to.
	Topics []string `env:"KAFKA_TOPICS" envSeparator:","`
	// OffsetOldest starts consuming from the oldest offset when no committed offset exists.
	// Default false = OffsetNewest.
	OffsetOldest bool `env:"KAFKA_OFFSET_OLDEST"`
	// MaxRetry is the number of in-memory retries before sending to DeadLetterTopic. Default: 3.
	MaxRetry int
	// RetryBackoff between in-memory retries. Default: 500ms.
	RetryBackoff time.Duration
	// DeadLetterTopic is the topic to which unrecoverable messages are forwarded.
	// If empty, failed messages are only logged and their offset is committed.
	DeadLetterTopic string `env:"KAFKA_DEAD_LETTER_TOPIC"`
}

func (c *ConsumerConfig) withDefaults() {
	if c.MaxRetry == 0 {
		c.MaxRetry = defaultMaxRetry
	}
	if c.RetryBackoff == 0 {
		c.RetryBackoff = defaultRetryBackoff
	}
}

// Message represents a consumed Kafka message.
type Message struct {
	Topic     string
	Partition int32
	Offset    int64
	Key       []byte
	Value     []byte
	Headers   map[string][]byte
	Timestamp time.Time
}

// HandlerFunc is called for each consumed message.
// A non-nil error causes the message to be retried up to MaxRetry times.
type HandlerFunc func(ctx context.Context, msg Message) error

// Consumer wraps a sarama ConsumerGroup with retry and dead-letter support.
type Consumer struct {
	group  sarama.ConsumerGroup
	cfg    ConsumerConfig
	dlq    *Producer // nil when DeadLetterTopic is empty
}

// NewConsumer creates a Kafka consumer group and verifies connectivity.
func NewConsumer(ctx context.Context, cfg ConsumerConfig) (*Consumer, error) {
	cfg.withDefaults()

	sc := cfg.Config.newSaramaConfig()
	sc.Consumer.Return.Errors = true
	if cfg.OffsetOldest {
		sc.Consumer.Offsets.Initial = sarama.OffsetOldest
	} else {
		sc.Consumer.Offsets.Initial = sarama.OffsetNewest
	}
	// Enable auto-offset-commit is off; we commit manually per message.
	sc.Consumer.Offsets.AutoCommit.Enable = false

	logger.Info(ctx, "connecting kafka consumer",
		slog.Any("brokers", cfg.Brokers),
		slog.String("group", cfg.Group),
		slog.Any("topics", cfg.Topics),
		logger.LogAttrTag(tag),
	)

	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.Group, sc)
	if err != nil {
		logger.Error(ctx, "failed to create kafka consumer group",
			logger.LogAttrError(err),
			logger.LogAttrTag(tag),
		)
		return nil, err
	}

	c := &Consumer{group: group, cfg: cfg}

	if cfg.DeadLetterTopic != "" {
		dlq, err := NewProducer(ctx, ProducerConfig{Config: cfg.Config})
		if err != nil {
			_ = group.Close()
			return nil, err
		}
		c.dlq = dlq
	}

	logger.Info(ctx, "kafka consumer connected",
		slog.String("group", cfg.Group),
		logger.LogAttrTag(tag),
	)

	return c, nil
}

// Subscribe starts the consume loop and blocks until ctx is cancelled.
// Rebalances are handled automatically. Each message is retried up to MaxRetry
// times; persistent failures are forwarded to DeadLetterTopic (if configured).
func (c *Consumer) Subscribe(ctx context.Context, handler HandlerFunc) error {
	h := &consumerGroupHandler{cfg: c.cfg, handler: handler, dlq: c.dlq}

	logger.Info(ctx, "kafka consumer starting",
		slog.String("group", c.cfg.Group),
		slog.Any("topics", c.cfg.Topics),
		logger.LogAttrTag(tag),
	)

	for {
		if err := c.group.Consume(ctx, c.cfg.Topics, h); err != nil {
			if errors.Is(err, sarama.ErrClosedConsumerGroup) || errors.Is(err, context.Canceled) {
				logger.Info(ctx, "kafka consumer stopped", logger.LogAttrTag(tag))
				return nil
			}
			logger.Error(ctx, "kafka consumer error",
				logger.LogAttrError(err),
				logger.LogAttrTag(tag),
			)
		}

		if ctx.Err() != nil {
			logger.Info(ctx, "kafka consumer stopped", logger.LogAttrTag(tag))
			return nil
		}
	}
}

// Close shuts down the consumer group and the DLQ producer (if any).
func (c *Consumer) Close() error {
	if c.dlq != nil {
		_ = c.dlq.Close()
	}
	return c.group.Close()
}

// ---- sarama ConsumerGroupHandler implementation ----

type consumerGroupHandler struct {
	cfg     ConsumerConfig
	handler HandlerFunc
	dlq     *Producer
}

func (h *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	logger.Info(session.Context(), "kafka consumer group rebalanced",
		slog.String("group", h.cfg.Group),
		slog.Any("claims", session.Claims()),
		logger.LogAttrTag(tag),
	)
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case m, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			h.process(session, m)
		case <-session.Context().Done():
			return nil
		}
	}
}

func (h *consumerGroupHandler) process(session sarama.ConsumerGroupSession, m *sarama.ConsumerMessage) {
	ctx := session.Context()
	msg := toMessage(m)

	var err error
	for attempt := 1; attempt <= h.cfg.MaxRetry; attempt++ {
		err = h.handler(ctx, msg)
		if err == nil {
			break
		}
		logger.Warn(ctx, "handler failed, will retry",
			logger.LogAttrError(err),
			slog.String("topic", msg.Topic),
			slog.Int64("offset", msg.Offset),
			slog.Int("attempt", attempt),
			slog.Int("max_retry", h.cfg.MaxRetry),
			logger.LogAttrTag(tag),
		)
		if attempt < h.cfg.MaxRetry {
			time.Sleep(h.cfg.RetryBackoff)
		}
	}

	if err != nil {
		logger.Error(ctx, "handler failed after max retry, dropping message",
			logger.LogAttrError(err),
			slog.String("topic", msg.Topic),
			slog.Int64("offset", msg.Offset),
			logger.LogAttrTag(tag),
		)
		if h.dlq != nil {
			h.sendToDeadLetter(ctx, msg, err)
		}
	}

	// Always commit offset so a bad message does not block the partition.
	session.MarkMessage(m, "")
	session.Commit()
}

func (h *consumerGroupHandler) sendToDeadLetter(ctx context.Context, msg Message, cause error) {
	// Carry original metadata as headers.
	dlqMsg := &sarama.ProducerMessage{
		Topic: h.cfg.DeadLetterTopic,
		Value: sarama.ByteEncoder(msg.Value),
		Headers: []sarama.RecordHeader{
			{Key: []byte("_source_topic"), Value: []byte(msg.Topic)},
			{Key: []byte("_source_offset"), Value: []byte(slog.Int64Value(msg.Offset).String())},
			{Key: []byte("_error"), Value: []byte(cause.Error())},
		},
	}
	if len(msg.Key) > 0 {
		dlqMsg.Key = sarama.ByteEncoder(msg.Key)
	}

	_, _, err := h.dlq.client.SendMessage(dlqMsg)
	if err != nil {
		logger.Error(ctx, "failed to send message to dead-letter topic",
			logger.LogAttrError(err),
			slog.String("dead_letter_topic", h.cfg.DeadLetterTopic),
			slog.String("source_topic", msg.Topic),
			logger.LogAttrTag(tag),
		)
	} else {
		logger.Info(ctx, "message forwarded to dead-letter topic",
			slog.String("dead_letter_topic", h.cfg.DeadLetterTopic),
			slog.String("source_topic", msg.Topic),
			slog.Int64("source_offset", msg.Offset),
			logger.LogAttrTag(tag),
		)
	}
}

func toMessage(m *sarama.ConsumerMessage) Message {
	headers := make(map[string][]byte, len(m.Headers))
	for _, h := range m.Headers {
		headers[string(h.Key)] = h.Value
	}
	return Message{
		Topic:     m.Topic,
		Partition: m.Partition,
		Offset:    m.Offset,
		Key:       m.Key,
		Value:     m.Value,
		Headers:   headers,
		Timestamp: m.Timestamp,
	}
}
