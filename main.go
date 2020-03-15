package main

import (
	"fmt"

	"github.com/nikofil/matrix-wc-bot/bot"
)

func main() {

	if bot, err := bot.NewWCBot("http://localhost:8008", "user", "user"); err == nil {
		bot.Run()
	} else {
		fmt.Println(err)
	}
}
