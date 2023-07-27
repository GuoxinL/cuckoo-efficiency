/*
   Created by guoxin in 2023/7/26 09:54
*/
package main

import (
	"chainmaker.org/chainmaker/common/v2/birdsnest"
	"chainmaker.org/chainmaker/localconf/v2"
	"flag"
	"fmt"
	logger "github.com/GuoxinL/cuckoo-efficiency/log"
	"github.com/GuoxinL/cuckoo-efficiency/txfilter"
	"github.com/GuoxinL/cuckoo-efficiency/txfilter/filtercommon"
	"github.com/spf13/viper"
	"math"
	"sync"
	"time"
)

func main() {
	cfg, err := InitCfg()
	if err != nil {
		fmt.Println(err)
		return
	}
	log, _ := logger.GetLogger1(cfg.Logger)

	go Generate(log)

	config, err := filtercommon.ToPbConfig(cfg.TxFilter, "chain1")
	if err != nil {
		fmt.Println(err)
		return
	}
	filter, err := txfilter.Factory().NewTxFilter(config, log, &TestStore{})
	if err != nil {
		return
	}

	for height := uint64(0); height < math.MaxUint64; height++ {
		front := TxQueue.PopFront()
		if front == nil {
			time.Sleep(time.Second)
			continue
		}
		existsCosts := time.Now()
		keys := front.([]birdsnest.Key)
		batch := DispatchTxVerifyTask(keys)
		//f := &fileterpb.Stat{}
		var wg sync.WaitGroup
		waitCount := len(batch)
		wg.Add(waitCount)
		for i := 0; i < waitCount; i++ {
			index := i
			go func() {
				defer wg.Done()
				txs := batch[index]
				for j := 0; j < len(txs); j++ {
					_, _, _ = filter.IsExists(txs[j].String())
				}
			}()
		}
		wg.Wait()
		existsCostsT := time.Since(existsCosts)
		txIds := birdsnest.ToStrings(keys)
		addCosts := time.Now()
		err = filter.Adds(txIds)
		addCostsT := time.Since(addCosts)
		log.Infof("commit block [%v] exists: %v, add: %v, count: %v, error:%v", height, existsCostsT, addCostsT, len(keys), err)
	}

}
func CalcTxVerifyWorkers(txCount int) int {
	if txCount>>12 > 0 {
		// more than 4095, then use 100 workers
		return 100
	} else if txCount>>11 > 0 {
		// more than 2047, then use 50 workers
		return 50
	} else if txCount>>10 > 0 {
		// more than 1023, then use 20 workers
		return 20
	} else if txCount>>8 > 0 {
		// more than 255, then use 10 workers
		return 10
	} else if txCount>>7 > 0 {
		// more than 127, then use 8 workers
		return 8
	} else if txCount>>5 > 0 {
		// more than 31, then use 5 workers
		return 5
	}
	// else use only 1 worker
	return 1
}

func DispatchTxVerifyTask(txs []birdsnest.Key) map[int][]birdsnest.Key {
	txCount := len(txs)
	batchCount := CalcTxVerifyWorkers(txCount)
	batchSize := txCount / batchCount
	batch := make(map[int][]birdsnest.Key)
	for i := 0; i < batchCount-1; i++ {
		batch[i] = txs[i*batchSize : i*batchSize+batchSize]
	}
	batch[batchCount-1] = txs[(batchCount-1)*batchSize:]
	return batch
}

func InitCfg() (*RootConfig, error) {
	var configPath string
	flag.StringVar(&configPath, "config", "../config/application.yaml", "Specify the configuration file path. default: ../config/application.yaml")
	flag.Parse()

	var (
		err       error
		confViper *viper.Viper
	)

	if confViper, err = initViper(configPath); err != nil {
		return nil, fmt.Errorf("load sdk config failed, %s", err)
	}

	rootConf := &RootConfig{}
	if err = confViper.Unmarshal(rootConf); err != nil {
		return nil, fmt.Errorf("unmarshal config file failed, %s", err)
	}

	return rootConf, nil
}

type RootConfig struct {
	TxFilter localconf.TxFilterConfig `mapstructure:"tx_filter"`
	Logger   logger.LogCfg            `mapstructure:"log"`
}

func initViper(confPath string) (*viper.Viper, error) {
	cmViper := viper.New()
	cmViper.SetConfigFile(confPath)
	if err := cmViper.ReadInConfig(); err != nil {
		return nil, err
	}

	return cmViper, nil
}
