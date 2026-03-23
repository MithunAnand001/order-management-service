package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"order-management-service/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Consumer interface {
	Consume(ctx context.Context) error
}

type rabbitMQConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     *config.Config
	logger  *zap.Logger
}

func NewRabbitMQConsumer(cfg *config.Config, logger *zap.Logger) (Consumer, error) {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &rabbitMQConsumer{
		conn:    conn,
		channel: ch,
		cfg:     cfg,
		logger:  logger,
	}, nil
}

func (c *rabbitMQConsumer) Consume(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.cfg.RabbitQueue,
		"",    // consumer
		false, // auto-ack (set to false for manual ack/retry)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			c.logger.Info("Received a message", zap.String("body", string(d.Body)))
			
			err := c.processMessage(d.Body)
			if err != nil {
				c.handleRetry(d)
			} else {
				d.Ack(false)
			}
		}
	}()

	c.logger.Info("Waiting for messages...")
	return nil
}

func (c *rabbitMQConsumer) processMessage(body []byte) error {
	// Business logic for post-order creation activities
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	
	c.logger.Info("Processing activity for order", zap.Any("payload", payload))
	return nil
}

func (c *rabbitMQConsumer) handleRetry(d amqp.Delivery) {
	// Extract retry count from headers
	retryCount := 0
	if d.Headers == nil {
		d.Headers = make(amqp.Table)
	}
	
	if val, ok := d.Headers["x-retry-count"].(int32); ok {
		retryCount = int(val)
	}

	if retryCount < c.cfg.MaxRetryAttempts {
		retryCount++
		c.logger.Warn("Retrying message", zap.Int("attempt", retryCount))
		
		// Wait before retrying (Exponential backoff simplified)
		delay := time.Duration(c.cfg.RetryBaseDelay * retryCount) * time.Millisecond
		time.Sleep(delay)

		// Publish back to exchange with incremented retry count
		headers := d.Headers
		headers["x-retry-count"] = int32(retryCount)

		err := c.channel.PublishWithContext(context.Background(),
			c.cfg.RabbitExchange,
			c.cfg.RabbitBinding,
			false,
			false,
			amqp.Publishing{
				ContentType: d.ContentType,
				Body:        d.Body,
				Headers:     headers,
			},
		)
		
		if err != nil {
			c.logger.Error("Failed to republish for retry", zap.Error(err))
			d.Nack(false, true) // Requeue to original queue as last resort
		} else {
			d.Ack(false)
		}
	} else {
		c.logger.Error("Max retry attempts reached. Moving to DLX.", zap.String("body", string(d.Body)))
		d.Nack(false, false) // Nack without requeue triggers DLX
	}
}
