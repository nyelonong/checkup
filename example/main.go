package main

import (
	"log"

	"github.com/nyelonong/checkup"
)

func main() {
	check, err := checkup.New("dep.yaml", true)
	if err != nil {
		log.Println(err)
	}

	if err := check.Checkup(); err != nil {
		log.Println(err)
	}
}
