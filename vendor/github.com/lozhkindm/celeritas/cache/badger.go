package cache

import (
	"time"

	"github.com/dgraph-io/badger/v3"
)

type BadgerCache struct {
	Conn   *badger.DB
	Prefix string
}

func (bc *BadgerCache) Has(key string) (bool, error) {
	if _, err := bc.Get(key); err != nil {
		return false, nil
	}
	return true, nil
}

func (bc *BadgerCache) Get(key string) (interface{}, error) {
	var value []byte

	err := bc.Conn.View(func(txn *badger.Txn) error {
		ent, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		err = ent.Value(func(val []byte) error {
			value = append(value, val...)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	entry, err := decode(value)
	if err != nil {
		return nil, err
	}

	return entry[key], nil
}

func (bc *BadgerCache) Set(key string, val interface{}, expires ...int) error {
	entry := Entry{}
	entry[key] = val

	value, err := encode(entry)
	if err != nil {
		return err
	}

	ent := badger.NewEntry([]byte(key), value)
	if len(expires) > 0 {
		ent.WithTTL(time.Second * time.Duration(expires[0]))
	}

	return bc.Conn.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(ent)
	})
}

func (bc *BadgerCache) Forget(key string) error {
	return bc.Conn.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func (bc *BadgerCache) Empty() error {
	return bc.emptyByMatch("")
}

func (bc *BadgerCache) EmptyByMatch(pattern string) error {
	return bc.emptyByMatch(pattern)
}

func (bc *BadgerCache) emptyByMatch(pattern string) error {
	deleteKeys := func(keysToDelete [][]byte) error {
		return bc.Conn.Update(func(txn *badger.Txn) error {
			for _, key := range keysToDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		})
	}

	collectSize := 100_000

	return bc.Conn.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.AllVersions = false
		opts.PrefetchValues = false
		iter := txn.NewIterator(opts)
		defer iter.Close()

		keysToDelete := make([][]byte, 0, collectSize)
		keysCollected := 0

		for iter.Seek([]byte(pattern)); iter.ValidForPrefix([]byte(pattern)); iter.Next() {
			key := iter.Item().KeyCopy(nil)
			keysToDelete = append(keysToDelete, key)
			keysCollected++
			if keysCollected == collectSize {
				if err := deleteKeys(keysToDelete); err != nil {
					return err
				}
				keysToDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}
		if keysCollected > 0 {
			if err := deleteKeys(keysToDelete); err != nil {
				return err
			}
		}
		return nil
	})
}
