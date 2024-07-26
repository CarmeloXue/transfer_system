package queue

import (
	"context"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

func NewTransactionProducer(listener primitive.TransactionListener) (rocketmq.TransactionProducer, error) {
	return nil, nil
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
