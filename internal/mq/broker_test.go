package mq

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestDeliveryRetryCount(t *testing.T) {
	tests := []struct {
		name    string
		headers amqp.Table
		want    int
	}{
		{name: "missing", headers: nil, want: 0},
		{name: "int32", headers: amqp.Table{retryCountHeader: int32(2)}, want: 2},
		{name: "int64", headers: amqp.Table{retryCountHeader: int64(3)}, want: 3},
		{name: "invalid", headers: amqp.Table{retryCountHeader: "3"}, want: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := DeliveryRetryCount(test.headers); got != test.want {
				t.Fatalf("DeliveryRetryCount() = %d, want %d", got, test.want)
			}
		})
	}
}

func TestPublishJSONReturnsUnroutableError(t *testing.T) {
	broker, _ := openIntegrationBroker(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := broker.PublishJSON(ctx, "test.no.bound.route", map[string]string{"test": "mandatory"})
	if !errors.Is(err, ErrUnroutable) {
		t.Fatalf("PublishJSON() error = %v, want ErrUnroutable", err)
	}
}

func TestRetryEventuallyMovesDeliveryToDeadLetterQueue(t *testing.T) {
	broker, channel := openIntegrationBroker(t)
	purgeQueues(t, channel, QueueFeedDistribute)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := broker.PublishJSON(ctx, RoutingKeyPostPublished, PostPublishedEvent{
		EventID:   "retry-integration-event",
		PostID:    10,
		AuthorID:  20,
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("publish initial delivery: %v", err)
	}

	delivery := getDeliveryEventually(t, ctx, channel, QueueFeedDistribute)
	for retry := 1; retry <= MaxDeliveryRetries; retry++ {
		if err := broker.RetryOrDeadLetter(ctx, QueueFeedDistribute, delivery, errors.New("temporary database error")); err != nil {
			t.Fatalf("retry %d: %v", retry, err)
		}
		delivery = getDeliveryEventually(t, ctx, channel, QueueFeedDistribute)
		if got := DeliveryRetryCount(delivery.Headers); got != retry {
			t.Fatalf("retry count = %d, want %d", got, retry)
		}
	}

	if err := broker.RetryOrDeadLetter(ctx, QueueFeedDistribute, delivery, errors.New("database still unavailable")); err != nil {
		t.Fatalf("dead-letter exhausted delivery: %v", err)
	}
	deadLetter := getDeliveryEventually(t, ctx, channel, DeadLetterQueueName(QueueFeedDistribute))
	if got := DeliveryRetryCount(deadLetter.Headers); got != MaxDeliveryRetries {
		t.Fatalf("DLQ retry count = %d, want %d", got, MaxDeliveryRetries)
	}
	reason, _ := deadLetter.Headers[deadLetterReasonHeader].(string)
	if !strings.Contains(reason, "retry_exhausted") {
		t.Fatalf("DLQ reason = %q, want retry_exhausted", reason)
	}
	if err := deadLetter.Ack(false); err != nil {
		t.Fatalf("ack DLQ delivery: %v", err)
	}
}

func openIntegrationBroker(t *testing.T) (*Broker, *amqp.Channel) {
	t.Helper()
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		t.Skip("RABBITMQ_URL is empty; RabbitMQ integration test skipped")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	broker, err := Open(ctx, url)
	if err != nil {
		t.Fatalf("open RabbitMQ: %v", err)
	}
	channel, err := broker.conn.Channel()
	if err != nil {
		_ = broker.Close()
		t.Fatalf("open RabbitMQ test channel: %v", err)
	}
	t.Cleanup(func() {
		_ = channel.Close()
		_ = broker.Close()
	})
	return broker, channel
}

func purgeQueues(t *testing.T, channel *amqp.Channel, queueName string) {
	t.Helper()
	for _, name := range []string{queueName, RetryQueueName(queueName), DeadLetterQueueName(queueName)} {
		if _, err := channel.QueuePurge(name, false); err != nil {
			t.Fatalf("purge queue %s: %v", name, err)
		}
	}
}

func getDeliveryEventually(t *testing.T, ctx context.Context, channel *amqp.Channel, queueName string) amqp.Delivery {
	t.Helper()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		delivery, ok, err := channel.Get(queueName, false)
		if err != nil {
			t.Fatalf("get from queue %s: %v", queueName, err)
		}
		if ok {
			return delivery
		}
		select {
		case <-ctx.Done():
			t.Fatal(fmt.Sprintf("timed out waiting for queue %s", queueName))
		case <-ticker.C:
		}
	}
}
