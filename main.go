package main

import (
	"fmt"

	"github.com/nikofil/matrix-wc-bot/bot"
)

func main() {
	if bot, err := bot.NewWCBot("http://localhost:8008", "user", "user", "matrix-wc-bot"); err == nil {
		fmt.Println("Exiting:", bot.Run())
	} else {
		fmt.Println(err)
	}
}
