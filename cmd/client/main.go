package main

import (
	"github.com/kuvalkin/gophkeeper/internal/client/cmd"
)

func main() {
	_ = cmd.NewRootCommand().Execute()
}
