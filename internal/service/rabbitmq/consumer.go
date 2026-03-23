package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"order-management-service/internal/config"
	"order-management-service/internal/service"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Consumer interface {
	Consume(ctx context.Context) error
}

type rabbitMQConsumer struct {
	conn        *amqp.Connection
	channel     *amqp.Channel
	cfg         *config.Config
	logger      *zap.Logger
	activitySvc service.OrderActivityService
}

func NewRabbitMQConsumer(
	cfg *config.Config,
	logger *zap.Logger,
	activitySvc service.OrderActivityService,
) (Consumer, error) {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &rabbitMQConsumer{
		conn:        conn,
		channel:     ch,
		cfg:         cfg,
		logger:      logger,
		activitySvc: activitySvc,
	}, nil
}

func (c *rabbitMQConsumer) Consume(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.cfg.RabbitQueue,
		"",    // consumer
		false, // auto-ack
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
			c.logger.Info("Received message from queue", zap.String("body", string(d.Body)))

			err := c.processMessage(ctx, d.Body)
			if err != nil {
				c.logger.Error("Failed to process message, initiating retry", zap.Error(err))
				c.handleRetry(d)
			} else {
				d.Ack(false)
			}
		}
	}()

	c.logger.Info("RabbitMQ Consumer started successfully")
	return nil
}

func (c *rabbitMQConsumer) processMessage(ctx context.Context, body []byte) error {
	var payload struct {
		OrderUUID uuid.UUID `json:"order_uuid"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	c.logger.Info("Delegating task to OrderActivityService", zap.String("order_uuid", payload.OrderUUID.String()))

	return c.activitySvc.HandleOrderCreatedActivity(ctx, payload.OrderUUID)
}

func (c *rabbitMQConsumer) handleRetry(d amqp.Delivery) {
	retryCount := 0
	if d.Headers == nil {
		d.Headers = make(amqp.Table)
	}

	if val, ok := d.Headers["x-retry-count"].(int32); ok {
		retryCount = int(val)
	}

	if retryCount < c.cfg.MaxRetryAttempts {
		retryCount++
		c.logger.Warn("Retrying background task", zap.Int("attempt", retryCount))

		delay := time.Duration(c.cfg.RetryBaseDelay*retryCount) * time.Millisecond
		time.Sleep(delay)

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
			d.Nack(false, true)
		} else {
			d.Ack(false)
		}
	} else {
		c.logger.Error("Max retry attempts reached. Moving to DLX.", zap.String("body", string(d.Body)))
		d.Nack(false, false)
	}
}
