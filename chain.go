package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
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

		block, err := getBlock(beaconHeight)
		if err != nil {
			panic(err)
		}
		resetTokenList := resetCheckPointForToken(block.Instructions, state)

		for tokenID, _ := range newPoolAmount.Amount {
			if _, ok := currentCheckpoint[tokenID]; !ok {
				currentCheckpoint[tokenID] = newPoolAmount
			} else {
				willReset := false
				for _, v := range resetTokenList {
					if v == tokenID {
						willReset = true
					}
				}
				if willReset {
					currentCheckpoint[tokenID] = newPoolAmount
				}
			}
		}

		changeHst, err := processChange(currentCheckpoint, *newPoolAmount)
		if err != nil {
			panic(err)
		}
		changeHst.Beacon = beaconHeight
		changeHst.CheckpointBeacon = currentState.CheckpointBeacon

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
		// sendSlackNoti(fmt.Sprintf("current change is %v ", amount))
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
retry:
	var state jsonresult.CurrentPDEState
	resultJson, err := SendQuery("getpdestate", []interface{}{beaconHeight})
	if err != nil {
		if strings.Contains(err.Error(), "Can't found ConsensusStateRootHash of beacon height") {
			log.Println("retry getNewState")
			goto retry
		}
		return nil, err
	}
	err = ParseResponse(resultJson, &state)
	if err != nil {
		if strings.Contains(err.Error(), "Can't found ConsensusStateRootHash of beacon height") {
			log.Println("retry getNewState")
			goto retry
		}
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

func processChange(checkpoint map[string]*PoolAmount, current PoolAmount) (*ChangeHistory, error) {
	changeHst := ChangeHistory{
		Value: make(map[string]float32),
	}
	for tokenID, amount := range current.Amount {
		checkpointAmount := checkpoint[tokenID].Amount[tokenID]
		value := amount - checkpointAmount
		changeHst.Value[tokenID] = value
	}
	log.Printf("processed change for %v token\n", len(current.Amount))
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

func resetCheckPointForToken(instList [][]string, state *jsonresult.CurrentPDEState) []string {
	var result []string
	tokenIDs := []string{USDC, USDT, BTC, DAI, ETH, BNB, XMR, LTC, DASH, MATIC, ZEC}
	for _, inst := range instList {
		metadataType, err := strconv.Atoi(inst[0])
		if err != nil {
			continue
		}
		contributionStatus := inst[2]
		switch metadataType {
		case metadata.PDEContributionMeta, metadata.PDEPRVRequiredContributionRequestMeta:
			if contributionStatus == common.PDEContributionMatchedChainStatus {
				var md metadata.PDEMatchedContribution
				err := json.Unmarshal([]byte(inst[3]), &md)
				if err != nil {
					panic(err)
				}
				for _, v := range tokenIDs {
					if md.TokenIDStr == v {
						result = append(result, v)
					}
				}
			}
			if contributionStatus == common.PDEContributionMatchedNReturnedChainStatus {
				var md metadata.PDEMatchedNReturnedContribution
				err := json.Unmarshal([]byte(inst[3]), &md)
				if err != nil {
					panic(err)
				}
				for _, v := range tokenIDs {
					if md.TokenIDStr == v {
						result = append(result, v)
					}
				}
			}
		case metadata.PDEWithdrawalRequestMeta:
			if contributionStatus != common.PDEWithdrawalRejectedChainStatus {
				contentBytes := []byte(inst[3])
				var md metadata.PDEWithdrawalAcceptedContent
				err = json.Unmarshal(contentBytes, &md)
				if err != nil {
					panic(err)
				}
				for _, v := range tokenIDs {
					if md.PairToken1IDStr == v || md.PairToken2IDStr == v {
						result = append(result, v)
					}
				}
			}
		}
	}
	log.Println("token need to reset", len(result))
	return result
}

func detectBeaconStartPoint() (uint64, error) {
	var result uint64

	var info *jsonresult.GetBlockChainInfoResult
	resultJson, err := SendQuery("getblockchaininfo", []interface{}{})
	if err != nil {
		return 0, err
	}
	err = ParseResponse(resultJson, &info)
	if err != nil {
		return 0, err
	}
	result = info.BestBlocks[-1].Height
	return result, nil
}
