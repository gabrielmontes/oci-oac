package main

import (
	"fmt"
	"oac-client/cmd"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: no .env file found in the current directory.")
	}

	cmd.Execute()
}
