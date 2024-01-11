package cache

import "time"

type Cacher interface {
	Get(string) ([]byte, error)
	Set(string, []byte, time.Duration) error
	Has(string) bool
	Delete(string) error
	ResetCache() error
	IsFull() bool
	IsEmpty() bool
}
