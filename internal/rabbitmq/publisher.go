package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"order-management-service/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageBroker interface {
	PublishOrderCreated(ctx context.Context, orderUUID string) error
	Close()
}

type rabbitMQBroker struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     *config.Config
}

func NewRabbitMQBroker(cfg *config.Config) (MessageBroker, error) {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Setup DLX
	err = ch.ExchangeDeclare(
		cfg.RabbitDLX,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare DLX: %w", err)
	}

	// Setup Main Exchange
	err = ch.ExchangeDeclare(
		cfg.RabbitExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Setup Main Queue with DLX
	args := amqp.Table{
		"x-dead-letter-exchange": cfg.RabbitDLX,
	}
	_, err = ch.QueueDeclare(
		cfg.RabbitQueue,
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.QueueBind(
		cfg.RabbitQueue,
		cfg.RabbitBinding,
		cfg.RabbitExchange,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &rabbitMQBroker{
		conn:    conn,
		channel: ch,
		cfg:     cfg,
	}, nil
}

func (r *rabbitMQBroker) PublishOrderCreated(ctx context.Context, orderUUID string) error {
	body, err := json.Marshal(map[string]string{"order_uuid": orderUUID})
	if err != nil {
		return err
	}

	err = r.channel.PublishWithContext(ctx,
		r.cfg.RabbitExchange,
		r.cfg.RabbitBinding,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("Failed to publish message: %v", err)
		return err
	}

	log.Printf("Published OrderCreated event for UUID: %s", orderUUID)
	return nil
}

func (r *rabbitMQBroker) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}
