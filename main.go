package main

import (
	"fmt"
	"os"

	"github.com/nikofil/matrix-wc-bot/bot"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Printf("Args: %s url username password device_name\n", os.Args[0])
		return
	}
	url := os.Args[1]
	user := os.Args[2]
	pass := os.Args[3]
	devName := os.Args[4]
	if bot, err := bot.NewWCBot(url, user, pass, devName); err == nil {
		err := bot.Run()
		fmt.Println("Exiting:", err)
	} else {
		fmt.Println(err)
	}
}
