package gobf

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"sync"

	"github.com/DavidHuie/gobf/db/mem"
)

type Db interface {
	GetBit(uint64) (bool, error)
	Init(uint64) error
	SetBit(uint64, bool) error
}

type BloomFilter struct {
	db       Db
	hash     hash.Hash64
	hashLock *sync.Mutex
	hashes   uint32
	seed     []byte
	size     uint64
}

func New(db Db, hash hash.Hash64, hashes uint32, seed uint64, size uint64) (*BloomFilter, error) {
	if err := db.Init(size); err != nil {
		return nil, err
	}

	seedBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(seedBytes, seed)

	return &BloomFilter{
		db:       db,
		hash:     hash,
		hashLock: &sync.Mutex{},
		hashes:   hashes,
		seed:     seedBytes,
		size:     size,
	}, nil
}

const (
	defaultSeed = 1337
)

func NewDefault(hashes uint32, size uint64) (*BloomFilter, error) {
	seedBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(seedBytes, defaultSeed)

	db := mem.New()
	if err := db.Init(size); err != nil {
		return nil, err
	}

	return &BloomFilter{
		db:       db,
		hash:     fnv.New64(),
		hashLock: &sync.Mutex{},
		hashes:   hashes,
		seed:     seedBytes,
		size:     size,
	}, nil
}

func (bf *BloomFilter) hashBytes(b []byte) uint64 {
	bf.hashLock.Lock()
	defer bf.hashLock.Unlock()

	bf.hash.Write(b)
	defer bf.hash.Reset()

	return bf.hash.Sum64()
}

func (bf *BloomFilter) hashPayload(p []byte, num uint32) uint64 {
	numBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(numBytes, num)

	fullPayload := append(p, bf.seed...)
	fullPayload = append(fullPayload, numBytes...)

	return bf.hashBytes(fullPayload) % bf.size
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
