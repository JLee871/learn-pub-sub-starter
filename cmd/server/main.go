package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

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

	pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("\nConnection was shutdown.")
}
