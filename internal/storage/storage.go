package storage

type Storage interface {
	Set(key string, value []byte) error

	Get(key string) ([]byte, error)

	Delete(key string) error

	Exists(key string) bool
}