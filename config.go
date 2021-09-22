package main

import (
	"flag"
)

var slackHookToken string
var startBeaconHeight uint64

func readArgs() {
	argSlackHook := flag.String("slack", "", "set slack hook token")
	argStartBeaconHeight := flag.Uint64("beacon", DefaultBeaconCheckPoint, "set start beacon height")
	flag.Parse()
	slackHookToken = *argSlackHook
	startBeaconHeight = *argStartBeaconHeight
}
