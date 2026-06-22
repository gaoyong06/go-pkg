package mq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/go-kratos/kratos/v2/log"
)

// Producer RocketMQ 消息生产者
type Producer struct {
	p     rocketmq.Producer
	topic string
	log   *log.Helper
}

// NewProducer 创建 RocketMQ 生产者
func NewProducer(nameServer []string, producerGroup string, topic string, logger log.Logger) (*Producer, error) {
	p, err := rocketmq.NewProducer(
		producer.WithNameServer(nameServer),
		producer.WithGroupName(producerGroup),
		producer.WithRetry(2),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create RocketMQ producer: %w", err)
	}

	if err := p.Start(); err != nil {
		return nil, fmt.Errorf("failed to start RocketMQ producer: %w", err)
	}

	return &Producer{
		p:     p,
		topic: topic,
		log:   log.NewHelper(log.With(logger, "module", "mq/producer")),
	}, nil
}

// Publish 发布消息，routingKey 作为 Tag 使用
func (p *Producer) Publish(ctx context.Context, routingKey string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := &primitive.Message{
		Topic: p.topic,
		Body:  body,
	}
	if routingKey != "" {
		msg.WithTag(routingKey)
	}

	result, err := p.p.SendSync(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	p.log.Infof("Published message to topic=%s, tag=%s, msgId=%s", p.topic, routingKey, result.MsgID)
	return nil
}

// Close 关闭连接
func (p *Producer) Close() error {
	if p.p != nil {
		return p.p.Shutdown()
	}
	return nil
}
