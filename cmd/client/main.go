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
	fmt.Println("Starting Peril client...")

	const rabbitConn = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConn)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Println("Peril game client connected to RabbitMQ!")

	name, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	gameState := gamelogic.NewGameState(name)

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gameState.GetUsername(),
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatal(err)
	}

InfiniteLoop:
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		command := words[0]

		switch command {
		case "spawn":
			err := gameState.CommandSpawn(words)
			if err != nil {
				fmt.Println(err)
				continue
			}

		case "move":
			move, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("Move to %v was successful.\n", move.ToLocation)

		case "status":
			gameState.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("Spamming not allowed yet!")

		case "quit":
			gamelogic.PrintQuit()
			break InfiniteLoop

		default:
			fmt.Println("Unfamiliar command.")
		}
	}

	fmt.Println("\nClient was closed.")
}
