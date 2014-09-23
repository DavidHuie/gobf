# gobf - bloom filters in golang

## Creating a basic bloom filter

If you want to create an in-memory bloom filter with the standard
FNV hash function, call the function `NewDefault`:

```go
hashes := 5
size := 1000
b := gobf.NewDefault(hashes, size)
```

Here, `hashes` refers to the number of hash functions to use and `size` refers
to the size of the bit array.

## Inserting into a bloom filter

Use the `Insert` method to insert a key into a bloom filter:

```go
err := b.Insert([]byte("my key"))
```

## Checking for the presence of a key

Similarly, use the `Present` method to check if the bloom
filter contains a key:

```go
present, err := b.Present([]byte("my key"))
```

## Deleting a key

Use the `Delete` method:

```go
err := b.Delete([]byte("my key"))
```

Note that since this isn't a counting filter, a delete may affect other keys.

## Configurable bloom filters

The general constructor takes in these arguments:

```go
func New(db gobf.Db, hash hash.Hash64, hashes uint32, seed uint64, size uint64) (*BloomFilter, error)
```

The argument `hash` enables the developer to use a different hash function as
long as it implements the `hash.Hash64` interface from the standard library.

If you want to use a different data store, provide a struct instance that implements
the `gobf.Db` interface:

```go
type Db interface {
     Init(uint64) error
     GetBit(uint64) (bool, error)
     SetBit(uint64, bool) error
}
```

## Copyright

Copyright (c) 2013 David Huie. See LICENSE.txt for
further details.