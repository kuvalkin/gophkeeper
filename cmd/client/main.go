package main

import (
	"fmt"
	stdLog "log"

	"github.com/kuvalkin/gophkeeper/internal/support/log"
)

func main() {
	err := log.InitLogger()
	if err != nil {
		stdLog.Fatal(fmt.Errorf("failed to initialize logger: %w", err))
	}

	defer func() {
		err = log.Logger().Sync()
		if err != nil {
			stdLog.Println(fmt.Errorf("failed to sync logger: %w", err))
		}
	}()

	log.Logger().Info("client")
}
