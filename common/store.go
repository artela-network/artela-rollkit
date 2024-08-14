package common

import "cosmossdk.io/core/store"

type prefixStore struct {
	prefix  []byte
	kvStore store.KVStore
}

func NewPrefixStore(kvStore store.KVStore, prefix []byte) store.KVStore {
	return &prefixStore{
		kvStore: kvStore,
		prefix:  prefix,
	}
}

func (p prefixStore) Get(key []byte) ([]byte, error) {
	return p.kvStore.Get(CloneAppend(p.prefix, key))
}

func (p prefixStore) Has(key []byte) (bool, error) {
	return p.kvStore.Has(CloneAppend(p.prefix, key))
}

func (p prefixStore) Set(key, value []byte) error {
	return p.kvStore.Set(CloneAppend(p.prefix, key), value)
}

func (p prefixStore) Delete(key []byte) error {
	return p.kvStore.Delete(CloneAppend(p.prefix, key))
}

func (p prefixStore) Iterator(start, end []byte) (store.Iterator, error) {
	return p.kvStore.Iterator(CloneAppend(p.prefix, start), CloneAppend(p.prefix, end))
}

func (p prefixStore) ReverseIterator(start, end []byte) (store.Iterator, error) {
	return p.kvStore.ReverseIterator(CloneAppend(p.prefix, start), CloneAppend(p.prefix, end))
}
