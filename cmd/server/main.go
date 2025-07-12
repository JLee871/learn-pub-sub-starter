package main

import (
	"fmt"
	"log"

	"github.com/JLee871/learn-pub-sub-starter/internal/gamelogic"
	"github.com/JLee871/learn-pub-sub-starter/internal/pubsub"
	"github.com/JLee871/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	const rabbitConn = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConn)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	fmt.Println("Connection was successful.")

	channel, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	_, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.Durable,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.Durable,
		handlerLogs(),
	)
	if err != nil {
		log.Fatal(err)
	}

	gamelogic.PrintServerHelp()

InfiniteLoop:
	for {
		strings := gamelogic.GetInput()
		if len(strings) == 0 {
			continue
		}

		first := strings[0]
		switch first {
		case "pause":
			fmt.Println("Sending a pause message.")
			err := pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			if err != nil {
				fmt.Printf("error occured: %v\n", err)
				continue
			}

		case "resume":
			fmt.Println("Sending a resume message.")
			err := pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			if err != nil {
				fmt.Printf("error occured: %v\n", err)
				continue
			}

		case "quit":
			fmt.Println("Exiting... connection was shutdown.")
			break InfiniteLoop

		default:
			fmt.Println("Unfamiliar command.")
		}
	}
}
