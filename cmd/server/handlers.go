package main

import (
	"fmt"

	"github.com/JLee871/learn-pub-sub-starter/internal/gamelogic"
	"github.com/JLee871/learn-pub-sub-starter/internal/pubsub"
	"github.com/JLee871/learn-pub-sub-starter/internal/routing"
)

func handlerLogs() func(gamelog routing.GameLog) pubsub.AckType {
	return func(gamelog routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		err := gamelogic.WriteLog(gamelog)
		if err != nil {
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
