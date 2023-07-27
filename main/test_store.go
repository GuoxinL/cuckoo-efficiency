/*
   Created by guoxin in 2023/7/26 17:22
*/
package main

import (
	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/pb-go/v2/store"
	"chainmaker.org/chainmaker/protocol/v2"
)

type TestStore struct {
}

func (t TestStore) QuerySingle(contractName, sql string, values ...interface{}) (protocol.SqlRow, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) QueryMulti(contractName, sql string, values ...interface{}) (protocol.SqlRows, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) ExecDdlSql(contractName, sql, version string) error {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) BeginDbTransaction(txName string) (protocol.SqlDBTransaction, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetDbTransaction(txName string) (protocol.SqlDBTransaction, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) CommitDbTransaction(txName string) error {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) RollbackDbTransaction(txName string) error {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) CreateDatabase(contractName string) error {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) DropDatabase(contractName string) error {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetContractDbName(contractName string) string {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetContractByName(name string) (*common.Contract, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetContractBytecode(name string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetMemberExtraData(member *accesscontrol.Member) (*accesscontrol.MemberExtraData, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) InitGenesis(genesisBlock *store.BlockWithRWSet) error {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) PutBlock(block *common.Block, txRWSets []*common.TxRWSet) error {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetBlockByHash(blockHash []byte) (*common.Block, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) BlockExists(blockHash []byte) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetHeightByHash(blockHash []byte) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetBlockHeaderByHeight(height uint64) (*common.BlockHeader, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetBlock(height uint64) (*common.Block, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetLastConfigBlock() (*common.Block, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetLastChainConfig() (*config.ChainConfig, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetBlockByTx(txId string) (*common.Block, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetBlockWithRWSets(height uint64) (*store.BlockWithRWSet, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetTx(txId string) (*common.Transaction, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetTxWithRWSet(txId string) (*common.TransactionWithRWSet, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetTxInfoWithRWSet(txId string) (*common.TransactionInfoWithRWSet, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetTxWithInfo(txId string) (*common.TransactionInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) TxExists(txId string) (bool, error) {
	return true, nil
}

func (t TestStore) TxExistsInFullDB(txId string) (bool, uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) TxExistsInIncrementDB(txId string, startHeight uint64) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) TxExistsInIncrementDBState(txId string, startHeight uint64) (bool, bool, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetTxInfoOnly(txId string) (*common.TransactionInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetTxHeight(txId string) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetTxConfirmedTime(txId string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetLastBlock() (*common.Block, error) {
	return &common.Block{
		Header: &common.BlockHeader{
			BlockHeight: 0,
		},
	}, nil
}

func (t TestStore) ReadObject(contractName string, key []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) ReadObjects(contractName string, keys [][]byte) ([][]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) SelectObject(contractName string, startKey []byte, limit []byte) (protocol.StateIterator, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetTxRWSet(txId string) (*common.TxRWSet, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetTxRWSetsByHeight(height uint64) ([]*common.TxRWSet, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetDBHandle(dbName string) protocol.DBHandle {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetArchivedPivot() uint64 {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) ArchiveBlock(archiveHeight uint64) error {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) RestoreBlocks(serializedBlocks [][]byte) error {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) Close() error {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetHistoryForKey(contractName string, key []byte) (protocol.KeyHistoryIterator, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetAccountTxHistory(accountId []byte) (protocol.TxHistoryIterator, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) GetContractTxHistory(contractName string) (protocol.TxHistoryIterator, error) {
	//TODO implement me
	panic("implement me")
}
