package main

import (
	"github.com/amirdaaee/tbuljoi/internal/bot"
	"github.com/amirdaaee/tbuljoi/internal/logging"
)

func main() {
	logging.SetupLogger()
	bot.StartBot()
}
