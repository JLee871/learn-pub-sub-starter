package main

import (
	"fmt"
	"time"

	"github.com/JLee871/learn-pub-sub-starter/internal/gamelogic"
	"github.com/JLee871/learn-pub-sub-starter/internal/pubsub"
	"github.com/JLee871/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)

		switch moveOutcome {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.Ack

		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack

		case gamelogic.MoveOutcomeMakeWar:
			key := routing.WarRecognitionsPrefix + "." + gs.GetUsername()
			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				key,
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		default:
			return pubsub.NackDiscard
		}
	}
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(war gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(war)
		log := &routing.GameLog{
			CurrentTime: time.Now(),
			Username:    gs.GetUsername(),
		}
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue

		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard

		case gamelogic.WarOutcomeOpponentWon:
			log.Message = fmt.Sprintf("%v won a war against %v", winner, loser)
			err := publishGameLog(log, ch)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		case gamelogic.WarOutcomeYouWon:
			log.Message = fmt.Sprintf("%v won a war against %v", winner, loser)
			err := publishGameLog(log, ch)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		case gamelogic.WarOutcomeDraw:
			log.Message = fmt.Sprintf("A war between %v and %v resulted in a draw.", winner, loser)
			err := publishGameLog(log, ch)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		default:
			fmt.Println("error: unknown war outcome")
			return pubsub.NackDiscard
		}
	}
}

func publishGameLog(gl *routing.GameLog, ch *amqp.Channel) error {
	err := pubsub.PublishGob(
		ch,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+gl.Username,
		gl,
	)
	return err
}
