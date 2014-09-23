package mem

import (
	"sync"
)

type Mem struct {
	sync.Mutex
	size uint64
	data []byte
}

func New() *Mem {
	return &Mem{}
}

func (m *Mem) Init(size uint64) error {
	m.Lock()
	defer m.Unlock()

	if len(m.data) > 0 {
		return nil
	}

	totalBytes := size / 8
	if size%8 != 0 {
		totalBytes += 1
	}

	m.size = size
	m.data = make([]byte, totalBytes)

	return nil
}

func (m *Mem) SetBit(n uint64, value bool) error {
	m.Lock()
	defer m.Unlock()

	r := m.size / n
	c := n % 8
	v := uint8(1) << c

	if value {
		m.data[r] = m.data[r] | byte(v)
	} else {
		m.data[r] = m.data[r] & ^byte(v)
	}

	return nil
}

func (m *Mem) GetBit(n uint64) (bool, error) {
	m.Lock()
	defer m.Unlock()

	r := m.size / n
	c := n % 8
	v := uint8(1) << c

	bit := (m.data[r] & v) >> c

	return uint8(bit) == 1, nil
}
