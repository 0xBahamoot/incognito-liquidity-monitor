package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

func startMonitoring() error {
	c := time.NewTicker(beaconCheckInterval)
	for {
		<-c.C
		err := getBeaconBlockAndPDEState()
		if err != nil {
			return err
		}
	}
}

func getBeaconBlockAndPDEState() error {
	beaconHeight := currentState.CurrentBeacon + 1
	if currentCheckpoint == nil {
		beaconHeight = currentState.CurrentBeacon
	}
	var newCheckpoint *PoolAmount
	if currentState.CurrentBeacon <= PDEXV3Breakpoint {
		state, err := getNewState(beaconHeight)
		if err != nil {
			return err
		}
		log.Println("start processing state", beaconHeight)
		priceList, err := getPrice()
		if err != nil {
			return err
		}
		newPricePoint := &PriceHistory{
			Beacon: beaconHeight,
			Value:  priceList,
		}
		newPoolAmount, err := processStateV1(state, beaconHeight)
		if err != nil {
			return err
		}

		if currentCheckpoint == nil {
			newCheckpoint = newPoolAmount
			newPoolAmount.CheckpointBeacon = beaconHeight
			err = saveCheckPointPoolAmount(*newPoolAmount)
			if err != nil {
				panic(err)
			}
		} else {
			block, err := getBlock(beaconHeight)
			if err != nil {
				panic(err)
			}
			willResetCheckPoint := isContainLiquidityInstr(block.Instructions)
			if willResetCheckPoint {
				newCheckpoint = newPoolAmount
				newPoolAmount.CheckpointBeacon = beaconHeight
				err = saveCheckPointPoolAmount(*newPoolAmount)
				if err != nil {
					panic(err)
				}
			} else {
				newCheckpoint = currentCheckpoint
				newPoolAmount.CheckpointBeacon = currentCheckpoint.CheckpointBeacon
			}
		}
		changeHst, err := processChange(*newCheckpoint, *newPoolAmount)
		if err != nil {
			panic(err)
		}
		changeHst.Beacon = beaconHeight
		changeHst.CheckpointBeacon = newPoolAmount.CheckpointBeacon

		err = savePriceHistory(*newPricePoint)
		if err != nil {
			panic(err)
		}
		err = savePoolAmount(*newPoolAmount)
		if err != nil {
			panic(err)
		}
		err = saveChangeHistory(*changeHst)
		if err != nil {
			panic(err)
		}
		if newCheckpoint != nil {
			currentCheckpoint = newCheckpoint
			currentState.CheckpointBeacon = beaconHeight
		}
		currentState.CurrentBeacon = beaconHeight
		err = saveState(*currentState)
		if err != nil {
			panic(err)
		}
		log.Println("finished process state", currentState.CurrentBeacon)
		amount, alert := willAlert(changeHst, newPricePoint)
		if alert {
			if amount > 0 {
				sendSlackNoti(fmt.Sprintf("Change have exceeded %v 🤑, current change is %v ", ChangeThreshold, amount))
			} else {
				sendSlackNoti(fmt.Sprintf("Change have gone below -%v 👋 💰 😭, current change is %v ", ChangeThreshold, amount))
			}
		}
		sendSlackNoti(fmt.Sprintf("current change is %v ", amount))
		log.Println(fmt.Sprintf("Current change is %v ", amount))
	} else {
		//TODO for pdexV3
	}
	return nil
}

func processStateV1(state *jsonresult.CurrentPDEState, beaconHeight uint64) (*PoolAmount, error) {
	newPoolAmount, err := extractPoolAmounts(state)
	if err != nil {
		return nil, err
	}
	newPoolAmount.Beacon = beaconHeight
	return newPoolAmount, nil
}

func processStateV2() {
	//TODO for pdexV3
}

func getNewState(height uint64) (*jsonresult.CurrentPDEState, error) {
	beaconHeight := struct {
		BeaconHeight uint64
	}{
		BeaconHeight: height,
	}
	var state jsonresult.CurrentPDEState
	resultJson, err := SendQuery("getpdestate", []interface{}{beaconHeight})
	if err != nil {
		return nil, err
	}
	err = ParseResponse(resultJson, &state)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func getBlock(height uint64) (*jsonresult.GetBeaconBlockResult, error) {
	var blocks []*jsonresult.GetBeaconBlockResult
	var block *jsonresult.GetBeaconBlockResult
	resultJson, err := SendQuery("retrievebeaconblockbyheight", []interface{}{height, "2"})
	if err != nil {
		return nil, err
	}
	err = ParseResponse(resultJson, &blocks)
	if err != nil {
		return nil, err
	}
	if len(blocks) > 1 {
		panic("len(blocks) > 1")
	}
	block = blocks[0]
	return block, err
}

func extractPoolAmounts(state *jsonresult.CurrentPDEState) (*PoolAmount, error) {
	newPoolAmount := PoolAmount{
		Amount: make(map[string]float32),
	}
	tokenIDs := []string{USDC, USDT, BTC, DAI, ETH, BNB, XMR, LTC, DASH, MATIC, ZEC}
	for _, pool := range state.PDEPoolPairs {
		for _, v := range tokenIDs {
			if pool.Token1IDStr == v {
				if _, ok := newPoolAmount.Amount[v]; !ok {
					newPoolAmount.Amount[v] = float32(pool.Token1PoolValue) * float32(1e-9)
				} else {
					newPoolAmount.Amount[v] += float32(pool.Token1PoolValue) * float32(1e-9)
				}
			}
			if pool.Token2IDStr == v {
				if _, ok := newPoolAmount.Amount[v]; !ok {
					newPoolAmount.Amount[v] = float32(pool.Token2PoolValue) * float32(1e-9)
				} else {
					newPoolAmount.Amount[v] += float32(pool.Token2PoolValue) * float32(1e-9)
				}
			}
		}
	}
	if len(newPoolAmount.Amount) != len(tokenIDs) {
		log.Println("some pools is gone")
		return &newPoolAmount, nil
	}
	return &newPoolAmount, nil
}

func processChange(checkpoint PoolAmount, current PoolAmount) (*ChangeHistory, error) {
	changeHst := ChangeHistory{
		Value: make(map[string]float32),
	}
	for tokenID, amount := range checkpoint.Amount {
		newAmount := current.Amount[tokenID]
		value := newAmount - amount
		changeHst.Value[tokenID] = value
	}
	log.Printf("processed change for %v token\n", len(checkpoint.Amount))
	return &changeHst, nil
}

func willAlert(change *ChangeHistory, price *PriceHistory) (float32, bool) {
	totalValue := float32(0)
	for tokenID, value := range change.Value {
		p := price.Value[tokenID]
		totalValue += (value * p)
	}
	if math.Abs(float64(totalValue)) > float64(ChangeThreshold) {
		return totalValue, true
	}
	return totalValue, false
}

func isContainLiquidityInstr(instList [][]string) bool {
	for _, inst := range instList {
		metadataType, err := strconv.Atoi(inst[0])
		if err != nil {
			continue
		}
		contributionStatus := inst[2]
		switch metadataType {
		case metadata.PDEContributionMeta, metadata.PDEPRVRequiredContributionRequestMeta:
			if contributionStatus == common.PDEContributionMatchedChainStatus || contributionStatus == common.PDEContributionMatchedNReturnedChainStatus {
				return true
			}
		case metadata.PDEWithdrawalRequestMeta:
			if contributionStatus != common.PDEWithdrawalRejectedChainStatus {
				return true
			}
		}
	}
	return false
}

func detectBeaconStartPoint() (uint64, error) {
	var result uint64

	return result, nil
}
