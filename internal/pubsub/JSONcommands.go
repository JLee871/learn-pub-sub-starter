package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	bytes, err := json.Marshal(val)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        bytes,
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
	if err != nil {
		return err
	}

	return nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	ch, queue, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueType,
	)
	if err != nil {
		return err
	}

	deliveries, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		defer ch.Close()
		for delivery := range deliveries {
			//json decoding
			var output T
			err := json.Unmarshal(delivery.Body, &output)
			if err != nil {
				fmt.Printf("could not unmarshal message: %v\n", err)
				continue
			}

			ackType := handler(output)

			switch ackType {
			case Ack:
				err = delivery.Ack(false)
				fmt.Println("message acknowledged")
				fmt.Print("> ")
			case NackRequeue:
				err = delivery.Nack(false, true)
				fmt.Println("message not acknowledged, requeued")
				fmt.Print("> ")
			case NackDiscard:
				err = delivery.Nack(false, false)
				fmt.Println("message not acknowledged, discarded")
				fmt.Print("> ")
			}

			if err != nil {
				fmt.Printf("could not acknowledge message: %v\n", err)
				continue
			}
		}
	}()

	return nil
}
