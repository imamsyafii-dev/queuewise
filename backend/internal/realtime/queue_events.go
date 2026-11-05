package realtime

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type QueueEvents struct {
	client *redis.Client
}

func NewQueueEvents(client *redis.Client) *QueueEvents {
	return &QueueEvents{client: client}
}

func (q *QueueEvents) Publish(ctx context.Context, branchSlug string) error {
	return q.client.Publish(ctx, channel(branchSlug), "updated").Err()
}

func (q *QueueEvents) Subscribe(ctx context.Context, branchSlug string) *redis.PubSub {
	return q.client.Subscribe(ctx, channel(branchSlug))
}

func channel(branchSlug string) string {
	return fmt.Sprintf("queuewise:queue:%s", branchSlug)
}
