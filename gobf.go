package gobf

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"sync"

	"github.com/DavidHuie/gobf/db/mem"
)

type Db interface {
	// Initializes the data store to contain the capacity specified by
	// the input uint
	Init(uint64) error

	GetBit(uint64) (bool, error)
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

// Returns a completely customizable bloom filter
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

// Returns an in-memory bloom filter that builds on the FNV hash
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

// Returns the a hash of the input bytes
func (bf *BloomFilter) hashBytes(b []byte) uint64 {
	bf.hashLock.Lock()
	defer bf.hashLock.Unlock()

	bf.hash.Write(b)
	defer bf.hash.Reset()

	return bf.hash.Sum64()
}

// Hashes the input bytes taking into account the seed and the hash number
func (bf *BloomFilter) hashPayload(p []byte, num uint32) uint64 {
	numBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(numBytes, num)

	fullPayload := append(p, bf.seed...)
	fullPayload = append(fullPayload, numBytes...)

	return bf.hashBytes(fullPayload) % bf.size
}

// Sets a key to either true || false in the bloom filter
func (bf *BloomFilter) setKeyToBool(key []byte, b bool) error {
	for i := uint32(0); i < bf.hashes; i++ {
		h := bf.hashPayload(key, i)
		if err := bf.db.SetBit(h, b); err != nil {
			return err
		}
	}
	return nil
}

// Inserts a key into the bloom filter
func (bf *BloomFilter) Insert(key []byte) error {
	return bf.setKeyToBool(key, true)
}

// Deletes a key from the bloom filter
func (bf *BloomFilter) Delete(key []byte) error {
	return bf.setKeyToBool(key, false)
}

// Returns true if the key is present in the bloom filter
func (bf *BloomFilter) Present(key []byte) (bool, error) {
	for i := uint32(0); i < bf.hashes; i++ {
		h := bf.hashPayload(key, i)
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
