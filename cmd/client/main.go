package main

import (
	"fmt"
	"log"

	"github.com/kuvalkin/gophkeeper/internal/client/cmd"
	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
)

func main() {
	conf, err := cmd.NewConfig()
	if err != nil {
		log.Fatal("error reading config:", err)
	}

	c, err := container.New(conf)
	if err != nil {
		log.Fatal("cant init application service container:", err)
		fmt.Println("cant init application service container:", err)
	}

	defer func() {
		err = c.Close()
		if err != nil {
			log.Fatal("error closing application service container:", err)
		}
	}()

	_ = cmd.NewRootCommand(c).Execute()
}
