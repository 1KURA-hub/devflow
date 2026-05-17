package mq

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeEvents          = "devflow.events"
	QueueFeedDistribute     = "devflow.feed.distribute"
	QueueNotificationCreate = "devflow.notification.create"
)

type Broker struct {
	conn *amqp.Connection
}

func Open(ctx context.Context, url string) (*Broker, error) {
	if url == "" {
		return nil, nil
	}

	conn, err := amqp.DialConfig(url, amqp.Config{
		Properties: amqp.Table{
			"connection_name": "devflow-server",
		},
	})
	if err != nil {
		return nil, err
	}

	broker := &Broker{conn: conn}
	if err := broker.declareTopology(ctx); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return broker, nil
}

func (b *Broker) Close() error {
	if b == nil || b.conn == nil {
		return nil
	}
	return b.conn.Close()
}

func (b *Broker) PublishJSON(ctx context.Context, routingKey string, payload any) error {
	if b == nil || b.conn == nil {
		return nil
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	channel, err := b.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	return channel.PublishWithContext(
		ctx,
		ExchangeEvents,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

func (b *Broker) Consume(queueName, consumerName string) (<-chan amqp.Delivery, error) {
	channel, err := b.conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := channel.Qos(1, 0, false); err != nil {
		_ = channel.Close()
		return nil, err
	}
	return channel.Consume(
		queueName,
		consumerName,
		false,
		false,
		false,
		false,
		nil,
	)
}

func (b *Broker) declareTopology(ctx context.Context) error {
	channel, err := b.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	if err := channel.ExchangeDeclare(
		ExchangeEvents,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	if err := declareQueueAndBind(ctx, channel, QueueFeedDistribute, "post.published"); err != nil {
		return err
	}
	if err := declareQueueAndBind(ctx, channel, QueueNotificationCreate, "interaction.*"); err != nil {
		return err
	}
	return channel.QueueBind(
		QueueNotificationCreate,
		"user.followed",
		ExchangeEvents,
		false,
		nil,
	)
}

func declareQueueAndBind(_ context.Context, channel *amqp.Channel, queueName, routingKey string) error {
	if _, err := channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}
	return channel.QueueBind(
		queueName,
		routingKey,
		ExchangeEvents,
		false,
		nil,
	)
}
