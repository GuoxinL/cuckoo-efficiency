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

var height uint64 = 0

var C = make(chan uint64, 1)

const shapeshift uint64 = 10000

const txcount = 10000

func Generate(log protocol.Logger) {
	pool, err := ants.NewPool(txcount)
	if err != nil {
		return
	}
	for ; height < math.MaxUint64; height++ {
		go func() {
			select {
			case C <- height:
			default:
			}
		}()
		var keys = make([]birdsnest.Key, txcount)
		wg := &sync.WaitGroup{}
		wg.Add(txcount)
		for txIndex := 0; txIndex < txcount; txIndex++ {
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
				// 1.6w*2 txpool and core
				if height > shapeshift {
					time.Sleep(time.Millisecond * 500)
				}
				break
			} else {
				if height > shapeshift {
					time.Sleep(time.Millisecond * 500)
				} else {
					time.Sleep(time.Microsecond * 10)
				}
			}
		}
	}
}

func TxPool(filter protocol.TxFilter, log protocol.Logger) {
	var h uint64
	for {
		select {
		case h = <-C:
		default:
			if h > shapeshift {
				key := birdsnest.GenTxId()
				_, _, _ = filter.IsExists(key, birdsnest.RuleType_AbsoluteExpireTime)
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
