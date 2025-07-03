package jobqueue

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client *redis.Client
	key    string
}

func NewRedisQueue(addr, password string, db int, key string) *RedisQueue {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisQueue{client: client, key: key}
}

func (q *RedisQueue) Enqueue(job *Job) error {
	b, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return q.client.LPush(context.Background(), q.key, b).Err()
}

func (q *RedisQueue) Dequeue() (*Job, error) {
	res, err := q.client.RPop(context.Background(), q.key).Result()
	if err != nil {
		return nil, err
	}
	var job Job
	if err := json.Unmarshal([]byte(res), &job); err != nil {
		return nil, err
	}
	return &job, nil
}
