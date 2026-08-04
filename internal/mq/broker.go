package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeEvents          = "devflow.events"
	ExchangeDeadLetter      = "devflow.dead_letter"
	ExchangeRetryReturn     = "devflow.retry.return"
	QueueFeedDistribute     = "devflow.feed.distribute"
	QueueNotificationCreate = "devflow.notification.create"

	MaxDeliveryRetries = 3
	retryDelayMillis   = int32(1000)

	retryQueueSuffix       = ".retry"
	deadLetterQueueSuffix  = ".dlq"
	retryCountHeader       = "x-devflow-retry-count"
	lastErrorHeader        = "x-devflow-last-error"
	deadLetterReasonHeader = "x-devflow-dead-letter-reason"
)

var (
	ErrPublishNack = errors.New("rabbitmq rejected published message")
	ErrUnroutable  = errors.New("rabbitmq message was unroutable")
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
	return b.publish(ctx, ExchangeEvents, routingKey, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
		Timestamp:    time.Now().UTC(),
	})
}

// RetryOrDeadLetter republishes a transiently failed delivery to a short-lived
// retry queue. After MaxDeliveryRetries retries, it sends the delivery to the
// queue-specific dead-letter queue. The original delivery is acknowledged only
// after the replacement publish is confirmed by RabbitMQ.
func (b *Broker) RetryOrDeadLetter(ctx context.Context, queueName string, delivery amqp.Delivery, cause error) error {
	if DeliveryRetryCount(delivery.Headers) >= MaxDeliveryRetries {
		return b.DeadLetter(ctx, queueName, delivery, "retry_exhausted: "+errorText(cause))
	}

	publishing := publishingFromDelivery(delivery)
	publishing.Headers[retryCountHeader] = int32(DeliveryRetryCount(delivery.Headers) + 1)
	publishing.Headers[lastErrorHeader] = errorText(cause)
	if err := b.publish(ctx, "", RetryQueueName(queueName), publishing); err != nil {
		return requeueAfterPublishFailure(delivery, err)
	}
	return delivery.Ack(false)
}

// DeadLetter explicitly publishes a delivery to a queue-specific DLQ. This is
// intentionally explicit instead of adding x-dead-letter arguments to existing
// main queues, because changing arguments on an already-declared RabbitMQ queue
// causes a PRECONDITION_FAILED during rolling upgrades.
func (b *Broker) DeadLetter(ctx context.Context, queueName string, delivery amqp.Delivery, reason string) error {
	publishing := publishingFromDelivery(delivery)
	publishing.Headers[deadLetterReasonHeader] = truncateHeader(reason)
	publishing.Headers["x-devflow-original-routing-key"] = delivery.RoutingKey
	if err := b.publish(ctx, ExchangeDeadLetter, queueName, publishing); err != nil {
		return requeueAfterPublishFailure(delivery, err)
	}
	return delivery.Ack(false)
}

func RetryQueueName(queueName string) string {
	return queueName + retryQueueSuffix
}

func DeadLetterQueueName(queueName string) string {
	return queueName + deadLetterQueueSuffix
}

func DeliveryRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	switch value := headers[retryCountHeader].(type) {
	case int:
		return value
	case int8:
		return int(value)
	case int16:
		return int(value)
	case int32:
		return int(value)
	case int64:
		return int(value)
	case uint:
		return int(value)
	case uint8:
		return int(value)
	case uint16:
		return int(value)
	case uint32:
		return int(value)
	case uint64:
		if value > uint64(^uint(0)>>1) {
			return MaxDeliveryRetries
		}
		return int(value)
	default:
		return 0
	}
}

func (b *Broker) publish(ctx context.Context, exchange, routingKey string, publishing amqp.Publishing) error {
	if b == nil || b.conn == nil {
		return errors.New("rabbitmq broker is unavailable")
	}

	channel, err := b.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	if err := channel.Confirm(false); err != nil {
		return fmt.Errorf("enable rabbitmq publisher confirms: %w", err)
	}
	returns := channel.NotifyReturn(make(chan amqp.Return, 1))
	confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	closes := channel.NotifyClose(make(chan *amqp.Error, 1))

	if err := channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		true,
		false,
		publishing,
	); err != nil {
		return err
	}

	for {
		select {
		case returned, ok := <-returns:
			if !ok {
				returns = nil
				continue
			}
			return fmt.Errorf("%w: exchange=%q routing_key=%q reply=%d %s",
				ErrUnroutable, exchange, routingKey, returned.ReplyCode, returned.ReplyText)
		case confirmation, ok := <-confirms:
			if !ok {
				return errors.New("rabbitmq publisher confirmation channel closed")
			}
			if !confirmation.Ack {
				return ErrPublishNack
			}
			// RabbitMQ sends basic.return before basic.ack for an unroutable
			// mandatory message. Drain a buffered return before accepting ACK.
			select {
			case returned, ok := <-returns:
				if !ok {
					return nil
				}
				return fmt.Errorf("%w: exchange=%q routing_key=%q reply=%d %s",
					ErrUnroutable, exchange, routingKey, returned.ReplyCode, returned.ReplyText)
			default:
				return nil
			}
		case closeErr, ok := <-closes:
			if !ok || closeErr == nil {
				return errors.New("rabbitmq publish channel closed before confirmation")
			}
			return fmt.Errorf("rabbitmq publish channel closed before confirmation: %w", closeErr)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func publishingFromDelivery(delivery amqp.Delivery) amqp.Publishing {
	headers := make(amqp.Table, len(delivery.Headers)+1)
	for key, value := range delivery.Headers {
		headers[key] = value
	}
	return amqp.Publishing{
		Headers:         headers,
		ContentType:     delivery.ContentType,
		ContentEncoding: delivery.ContentEncoding,
		DeliveryMode:    amqp.Persistent,
		Priority:        delivery.Priority,
		CorrelationId:   delivery.CorrelationId,
		ReplyTo:         delivery.ReplyTo,
		Expiration:      delivery.Expiration,
		MessageId:       delivery.MessageId,
		Timestamp:       delivery.Timestamp,
		Type:            delivery.Type,
		UserId:          delivery.UserId,
		AppId:           delivery.AppId,
		Body:            delivery.Body,
	}
}

func requeueAfterPublishFailure(delivery amqp.Delivery, publishErr error) error {
	if nackErr := delivery.Nack(false, true); nackErr != nil {
		return errors.Join(publishErr, fmt.Errorf("requeue original delivery: %w", nackErr))
	}
	return publishErr
}

func errorText(err error) string {
	if err == nil {
		return "unspecified processing error"
	}
	return truncateHeader(err.Error())
}

func truncateHeader(value string) string {
	const maxHeaderLength = 512
	value = strings.TrimSpace(value)
	if len(value) <= maxHeaderLength {
		return value
	}
	return value[:maxHeaderLength]
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
	if err := channel.ExchangeDeclare(
		ExchangeDeadLetter,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}
	if err := channel.ExchangeDeclare(
		ExchangeRetryReturn,
		"direct",
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
	if err := channel.QueueBind(
		queueName,
		routingKey,
		ExchangeEvents,
		false,
		nil,
	); err != nil {
		return err
	}
	if err := channel.QueueBind(
		queueName,
		queueName,
		ExchangeRetryReturn,
		false,
		nil,
	); err != nil {
		return err
	}

	if _, err := channel.QueueDeclare(
		RetryQueueName(queueName),
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-message-ttl":             retryDelayMillis,
			"x-dead-letter-exchange":    ExchangeRetryReturn,
			"x-dead-letter-routing-key": queueName,
		},
	); err != nil {
		return err
	}

	if _, err := channel.QueueDeclare(
		DeadLetterQueueName(queueName),
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}
	return channel.QueueBind(
		DeadLetterQueueName(queueName),
		queueName,
		ExchangeDeadLetter,
		false,
		nil,
	)
}
