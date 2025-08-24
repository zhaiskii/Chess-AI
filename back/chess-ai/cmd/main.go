package main

import (
	"fmt"
	"net/http"
	"chess-ai/internal/config"
)

func main() {
	cfg, err := config.MustLoad()
	if err != nil {
		fmt.Errorf("failed loading config", err.Error())
	}
}