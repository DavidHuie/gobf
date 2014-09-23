package gobf

import (
	"testing"
)

type payloadToHash struct {
	payload []byte
	hash    uint64
}

func TestHashPayload(t *testing.T) {
	b := bf()

	examples := []payloadToHash{
		{[]byte("a"), 90543005},
		{[]byte("david"), 2902346954},
		{[]byte("this is a bloom filter"), 1446046864},
		{[]byte("asdf"), 281143284},
	}

	for _, p := range examples {
		if h := b.hashPayload(p.payload, 3); h != p.hash {
			t.Errorf("expected %d, got %d", p.hash, h)
		}
	}

}

func TestBf(t *testing.T) {
	b := bf()

	examples := [][]byte{
		[]byte("david huie"),
		[]byte("welcome to gobf!"),
	}

	for _, e := range examples {
		if err := b.Insert(e); err != nil {
			t.Error(err)
		}
		val, err := b.Present(e)
		if err != nil {
			t.Error(err)
		}
		if !val {
			t.Errorf("expected %v to be present", e)
		}
		if err := b.Delete(e); err != nil {
			t.Error(err)
		}
		val, err = b.Present(e)
		if err != nil {
			t.Error(err)
		}
		if val {
			t.Errorf("expected %v to not be present", e)
		}
	}
}

func bf() *BloomFilter {
	bf, err := NewDefault(5, 3000000000)
	if err != nil {
		panic(err)
	}
	return bf
}
