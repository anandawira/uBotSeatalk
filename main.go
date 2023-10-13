package main

import (
	"context"

	"github.com/anandawira/uBotSeatalk/internal/pkg/seatalkbot"
	"github.com/anandawira/uBotSeatalk/internal/service/facts"
)

func main() {
	factSvc := facts.NewService(seatalkbot.NewClient())
	err := factSvc.SendTodayFact(context.Background())
	if err != nil {
		panic(err)
	}
}
