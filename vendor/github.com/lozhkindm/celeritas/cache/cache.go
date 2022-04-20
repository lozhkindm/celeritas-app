package cache

import (
	"bytes"
	"encoding/gob"
)

type Entry map[string]interface{}

type Cache interface {
	Has(string) (bool, error)
	Get(string) (interface{}, error)
	Set(string, interface{}, ...int) error
	Forget(string) error
	Empty() error
	EmptyByMatch(string) error
}

func encode(entry Entry) ([]byte, error) {
	bb := bytes.Buffer{}
	e := gob.NewEncoder(&bb)
	if err := e.Encode(entry); err != nil {
		return nil, err
	}
	return bb.Bytes(), nil
}

func decode(b []byte) (Entry, error) {
	entry := Entry{}
	bb := bytes.Buffer{}
	bb.Write(b)
	d := gob.NewDecoder(&bb)
	if err := d.Decode(&entry); err != nil {
		return nil, err
	}
	return entry, nil
}
