package rabbitmq

import (
	"crypto/md5"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rishabh-sonic/orbit/pkg/config"
)

const (
	numShards = 16
)

type Client struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	cfg      *config.Config
}

func New(cfg *config.Config) (*Client, error) {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}

	client := &Client{conn: conn, ch: ch, cfg: cfg}
	if err := client.declareTopology(); err != nil {
		client.Close()
		return nil, err
	}
	return client, nil
}

func (c *Client) declareTopology() error {
	// Declare exchange
	if err := c.ch.ExchangeDeclare(
		c.cfg.RabbitMQExchange, "topic", true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	if c.cfg.RabbitMQShardingEnabled {
		// Declare 16 sharded queues
		for i := 0; i < numShards; i++ {
			qName := fmt.Sprintf("notifications-queue-%x", i)
			routingKey := fmt.Sprintf("notifications.shard.%x", i)
			if _, err := c.ch.QueueDeclare(qName, true, false, false, false, nil); err != nil {
				return fmt.Errorf("declare queue %s: %w", qName, err)
			}
			if err := c.ch.QueueBind(qName, routingKey, c.cfg.RabbitMQExchange, false, nil); err != nil {
				return fmt.Errorf("bind queue %s: %w", qName, err)
			}
		}
	} else {
		// Legacy single queue
		if _, err := c.ch.QueueDeclare("notifications-queue", true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare legacy queue: %w", err)
		}
		if err := c.ch.QueueBind(
			"notifications-queue", "notifications.routingkey", c.cfg.RabbitMQExchange, false, nil,
		); err != nil {
			return fmt.Errorf("bind legacy queue: %w", err)
		}
	}
	return nil
}

// Publish sends a message to the shard queue for the given target username.
func (c *Client) Publish(targetUsername string, body []byte) error {
	routingKey := routingKeyForUsername(targetUsername, c.cfg.RabbitMQShardingEnabled)
	return c.ch.Publish(
		c.cfg.RabbitMQExchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

// Consume returns a delivery channel for a specific shard queue (used by ws service).
func (c *Client) Consume(queueName string) (<-chan amqp.Delivery, error) {
	return c.ch.Consume(queueName, "", false, false, false, false, nil)
}

func (c *Client) Channel() *amqp.Channel {
	return c.ch
}

func (c *Client) Close() {
	if c.ch != nil {
		if err := c.ch.Close(); err != nil {
			slog.Error("rabbitmq channel close", "err", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			slog.Error("rabbitmq conn close", "err", err)
		}
	}
}

// ShardQueues returns all 16 queue names.
func ShardQueues() []string {
	queues := make([]string, numShards)
	for i := 0; i < numShards; i++ {
		queues[i] = fmt.Sprintf("notifications-queue-%x", i)
	}
	return queues
}

func routingKeyForUsername(username string, sharding bool) string {
	if !sharding {
		return "notifications.routingkey"
	}
	sum := md5.Sum([]byte(username))
	shardIdx := int(sum[0]) % numShards
	return fmt.Sprintf("notifications.shard.%x", shardIdx)
}
