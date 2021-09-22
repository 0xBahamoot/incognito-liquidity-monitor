package main

type State struct {
	CurrentBeacon    uint64
	CheckpointBeacon map[string]uint64
}

type PoolAmount struct {
	Amount map[string]float32 //already multiplied by 1e-9
	Beacon uint64
}

type PriceHistory struct {
	Beacon uint64
	Value  map[string]float32
}

type ChangeHistory struct {
	Value            map[string]float32
	Beacon           uint64
	CheckpointBeacon map[string]uint64
}
