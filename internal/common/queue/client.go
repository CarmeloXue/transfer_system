package queue

import (
	"context"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

func NewProducer() (rocketmq.Producer, error) {
	p, err := rocketmq.NewProducer(
		producer.WithNameServer([]string{"localhost:9876"}),
		producer.WithGroupName("testGroup"),
	)
	if err != nil {
		return nil, err
	}
	err = p.Start()
	if err != nil {
		fmt.Printf("start producer error: %s\n", err.Error())
		return nil, err
	}
	return p, nil
}

type ConsumeHandler func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error)

func NewConsumer(h ConsumeHandler) error {
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{"localhost:9876"}),
		consumer.WithGroupName("testGroup"),
	)
	if err != nil {
		return err
	}

	err = c.Subscribe("test", consumer.MessageSelector{}, h)
	if err != nil {
		fmt.Printf("subscribe error: %s\n", err.Error())
		return err
	}

	err = c.Start()
	if err != nil {
		fmt.Printf("start consumer error: %s\n", err.Error())
		return err
	}
	return nil
}
