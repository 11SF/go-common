package redisclient

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/11SF/go-common/logger"
	"github.com/redis/go-redis/v9"
)

const (
	mqTag         = "redis stream"
	startFromNew  = "$"
	startFromOld  = "0"
	defaultBlock  = 5 * time.Second
	defaultBatch  = int64(10)
	defaultRetry  = 3
)

// StreamConfig holds configuration for a Redis Stream message queue.
type StreamConfig struct {
	Stream        string        // Stream key name
	Group         string        // Consumer group name
	Consumer      string        // Consumer name (unique per instance)
	BlockDuration time.Duration // How long to block waiting for new messages (default 5s)
	BatchSize     int64         // Max messages to fetch per read (default 10)
	MaxRetry      int           // Max delivery attempts before sending to dead-letter (default 3)
}

func (c *StreamConfig) withDefaults() {
	if c.BlockDuration == 0 {
		c.BlockDuration = defaultBlock
	}
	if c.BatchSize == 0 {
		c.BatchSize = defaultBatch
	}
	if c.MaxRetry == 0 {
		c.MaxRetry = defaultRetry
	}
}

// Message represents a single Redis Stream entry.
type Message struct {
	ID     string
	Values map[string]any
}

// HandlerFunc is the callback invoked for each consumed message.
// Returning a non-nil error causes the message to remain pending (not ACKed).
type HandlerFunc func(ctx context.Context, msg Message) error

// MessageQueue wraps a Redis client to provide producer/consumer helpers over Redis Streams.
type MessageQueue struct {
	client redis.UniversalClient
	config StreamConfig
}

// NewMessageQueue creates a MessageQueue.
// It also ensures the consumer group exists (creating the stream if needed).
func NewMessageQueue(ctx context.Context, client redis.UniversalClient, config StreamConfig) (*MessageQueue, error) {
	config.withDefaults()
	mq := &MessageQueue{client: client, config: config}
	if err := mq.ensureGroup(ctx); err != nil {
		return nil, err
	}
	return mq, nil
}

// Publish adds a message to the stream. Returns the assigned message ID.
func (mq *MessageQueue) Publish(ctx context.Context, values map[string]any) (string, error) {
	id, err := mq.client.XAdd(ctx, &redis.XAddArgs{
		Stream: mq.config.Stream,
		MaxLen: 0, // no trim; caller can configure externally
		Values: values,
	}).Result()
	if err != nil {
		logger.Error(ctx, "failed to publish message",
			logger.LogAttrError(err),
			logger.LogAttrTag(mqTag),
			slog.String("stream", mq.config.Stream),
		)
		return "", err
	}
	logger.Debug(ctx, "message published",
		slog.String("stream", mq.config.Stream),
		slog.String("id", id),
		logger.LogAttrTag(mqTag),
	)
	return id, nil
}

// Subscribe starts consuming messages in a blocking loop until ctx is cancelled.
// Each message is passed to handler; if handler returns nil the message is ACKed.
// Pending (unACKed) messages from a previous run are redelivered on startup.
func (mq *MessageQueue) Subscribe(ctx context.Context, handler HandlerFunc) error {
	logger.Info(ctx, "starting consumer",
		slog.String("stream", mq.config.Stream),
		slog.String("group", mq.config.Group),
		slog.String("consumer", mq.config.Consumer),
		logger.LogAttrTag(mqTag),
	)

	// First, redeliver any pending messages from a previous run.
	if err := mq.recoverPending(ctx, handler); err != nil {
		return err
	}

	// Then consume new messages.
	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, "consumer stopped", logger.LogAttrTag(mqTag))
			return ctx.Err()
		default:
		}

		streams, err := mq.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    mq.config.Group,
			Consumer: mq.config.Consumer,
			Streams:  []string{mq.config.Stream, startFromNew},
			Count:    mq.config.BatchSize,
			Block:    mq.config.BlockDuration,
		}).Result()

		if err != nil {
			if errors.Is(err, redis.Nil) || errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			if errors.Is(err, context.Canceled) {
				return nil
			}
			logger.Error(ctx, "error reading from stream",
				logger.LogAttrError(err),
				logger.LogAttrTag(mqTag),
			)
			continue
		}

		for _, stream := range streams {
			mq.handleMessages(ctx, stream.Messages, handler)
		}
	}
}

// Ack explicitly acknowledges a message by ID.
func (mq *MessageQueue) Ack(ctx context.Context, msgID string) error {
	return mq.client.XAck(ctx, mq.config.Stream, mq.config.Group, msgID).Err()
}

// recoverPending redelivers all pending (unACKed) messages assigned to this consumer.
func (mq *MessageQueue) recoverPending(ctx context.Context, handler HandlerFunc) error {
	cursor := startFromOld
	for {
		streams, err := mq.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    mq.config.Group,
			Consumer: mq.config.Consumer,
			Streams:  []string{mq.config.Stream, cursor},
			Count:    mq.config.BatchSize,
		}).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return nil
			}
			return err
		}

		var lastID string
		for _, stream := range streams {
			if len(stream.Messages) == 0 {
				return nil
			}
			mq.handleMessages(ctx, stream.Messages, handler)
			lastID = stream.Messages[len(stream.Messages)-1].ID
		}

		if lastID == "" {
			return nil
		}
		cursor = lastID
	}
}

// handleMessages processes a batch and ACKs successful ones.
// Messages that exceed MaxRetry are moved to the dead-letter stream.
func (mq *MessageQueue) handleMessages(ctx context.Context, messages []redis.XMessage, handler HandlerFunc) {
	for _, m := range messages {
		msg := Message{ID: m.ID, Values: m.Values}

		deliveryCount, _ := mq.deliveryCount(ctx, m.ID)
		if deliveryCount > int64(mq.config.MaxRetry) {
			logger.Warn(ctx, "message exceeded max retry, moving to dead-letter",
				slog.String("stream", mq.config.Stream),
				slog.String("id", m.ID),
				slog.Int64("delivery_count", deliveryCount),
				logger.LogAttrTag(mqTag),
			)
			mq.sendToDeadLetter(ctx, msg)
			_ = mq.Ack(ctx, m.ID)
			continue
		}

		if err := handler(ctx, msg); err != nil {
			logger.Error(ctx, "handler error, message will be retried",
				logger.LogAttrError(err),
				slog.String("stream", mq.config.Stream),
				slog.String("id", m.ID),
				logger.LogAttrTag(mqTag),
			)
			continue
		}

		if err := mq.Ack(ctx, m.ID); err != nil {
			logger.Error(ctx, "failed to ack message",
				logger.LogAttrError(err),
				slog.String("id", m.ID),
				logger.LogAttrTag(mqTag),
			)
		}
	}
}

// deliveryCount returns how many times a pending message has been delivered.
func (mq *MessageQueue) deliveryCount(ctx context.Context, msgID string) (int64, error) {
	pending, err := mq.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: mq.config.Stream,
		Group:  mq.config.Group,
		Start:  msgID,
		End:    msgID,
		Count:  1,
	}).Result()
	if err != nil || len(pending) == 0 {
		return 0, err
	}
	return pending[0].RetryCount, nil
}

// sendToDeadLetter publishes the message to a dead-letter stream (<stream>:dead-letter).
func (mq *MessageQueue) sendToDeadLetter(ctx context.Context, msg Message) {
	dlStream := mq.config.Stream + ":dead-letter"
	values := make(map[string]any, len(msg.Values)+2)
	for k, v := range msg.Values {
		values[k] = v
	}
	values["_original_id"] = msg.ID
	values["_source_stream"] = mq.config.Stream

	if err := mq.client.XAdd(ctx, &redis.XAddArgs{
		Stream: dlStream,
		Values: values,
	}).Err(); err != nil {
		logger.Error(ctx, "failed to send message to dead-letter stream",
			logger.LogAttrError(err),
			slog.String("dead_letter_stream", dlStream),
			slog.String("id", msg.ID),
			logger.LogAttrTag(mqTag),
		)
	}
}

// ensureGroup creates the consumer group (and stream) if they don't already exist.
func (mq *MessageQueue) ensureGroup(ctx context.Context) error {
	err := mq.client.XGroupCreateMkStream(ctx, mq.config.Stream, mq.config.Group, startFromOld).Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		logger.Error(ctx, "failed to create consumer group",
			logger.LogAttrError(err),
			slog.String("stream", mq.config.Stream),
			slog.String("group", mq.config.Group),
			logger.LogAttrTag(mqTag),
		)
		return err
	}
	return nil
}
