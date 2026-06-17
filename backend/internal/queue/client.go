package queue

import (
        "github.com/hibiken/asynq"
)

type Client struct {
        client *asynq.Client
}

func NewClient(redisAddr string) *Client {
        return &Client{
                client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
        }
}

func (c *Client) Close() error {
        return c.client.Close()
}

func (c *Client) EnqueueFileProcessing(payload FileProcessingPayload) error {
        task, err := NewFileProcessingTask(payload)
        if err != nil {
                return err
        }
        _, err = c.client.Enqueue(task, asynq.Queue("default"))
        return err
}
