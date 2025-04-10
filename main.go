package main

import (
	
	"log"
	"fmt"
	
	"github.com/dubbersthehoser/gator/internal/config"
)

func main() {

	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	cfg.SetUser("brandon")

	cfg, err = config.Read()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", cfg)
}
