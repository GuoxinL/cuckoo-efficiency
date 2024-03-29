/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package txfilter transaction filter factory test
package txfilter

import (
	"chainmaker.org/chainmaker/protocol/v2"
	"reflect"
	"strconv"
	"testing"

	bn "chainmaker.org/chainmaker/common/v2/birdsnest"
	sbn "chainmaker.org/chainmaker/common/v2/shardingbirdsnest"
	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2/mock"
	"github.com/GuoxinL/cuckoo-efficiency/txfilter/birdnest"
	"github.com/GuoxinL/cuckoo-efficiency/txfilter/filtercommon"
	"github.com/GuoxinL/cuckoo-efficiency/txfilter/filterdefault"
	mapimpl "github.com/GuoxinL/cuckoo-efficiency/txfilter/map"
	"github.com/GuoxinL/cuckoo-efficiency/txfilter/shardingbirdsnest"
	"github.com/golang/mock/gomock"
)

func Test_txFilterFactory_NewTxFilter(t *testing.T) {
	type args struct {
		conf  *filtercommon.TxFilterConfig
		log   protocol.Logger
		store protocol.BlockchainStore
	}

	var (
		log   = newMockLogger(t)
		store = newMockBlockchainStore(t)
		block = createBlockByHash(0, []byte("123456"))
	)

	store.EXPECT().GetLastBlock().Return(block, nil).AnyTimes()
	log.EXPECT().DebugDynamic(gomock.Any()).AnyTimes()
	log.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
	defaultConf := GetTestDefaultConfig("test", 0)

	log.EXPECT().Warn(gomock.Any()).AnyTimes()
	tests := []struct {
		name    string
		args    args
		want    protocol.TxFilter
		wantErr bool
	}{
		{
			name: "test0",
			args: args{
				conf:  nil,
				log:   log,
				store: store,
			},
			want: func() protocol.TxFilter {
				return filterdefault.New(store)
			}(),
			wantErr: false,
		},
		{
			name: "test1",
			args: args{
				conf: &filtercommon.TxFilterConfig{
					Type:      filtercommon.TxFilterTypeDefault,
					BirdsNest: defaultConf.Birdsnest,
				},
				log:   log,
				store: store,
			},
			want: func() protocol.TxFilter {
				return filterdefault.New(store)
			}(),
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				conf: &filtercommon.TxFilterConfig{
					Type:      filtercommon.TxFilterTypeBirdsNest,
					BirdsNest: defaultConf.Birdsnest,
				},
				log:   log,
				store: store,
			},
			want: func() protocol.TxFilter {
				txFilter, err := birdnest.New(defaultConf.Birdsnest, log, store)
				if err != nil {
					t.Log(err)
					return nil
				}
				return txFilter
			}(),
			wantErr: false,
		},
		{
			name: "test3",
			args: args{
				conf: &filtercommon.TxFilterConfig{
					Type:              filtercommon.TxFilterTypeMap,
					BirdsNest:         defaultConf.Birdsnest,
					ShardingBirdsNest: defaultConf,
				},
				log:   log,
				store: store,
			},
			want: func() protocol.TxFilter {
				txFilter := mapimpl.New()
				return txFilter
			}(),
			wantErr: false,
		},
		{
			name: "test4",
			args: args{
				conf: &filtercommon.TxFilterConfig{
					Type:              filtercommon.TxFilterTypeShardingBirdsNest,
					BirdsNest:         defaultConf.Birdsnest,
					ShardingBirdsNest: defaultConf,
				},
				log:   log,
				store: store,
			},
			want: func() protocol.TxFilter {
				txFilter, err := shardingbirdsnest.New(defaultConf, log, store)
				if err != nil {
					t.Log(err)
					return nil
				}
				return txFilter
			}(),
			wantErr: false,
		},
		{
			name: "test5",
			args: args{
				conf: &filtercommon.TxFilterConfig{
					Type:              5,
					BirdsNest:         defaultConf.Birdsnest,
					ShardingBirdsNest: defaultConf,
				},
				log:   log,
				store: store,
			},
			want: func() protocol.TxFilter {
				return filterdefault.New(store)
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := &txFilterFactory{}
			got, err := cf.NewTxFilter(tt.args.conf, tt.args.log, tt.args.store)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTxFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.name == "test2" || tt.name == "test4" {
				if !reflect.DeepEqual(got.GetHeight(), tt.want.GetHeight()) {
					t.Errorf("NewTxFilter() got = %v, want %v", got, tt.want)
				}
			} else {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewTxFilter() got = %v, want %v", got, tt.want)
				}
			}

		})
	}
}

func newMockBlockchainStore(t *testing.T) *mock.MockBlockchainStore {
	ctrl := gomock.NewController(t)
	blockchainStore := mock.NewMockBlockchainStore(ctrl)
	return blockchainStore
}

func newMockLogger(t *testing.T) *mock.MockLogger {
	ctrl := gomock.NewController(t)
	logger := mock.NewMockLogger(ctrl)
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any()).AnyTimes()

	return logger
}

func createBlockByHash(height uint64, hash []byte) *commonpb.Block {
	//var hash = []byte("0123456789")
	var version = uint32(1)
	var block = &commonpb.Block{
		Header: &commonpb.BlockHeader{
			ChainId:        "Chain1",
			BlockHeight:    height,
			PreBlockHash:   hash,
			BlockHash:      hash,
			PreConfHeight:  0,
			BlockVersion:   version,
			DagHash:        hash,
			RwSetRoot:      hash,
			TxRoot:         hash,
			BlockTimestamp: 0,
			Proposer:       &accesscontrol.Member{MemberInfo: hash},
			ConsensusArgs:  nil,
			TxCount:        1,
			Signature:      []byte(""),
		},
		Dag: &commonpb.DAG{
			Vertexes: nil,
		},
		Txs: nil,
	}

	return block
}

func GetTestDefaultConfig(path string, i int) *sbn.ShardingBirdsNestConfig {
	return &sbn.ShardingBirdsNestConfig{
		Length:  10,
		Timeout: 10,
		ChainId: "chain1",
		Birdsnest: &bn.BirdsNestConfig{
			ChainId: "chain1",
			Length:  5,
			Rules: &bn.RulesConfig{
				AbsoluteExpireTime: 10000,
			},
			Cuckoo: &bn.CuckooConfig{
				KeyType:       bn.KeyType_KTDefault,
				TagsPerBucket: 4,
				BitsPerItem:   9,
				MaxNumKeys:    10,
				TableType:     1,
			},
			Snapshot: &bn.SnapshotSerializerConfig{
				Type:        bn.SerializeIntervalType_Timed,
				Timed:       &bn.TimedSerializeIntervalConfig{Interval: 20},
				BlockHeight: &bn.BlockHeightSerializeIntervalConfig{Interval: 20},
				Path:        path + strconv.Itoa(i),
			},
		},
		Snapshot: &bn.SnapshotSerializerConfig{
			Type:        bn.SerializeIntervalType_Timed,
			Timed:       &bn.TimedSerializeIntervalConfig{Interval: 20},
			BlockHeight: &bn.BlockHeightSerializeIntervalConfig{Interval: 20},
			Path:        path + strconv.Itoa(i),
		},
	}
}
