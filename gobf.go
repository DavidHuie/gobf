package gobf

import (
	"encoding/binary"
	"hash"
	"sync"
)

type Db interface {
	Init(uint64) error
	SetBit(uint64, bool) error
	GetBit(uint64) (bool, error)
}

type BloomFilter struct {
	hash     hash.Hash64
	db       Db
	size     uint64
	hashes   uint32
	seed     []byte
	hashLock *sync.Mutex
}

func NewBloomFilter(hash hash.Hash64, db Db, size uint64, hashes uint32, seed uint64) (*BloomFilter, error) {
	if err := db.Init(size); err != nil {
		return nil, err
	}

	seedBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(seedBytes, seed)

	return &BloomFilter{
		hash:     hash,
		db:       db,
		size:     size,
		hashes:   hashes,
		seed:     seedBytes,
		hashLock: &sync.Mutex{},
	}, nil
}

func (bf *BloomFilter) hashBytes(b []byte) uint64 {
	bf.hashLock.Lock()
	defer bf.hashLock.Unlock()

	bf.hash.Write(b)
	return bf.hash.Sum64()
}

func (bf *BloomFilter) hashPayload(p []byte, num uint32) uint64 {
	numBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(numBytes, num)

	fullPayload := append(p, bf.seed...)
	fullPayload = append(fullPayload, numBytes...)

	return bf.hashBytes(fullPayload)
}

func (bf *BloomFilter) setKeyToBool(key []byte, b bool) error {
	for i := uint32(0); i < bf.hashes; i++ {
		h := bf.hashPayload(key, i)
		if err := bf.db.SetBit(h, b); err != nil {
			return err
		}
	}
	return nil
}

func (bf *BloomFilter) Insert(p []byte) error {
	return bf.setKeyToBool(p, true)
}

func (bf *BloomFilter) Delete(p []byte) error {
	return bf.setKeyToBool(p, false)
}

func (bf *BloomFilter) Present(p []byte) (bool, error) {
	for i := uint32(0); i < bf.hashes; i++ {
		h := bf.hashPayload(p, i)
		val, err := bf.db.GetBit(h)
		if err != nil {
			return false, err
		}
		if !val {
			return false, nil
		}
	}
	return true, nil
}
