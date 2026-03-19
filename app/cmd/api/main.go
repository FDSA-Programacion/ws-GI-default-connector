package main

import (
	"log"
	"os"
	_ "time/tzdata"
	"ws-int-httr/cmd/api/bootstrap"
)

func main() {
	if err := bootstrap.Run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
