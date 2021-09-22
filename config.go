package main

import (
	"flag"
	"log"
)

var slackHookToken string
var startBeaconHeight uint64

func readArgs() {
	argSlackHook := flag.String("slack", "", "set slack hook token")
	argStartBeaconHeight := flag.Uint64("beacon", 0, "set start beacon height")
	flag.Parse()
	slackHookToken = *argSlackHook
	startBeaconHeight = *argStartBeaconHeight
	if startBeaconHeight == 0 {
		h, err := detectBeaconStartPoint()
		if err != nil {
			log.Println("detectBeaconStartPoint", err)
			startBeaconHeight = DefaultBeaconCheckPoint
		} else {
			startBeaconHeight = h
		}

		log.Println("startBeaconHeight", startBeaconHeight)
	}
}
