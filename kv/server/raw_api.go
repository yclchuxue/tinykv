package server

import (
	"context"
	"github.com/pingcap-incubator/tinykv/proto/pkg/kvrpcpb"
	"github.com/pingcap-incubator/tinykv/kv/storage"
)

// The functions below are Server's Raw API. (implements TinyKvServer).
// Some helper methods can be found in sever.go in the current directory

// RawGet return the corresponding Get response based on RawGetRequest's CF and Key fields
func (server *Server) RawGet(_ context.Context, req *kvrpcpb.RawGetRequest) (*kvrpcpb.RawGetResponse, error) {
	// Your Code Here (1).
	s,err := server.storage.Reader(nil)

	val,err := s.GetCF(req.Cf, req.Key)

	T := false 

	if err != nil{
		return &kvrpcpb.RawGetResponse{
			NotFound: true,
		},nil
	}

	if err == nil{
		T = true
	}

	return &kvrpcpb.RawGetResponse{
		Value: val,
		NotFound: T,
	}, nil
}

// RawPut puts the target data into storage and returns the corresponding response
func (server *Server) RawPut(_ context.Context, req *kvrpcpb.RawPutRequest) (*kvrpcpb.RawPutResponse, error) {
	// Your Code Here (1).
	// Hint: Consider using Storage.Modify to store data to be modified

	err := server.storage.Write(nil, []storage.Modify{
		{
			Data : storage.Put{
				Cf:    req.Cf,
				Key:   req.Key,
				Value: req.Value,
			},
		},
	})

	return nil, err
}

// RawDelete delete the target data from storage and returns the corresponding response
func (server *Server) RawDelete(_ context.Context, req *kvrpcpb.RawDeleteRequest) (*kvrpcpb.RawDeleteResponse, error) {
	// Your Code Here (1).
	// Hint: Consider using Storage.Modify to store data to be deleted
	err := server.storage.Write(nil, []storage.Modify{
		{
			Data : storage.Delete{
				Cf:    req.Cf,
				Key:   req.Key,
			},
		},
	})

	return nil, err
}

// RawScan scan the data starting from the start key up to limit. and return the corresponding result
func (server *Server) RawScan(_ context.Context, req *kvrpcpb.RawScanRequest) (*kvrpcpb.RawScanResponse, error) {
	// Your Code Here (1).
	// Hint: Consider using reader.IterCF
	s, err := server.storage.Reader(req.Context)
	if err != nil{
		return &kvrpcpb.RawScanResponse{},err
	}
	limit := req.Limit
	var array []*kvrpcpb.KvPair
	iter := s.IterCF(req.Cf)
	iter.Seek(req.StartKey)
	for ;iter.Valid();iter.Next(){
		item := iter.Item()
		val,_ := item.Value()
		ke := item.Key()
		/*
		pair :=  &kvrpcpb.KvPair{
			Key:  ke,
			Value : val,
		}
		*/
		pair := new(kvrpcpb.KvPair)
		pair.Key = ke
		pair.Value = val
		array = append(array, pair)

		limit--

		if limit == 0{
			break
		}
	}

	iter.Close()
	return &kvrpcpb.RawScanResponse{
		Kvs: array,
	}, nil
}
