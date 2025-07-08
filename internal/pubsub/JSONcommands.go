package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
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
	handler func(T),
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
			var output T
			err := json.Unmarshal(delivery.Body, &output)
			if err != nil {
				fmt.Printf("could not unmarshal message: %v\n", err)
				continue
			}

			handler(output)

			err = delivery.Ack(false)
			if err != nil {
				fmt.Printf("could not acknowledge message: %v\n", err)
				continue
			}
		}
	}()

	return nil
}
