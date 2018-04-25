package iostdb

import "sync"

func CopyBytes(b []byte) (copiedBytes []byte) {
	if b == nil {
		return nil
	}
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)
	return
}

type MemDatabase struct {
	db   map[string][]byte
	lock sync.RWMutex
}

func NewMemDatabase() (*MemDatabase, error) {
	return &MemDatabase{db: make(map[string][]byte)}, nil
}

func NewMemDatabaseWithCap(size int) (*MemDatabase, error) {
	return &MemDatabase{db: make(map[string][]byte, size)}, nil
}

func (db *MemDatabase) Put(key []byte, value []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()
	db.db[string(key)] = CopyBytes(value)
	return nil
}
