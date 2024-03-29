/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package shardingbirdsnest transaction filter mplementation
package shardingbirdsnest

import (
	"bytes"
	"chainmaker.org/chainmaker/protocol/v2"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	bn "chainmaker.org/chainmaker/common/v2/birdsnest"
	sbn "chainmaker.org/chainmaker/common/v2/shardingbirdsnest"
	"chainmaker.org/chainmaker/pb-go/v2/txfilter"
	"github.com/GuoxinL/cuckoo-efficiency/txfilter/filtercommon"
)

const (
	shardingLogTemplate = "sharding[%d]%v "
)

// TxFilter Sharding transaction filter
type TxFilter struct {
	// log Log output protocol.Logger
	log protocol.Logger
	// bn Sharding Bird's Nest implementation
	bn *sbn.ShardingBirdsNest
	// store block store protocol.BlockchainStore
	store protocol.BlockchainStore
	// exitC Exit channel
	exitC chan struct{}
	// l read write lock
	l sync.RWMutex
}

// ValidateRule validate rules
func (f *TxFilter) ValidateRule(txId string, ruleType ...bn.RuleType) error {
	key, err := bn.ToTimestampKey(txId)
	if err != nil {
		return nil
	}
	err = f.bn.ValidateRule(key, ruleType...)
	if err != nil {
		return err
	}
	return nil
}

// New transaction filter init
func New(config *sbn.ShardingBirdsNestConfig, log protocol.Logger, store protocol.BlockchainStore) (protocol.TxFilter, error) {
	// Because it is compatible with Normal type, the transaction ID cannot be converted to time transaction ID, so the
	// database can be queried directly. Therefore, the transaction ID type is fixed as TimestampKey
	config.Birdsnest.Cuckoo.KeyType = bn.KeyType_KTTimestampKey

	initLasts := time.Now()
	exitC := make(chan struct{})
	shardingBirdsNest, err := sbn.NewShardingBirdsNest(config, exitC, bn.LruStrategy, sbn.NewModuloSA(int(config.Length)), filtercommon.NewLogger(log))
	if err != nil {
		if err != bn.ErrCannotModifyTheNestConfiguration {
			log.Errorf("new filter fail, error: %v", err)
			return nil, err
		}
		log.Warnf("new filter, %v", err)
	}
	txFilter := &TxFilter{
		log:   log,
		bn:    shardingBirdsNest,
		exitC: exitC,
		store: store,
	}
	shardingBirdsNest.Start()

	// Chase block height
	err = filtercommon.ChaseBlockHeight(store, txFilter, log)
	if err != nil {
		return nil, err
	}
	log.Infof("shading filter init success, sharding: %v, birdsnest: %v max keys: %v, cost: %v", config.Length, config.Birdsnest.Length, config.Birdsnest.Cuckoo.MaxNumKeys, time.Since(initLasts))
	return txFilter, nil
}

// GetHeight get height from transaction filter
func (f *TxFilter) GetHeight() uint64 {
	return f.bn.GetHeight()
}

// SetHeight set height from transaction filter
func (f *TxFilter) SetHeight(height uint64) {
	f.bn.SetHeight(height)
}

// IsExistsAndReturnHeight is exists and return height
func (f *TxFilter) IsExistsAndReturnHeight(txId string, ruleType ...bn.RuleType) (exists bool, height uint64, stat *txfilter.Stat, err error) {
	start := time.Now()
	defer filtercommon.ShowLog("IsExistsAndReturnHeight", txId, start, f.log)
	exists, stat, err = f.IsExists(txId, ruleType...)
	if err != nil {
		return false, 0, stat, err
	}
	return exists, f.GetHeight(), stat, nil
}

// Add txId to transaction filter
func (f *TxFilter) Add(txId string) error {
	start := time.Now()
	defer filtercommon.ShowLog("Add", txId, start, f.log)
	// Convert the transaction ID to TimestampKey
	key, err := bn.ToTimestampKey(txId)
	if err != nil {
		return nil
	}
	f.l.Lock()
	err = f.bn.Add(key)
	f.l.Unlock()
	if err != nil {
		f.log.Errorf("filter add fail, txid: %v error: %v", txId, err)
		return err
	}
	f.log.DebugDynamic(filtercommon.LoggingFixLengthFunc("filter add txid: %v, cost: %v", txId, time.Since(start)))
	return nil
}

// Adds batch Add txId
func (f *TxFilter) Adds(txIds []string) error {
	start := time.Now()
	defer filtercommon.ShowLog("Adds", strings.Join(txIds, ","), start, f.log)
	// Convert the transaction ID to TimestampKey
	timestampKeys, _ := bn.ToTimestampKeysAndNormalKeys(txIds)
	if len(timestampKeys) <= 0 {
		return nil
	}
	f.l.Lock()
	err := f.bn.Adds(timestampKeys)
	f.l.Unlock()
	if err != nil {
		f.log.Errorf("filter adds fail, ids: %v, error: %v", len(txIds), err)
		return err
	}
	f.addsPrintInfo(txIds, start)
	return nil
}

func (f *TxFilter) addsPrintInfo(txIds []string, start time.Time) {
	f.log.DebugDynamic(filtercommon.LoggingFixLengthFunc("filter adds success, ids: %v height: %v, cost: %v infos:%v ", len(txIds), f.GetHeight(), time.Since(start), func() string {
		var (
			bt    bytes.Buffer
			total uint64
		)
		for sharding, infos := range f.bn.Infos() {
			bt.WriteString(fmt.Sprintf(shardingLogTemplate, sharding, infos[3]))
			total += infos[3]
		}
		bt.WriteString("total:")
		bt.WriteString(strconv.FormatUint(total, 10))
		return bt.String()
	}()))
}

// AddsAndSetHeight batch add tx id and set height
func (f *TxFilter) AddsAndSetHeight(txIds []string, height uint64) error {
	start := time.Now()
	defer filtercommon.ShowLog("AddsAndSetHeight", strings.Join(txIds, ","), start, f.log)

	// Convert the transaction ID to TimestampKey
	timestampKeys, _ := bn.ToTimestampKeysAndNormalKeys(txIds)
	if len(timestampKeys) <= 0 {
		f.SetHeight(height)
		f.log.DebugDynamic(filtercommon.LoggingFixLengthFunc("adds and set height, no timestamp keys height: %d", height))
		return nil
	}
	f.l.Lock()
	err := f.bn.AddsAndSetHeight(timestampKeys, height)
	f.l.Unlock()
	if err != nil {
		return err
	}
	f.addsPrintInfo(txIds, start)
	return nil
}

// IsExists Check whether TxId exists in the transaction filter
func (f *TxFilter) IsExists(txId string, ruleType ...bn.RuleType) (exists bool, stat *txfilter.Stat, err error) {
	var (
		key      bn.TimestampKey
		contains bool
		costs    [][]time.Duration
		dbCost   time.Duration

		total time.Duration
		exist time.Duration

		before = make([]time.Duration, 0, 8)
		after  = make([]time.Duration, 2)
	)
	start := time.Now()
	defer func() {
		total = time.Since(start)
		if len(ruleType) != 0 {
			filtercommon.ShowExistsLog("TxPoolIsExists", exists, txId, total, before, exist, after, costs, f.log)
		} else {
			filtercommon.ShowExistsLog("IsExists", exists, txId, total, before, exist, after, costs, f.log)
		}
	}()

	// Convert the transaction ID to TimestampKey
	var costsa []time.Duration
	key, costsa, err = bn.ToTimestampKeyTime(txId)
	since := time.Since(start)
	before = append(before, costsa...)
	before = append(before, 0)
	before = append(before, since)
	if err != nil {
		exists, dbCost, err = f.findDb(txId)
		if err != nil {
			err = fmt.Errorf("%v, txid type: normal", err)
		}
		return exists, filtercommon.NewStat0(0, dbCost, nil), err
	}
	before = append(before, time.Since(start))
	f.l.RLock()
	defer f.l.RUnlock()
	before = append(before, time.Since(start))

	// TODO cuckoo 慢
	// cost(total:240.713523ms,before:1.429µs,exist:240.710835ms,after:903ns),
	// filters: [[97ns 591ns] [97ns 240.703359ms] [311ns 618ns] [74ns 612ns] [101ns 527ns] [76ns 565ns]]
	existStart := time.Now()
	contains, costs, err = f.bn.Contains(key, ruleType...)
	exist = time.Since(existStart)

	// TODO
	afterStart := time.Now()
	defer func() {
		after = append(after, time.Since(afterStart))
	}()
	if err != nil {
		// If not, query DB
		if err == bn.ErrKeyTimeIsNotInTheFilterRange {
			exists, dbCost, err = f.findDb(txId)
			if err != nil {
				err = fmt.Errorf("%v, key time is not in the filter range", err)
			}
			stat1 := filtercommon.NewStat1(exist, dbCost, costs)
			after[0] = time.Since(afterStart)
			return exists, stat1, err
		}
		f.log.Errorf("[%v] query from filter fail, error:%v", txId, err)
		stat0 := filtercommon.NewStat0(exist, 0, costs)
		after[1] = time.Since(afterStart)
		// TODO 频繁创建对象
		return contains, stat0, err
	}

	if contains {
		exists, dbCost, err = f.findDb(txId)
		if err != nil {
			err = fmt.Errorf("%v, %v positive", err, exists)
		}
		return exists, filtercommon.NewStat1(exist, dbCost, costs), err
	}

	//f.log.DebugDynamic(filtercommon.LoggingFixLengthFunc("[%v] does not exist in filter, cost: %v", txId, time.Since(start)))
	return contains, filtercommon.NewStat0(exist, 0, costs), nil
}

// Close transaction filter
func (f *TxFilter) Close() {
	close(f.exitC)
}

func (f *TxFilter) findDb(txId string) (bool, time.Duration, error) {
	start := time.Now()
	defer filtercommon.ShowLog("findDb", txId, start, f.log)

	exists, err := f.store.TxExists(txId)
	costs := time.Since(start)
	if err != nil {
		f.log.Errorf("[%v] filter check exists, query from db fail, error:%v", txId, err)
		return false, costs, err
	}
	return exists, costs, err
}
