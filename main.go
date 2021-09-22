package main

import "log"

var currentState *State
var currentCheckpoint map[string]*PoolAmount

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
	currentCheckpoint = make(map[string]*PoolAmount)
	if state == nil {
		log.Println("initalize new state")
		currentState = &State{
			CurrentBeacon:    startBeaconHeight,
			CheckpointBeacon: initTokensCheckpoint(startBeaconHeight),
		}
	} else {
		if startBeaconHeight < state.CurrentBeacon {
			panic("invalid start height")
		}
		if startBeaconHeight > state.CurrentBeacon+2 {
			log.Println("current state height is too far from startBeaconHeight ", startBeaconHeight-state.CurrentBeacon)
			log.Println("initalize new state")
			currentState = &State{
				CurrentBeacon:    startBeaconHeight,
				CheckpointBeacon: initTokensCheckpoint(startBeaconHeight),
			}
		} else {
			currentState = state
			currentCheckpoint, err = loadCheckPointAmount(currentState.CheckpointBeacon)
			if err != nil {
				panic(err)
			}
		}

	}
	err = startMonitoring()
	if err != nil {
		panic(err)
	}
}
