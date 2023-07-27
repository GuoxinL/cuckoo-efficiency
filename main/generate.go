/*
   Created by guoxin in 2023/7/26 11:37
*/
package main

import (
	"chainmaker.org/chainmaker/common/v2/birdsnest"
	"chainmaker.org/chainmaker/protocol/v2"
	"github.com/panjf2000/ants/v2"
	"github.com/phf/go-queue/queue"
	"math"
	"sync"
	"time"
)

var TxQueue = queue.New()

func Generate(log protocol.Logger) {
	pool, err := ants.NewPool(10000)
	if err != nil {
		return
	}
	for height := uint64(0); height < math.MaxUint64; height++ {
		var keys = make([]birdsnest.Key, 10000)
		wg := &sync.WaitGroup{}
		wg.Add(10000)
		for txIndex := 0; txIndex < 10000; txIndex++ {
			err = pool.Submit(generateKey(txIndex, keys, wg))
			if err != nil {
				log.Error(err)
				return
			}
		}
		wg.Wait()
		for {
			if TxQueue.Len() < 20 {
				TxQueue.PushBack(keys)
				break
			} else {
				time.Sleep(time.Second)
			}
		}
	}
}

func generateKey(txIndex int, keys []birdsnest.Key, wg *sync.WaitGroup) func() {
	return func() {
		defer wg.Done()
		keys[txIndex] = birdsnest.GenTimestampKey()
	}
}
