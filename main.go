package main

import (
	"conversions/app"
	"fmt"
	"log"
)

func main() {
	cmd := app.NewServerCommand()
     fmt.Printf("1")
	 log.Print("xxxx")
	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}