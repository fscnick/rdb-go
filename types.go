package rdb

import "time"

type iterator interface {
	Next() (interface{}, error)
}

type collectionHead struct {
	DataKey DataKey
	Length  int64
}

type collectionEntry struct {
	DataKey DataKey
	Index   int64
	Length  int64
	Value   interface{}
}

type collectionSlice struct {
	DataKey DataKey
	Value   []interface{}
}

type collectionMapper interface {
	MapHead(*collectionHead) (interface{}, error)
	MapEntry(*collectionEntry) (interface{}, error)
	MapSlice(*collectionSlice) (interface{}, error)
}

type Aux struct {
	Key   string
	Value string
}

type DatabaseSize struct {
	Size   int64
	Expire int64
}

// DataKey contains the database, the key and the expiry of data.
type DataKey struct {
	Database int64
	Key      string
	Expiry   *time.Time
}

// StringData contains the key and the value of string data.
type StringData struct {
	DataKey
	Value string
}
