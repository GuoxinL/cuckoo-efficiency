/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package txfilter transaction filter factory
package txfilter

import (
	"sync"

	"github.com/GuoxinL/cuckoo-efficiency/txfilter/filtercommon"

	mapimpl "github.com/GuoxinL/cuckoo-efficiency/txfilter/map"

	"chainmaker.org/chainmaker/protocol/v2"
	"github.com/GuoxinL/cuckoo-efficiency/txfilter/birdnest"
	"github.com/GuoxinL/cuckoo-efficiency/txfilter/filterdefault"
	"github.com/GuoxinL/cuckoo-efficiency/txfilter/shardingbirdsnest"
)

// txFilterFactory Transaction filter factory
type txFilterFactory struct {
}

var once sync.Once
var _instance *txFilterFactory

// Factory return the global tx filter factory.
//nolint: revive
func Factory() *txFilterFactory {
	once.Do(func() { _instance = new(txFilterFactory) })
	return _instance
}

// NewTxFilter new transaction filter
func (cf *txFilterFactory) NewTxFilter(conf *filtercommon.TxFilterConfig, log protocol.Logger,
	store protocol.BlockchainStore) (protocol.TxFilter, error) {
	if conf == nil {
		log.Warn("txfilter conf is nil, use default type: store")
		return filterdefault.New(store), nil
	}
	switch conf.Type {
	// default txfilter
	case filtercommon.TxFilterTypeDefault:
		return filterdefault.New(store), nil
		// bird's nest txfilter
	case filtercommon.TxFilterTypeBirdsNest:
		return birdnest.New(conf.BirdsNest, log, store)
		// map txfilter
	case filtercommon.TxFilterTypeMap:
		return mapimpl.New(), nil
		// sharding bird's nest txfilter
	case filtercommon.TxFilterTypeShardingBirdsNest:
		return shardingbirdsnest.New(conf.ShardingBirdsNest, log, store)
	default:
		log.Warnf("txfilter type: %v not support, use default type: store", conf.Type)
		return filterdefault.New(store), nil
	}
}
