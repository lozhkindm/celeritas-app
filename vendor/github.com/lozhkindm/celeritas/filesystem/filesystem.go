package filesystem

import "time"

type FileSystem interface {
	Put(filename, folder string) error
	Get(dst string, items ...string) error
	List(prefix string) ([]ListEntry, error)
	Delete(toDelete []string) (bool, error)
}

type ListEntry struct {
	Etag         string
	LastModified time.Time
	Key          string
	Size         float64
	IsDir        bool
}
