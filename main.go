package main

import (
	"conversions/app"
	"fmt"
	_ "k8s.io/api/core/v1"
	"log"
)

func main() {
	cmd := app.NewServerCommand()
 fmt.Printf("1")
	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}