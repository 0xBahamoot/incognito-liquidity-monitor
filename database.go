package main

import (
	"encoding/json"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	lvdbErrors "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var lvdb *leveldb.DB

func openDB() error {
	handles := -1
	cache := 8
	userDBPath := "userdb"
	db, err := leveldb.OpenFile(userDBPath, &opt.Options{
		OpenFilesCacheCapacity: handles,
		BlockCacheCapacity:     cache / 2 * opt.MiB,
		WriteBuffer:            cache * opt.MiB, // Two of these are used internally
		Filter:                 filter.NewBloomFilter(10),
	})
	if _, corrupted := err.(*lvdbErrors.ErrCorrupted); corrupted {
		db, err = leveldb.RecoverFile(userDBPath, nil)
		if err != nil {
			return err
		}
	}
	lvdb = db
	return nil
}

func loadState() (*State, error) {
	var result State
	value, err := lvdb.Get([]byte(prefixState), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	err = json.Unmarshal(value, &result)
	return &result, err
}

func saveState(state State) error {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		return err
	}
	key := []byte(prefixState)
	return lvdb.Put(key, stateBytes, nil)
}

func savePriceHistory(price PriceHistory) error {
	key := []byte(prefixPriceHistory)
	key = append(key, []byte(fmt.Sprintf("%v", price.Beacon))...)
	dataBytes, err := json.Marshal(price)
	if err != nil {
		return err
	}
	return lvdb.Put(key, dataBytes, nil)
}

func saveChangeHistory(change ChangeHistory) error {
	key := []byte(prefixChangeHistory)
	key = append(key, []byte(fmt.Sprintf("%v", change.Beacon))...)
	dataBytes, err := json.Marshal(change)
	if err != nil {
		return err
	}
	return lvdb.Put(key, dataBytes, nil)
}

func savePoolAmount(pool PoolAmount) error {
	key := []byte(prefixPoolAmount)
	key = append(key, []byte(fmt.Sprintf("%v", pool.Beacon))...)
	dataBytes, err := json.Marshal(pool)
	if err != nil {
		return err
	}
	return lvdb.Put(key, dataBytes, nil)
}

func getPoolAmount(height uint64) (*PoolAmount, error) {
	var result PoolAmount
	key := []byte(prefixPoolAmount)
	key = append(key, []byte(fmt.Sprintf("%v", height))...)
	value, err := lvdb.Get([]byte(prefixState), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	err = json.Unmarshal(value, &result)
	return &result, err
}

func loadCheckPointAmount(checkpoints map[string]uint64) (map[string]*PoolAmount, error) {
	result := make(map[string]*PoolAmount)
	for tokenID, v := range checkpoints {
		pool, err := getPoolAmount(v)
		if err != nil {
			return nil, err
		}
		result[tokenID] = pool
	}
	return result, nil
}
