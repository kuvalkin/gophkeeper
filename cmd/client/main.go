package main

import (
	"fmt"
	"os"

	"github.com/kuvalkin/gophkeeper/internal/client/cmd"
	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
)

func main() {
	conf, err := cmd.NewConfig()
	if err != nil {
		fmt.Println("error reading config:", err)

		os.Exit(1)
	}

	c, err := container.New(conf)
	if err != nil {
		fmt.Println("cant init application service container:", err)

		os.Exit(1)
	}

	defer func() {
		err = c.Close()
		if err != nil {
			fmt.Println("error closing application service container:", err)
			os.Exit(1)
		}
	}()

	_ = cmd.NewRootCommand(c).Execute()
}
