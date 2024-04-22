package main

import (
	"conversions/app"
	_ "k8s.io/api/core/v1"
	"log"
)

func main() {
	cmd := app.NewServerCommand()

	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}