package mq

import "context"

type Publisher struct {
	broker *Broker
}

func NewPublisher(broker *Broker) *Publisher {
	if broker == nil {
		return nil
	}
	return &Publisher{broker: broker}
}

func (p *Publisher) PublishNotification(ctx context.Context, event NotificationEvent) error {
	if p == nil {
		return nil
	}
	return p.broker.PublishJSON(ctx, notificationRoutingKey(event.Type), event)
}

func (p *Publisher) PublishPostPublished(ctx context.Context, event PostPublishedEvent) error {
	if p == nil {
		return nil
	}
	return p.broker.PublishJSON(ctx, RoutingKeyPostPublished, event)
}

func notificationRoutingKey(notificationType string) string {
	switch notificationType {
	case "follow":
		return RoutingKeyUserFollowed
	case "like":
		return RoutingKeyInteractionLiked
	case "favorite":
		return RoutingKeyInteractionFavorited
	case "comment":
		return RoutingKeyInteractionCommented
	default:
		return "interaction.unknown"
	}
}
