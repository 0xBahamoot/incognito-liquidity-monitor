package main

import "log"

var currentState *State
var currentCheckpoint *PoolAmount

func main() {
	readArgs()
	// go slackHook()
	err := openDB()
	if err != nil {
		panic(err)
	}
	state, err := loadState()
	if err != nil {
		panic(err)
	}
	if state == nil {
		log.Println("initalize new state")
		currentState = &State{
			CurrentBeacon:    startBeaconHeight,
			CheckpointBeacon: startBeaconHeight,
		}
	} else {
		currentState = state
		currentCheckpoint, err = loadCheckPointAmount(currentState.CheckpointBeacon)
		if err != nil {
			panic(err)
		}
	}
	err = startMonitoring()
	if err != nil {
		panic(err)
	}
}
