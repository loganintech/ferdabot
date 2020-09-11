package main

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	bot := NewBot()

	if err := bot.Setup(); err != nil {
		fmt.Printf("Error occurred starting the bot %s\n", err)
		os.Exit(1)
	}

	if closeErr := bot.Run(); closeErr != nil {
		fmt.Printf("Error closing connection %s\n", closeErr.Error())
		os.Exit(2)
	}

}
