package standalone_storage

import (
	
	"path/filepath"

	"github.com/Connor1996/badger"
	"github.com/pingcap-incubator/tinykv/kv/config"
	"github.com/pingcap-incubator/tinykv/kv/storage"
	"github.com/pingcap-incubator/tinykv/kv/util/engine_util"
	"github.com/pingcap-incubator/tinykv/proto/pkg/kvrpcpb"

)
//"go/types"
//	"github.com/pingcap/tidb/util/ranger"
// StandAloneStorage is an implementation of `Storage` for a single-node TinyKV instance. It does not
// communicate with other nodes and all data is stored locally.
type StandAloneStorage struct {
	// Your Data Here (1).
	eng *engine_util.Engines
	con *config.Config
	Txn *badger.Txn
}

func NewStandAloneStorage(conf *config.Config) *StandAloneStorage {
	// Your Code Here (1).
	kvPath := filepath.Join(conf.DBPath, "kv")
	raftPath := filepath.Join(conf.DBPath, "raft")

	raftDB := engine_util.CreateDB(raftPath, true)
	kvDB := engine_util.CreateDB(kvPath, false)         //open（）

	return &StandAloneStorage{
		eng : engine_util.NewEngines(kvDB, raftDB, kvPath, raftPath),
		con : conf,
	}
}

func (s *StandAloneStorage) Start() error {
	// Your Code Here (1).
	return nil
}

func (s *StandAloneStorage) Stop() error {
	// Your Code Here (1).
	return s.eng.Close()
}
/*
type Storage_Reader struct{
	Txn *badger.Txn
}

func (s *StandAloneStorage) Reader(ctx *kvrpcpb.Context) (storage.StorageReader, error) {
	// Your Code Here (1).

	txn := s.eng.Kv.NewTransaction(false)
	return &Storage_Reader{
		Txn : txn,
	}, nil
}

func (s *Storage_Reader) GetCF(cf string, key []byte) ([]byte, error){
	return engine_util.GetCFFromTxn(s.Txn, cf, key)
}

func (s *Storage_Reader) IterCF(cf string) engine_util.DBIterator{
	return engine_util.NewCFIterator(cf, s.Txn)
}

func (s *Storage_Reader) Close(){
	s.Txn.Discard()
}

func (s *StandAloneStorage) Write(ctx *kvrpcpb.Context, batch []storage.Modify) error {
	// Your Code Here (1).
	for _, p := range batch{
		switch p.Data.(type){
		case storage.Put:
			err := engine_util.PutCF( s.eng.Kv, p.Cf(), p.Key(), p.Value())
			if err != nil{
				return err
			}
		case storage.Delete:
			err := engine_util.DeleteCF(s.eng.Kv, p.Cf(), p.Key())
			if err != nil{
				return err
			}
		}
	}

	return nil
}
*/
func (s *StandAloneStorage) Reader(ctx *kvrpcpb.Context) (storage.StorageReader, error) {
	// Your Code Here (1).

	txn := s.eng.Kv.NewTransaction(false)
	return &StandAloneStorage{
		Txn : txn,
	}, nil
}

func (s *StandAloneStorage) GetCF(cf string, key []byte) ([]byte, error){
	iter,_ := engine_util.GetCFFromTxn(s.Txn, cf, key)

	return iter,nil
}

func (s *StandAloneStorage) IterCF(cf string) engine_util.DBIterator{
	return engine_util.NewCFIterator(cf, s.Txn)
}

func (s *StandAloneStorage) Close(){
	s.Txn.Discard()
}

func (s *StandAloneStorage) Write(ctx *kvrpcpb.Context, batch []storage.Modify) error {
	// Your Code Here (1).
	for _, p := range batch{
		switch p.Data.(type){
		case storage.Put:
			err := engine_util.PutCF( s.eng.Kv, p.Cf(), p.Key(), p.Value())
			if err != nil{
				return err
			}
		case storage.Delete:
			err := engine_util.DeleteCF(s.eng.Kv, p.Cf(), p.Key())
			if err != nil{
				return err
			}
		}
	}

	return nil
}