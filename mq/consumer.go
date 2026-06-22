package mq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/go-kratos/kratos/v2/log"
)

// Consumer RocketMQ 消息消费者
type Consumer struct {
	c             rocketmq.PushConsumer
	topic         string
	consumerGroup string
	log           *log.Helper
}

// NewConsumer 创建 RocketMQ 消费者；routingKeys 作为 Tag 使用，多个 Tag 用 || 连接
func NewConsumer(nameServer []string, consumerGroup string, topic string, routingKeys []string, logger log.Logger) (*Consumer, error) {
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNsResolver(primitive.NewPassthroughResolver(nameServer)),
		consumer.WithGroupName(consumerGroup),
		consumer.WithConsumerModel(consumer.Clustering),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create RocketMQ consumer: %w", err)
	}

	if err := c.Start(); err != nil {
		return nil, fmt.Errorf("failed to start RocketMQ consumer: %w", err)
	}

	return &Consumer{
		c:             c,
		topic:         topic,
		consumerGroup: consumerGroup,
		log:           log.NewHelper(log.With(logger, "module", "mq/consumer")),
	}, nil
}

// Consume 消费消息并阻塞直到 ctx 取消
func (c *Consumer) Consume(ctx context.Context, handler func(ctx context.Context, body []byte) error, routingKeys ...string) error {
	var selector consumer.MessageSelector
	if len(routingKeys) > 0 {
		tagExpr := ""
		for i, tag := range routingKeys {
			if i > 0 {
				tagExpr += " || "
			}
			tagExpr += tag
		}
		selector = consumer.MessageSelector{
			Type:       consumer.TAG,
			Expression: tagExpr,
		}
		c.log.Infof("Using message selector: type=TAG, expression=%s", tagExpr)
	} else {
		selector = consumer.MessageSelector{
			Type:       consumer.TAG,
			Expression: "*",
		}
	}

	err := c.c.Subscribe(c.topic, selector, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			c.log.Infof("Received message: topic=%s, tag=%s, msgId=%s", msg.Topic, msg.GetTags(), msg.MsgId)
			if err := handler(ctx, msg.Body); err != nil {
				c.log.Errorf("Failed to handle message: %v", err)
				return consumer.ConsumeRetryLater, err
			}
		}
		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		return fmt.Errorf("failed to register message handler: %w", err)
	}

	c.log.Infof("Started consuming messages from topic=%s, consumerGroup=%s, selector=%s", c.topic, c.consumerGroup, selector.Expression)
	<-ctx.Done()
	c.log.Info("Consumer context cancelled, stopping...")
	return nil
}

// Close 关闭连接
func (c *Consumer) Close() error {
	if c.c != nil {
		return c.c.Shutdown()
	}
	return nil
}

// UnmarshalMessage 解析消息 JSON
func UnmarshalMessage(body []byte, v interface{}) error {
	return json.Unmarshal(body, v)
}
