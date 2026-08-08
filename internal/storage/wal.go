package storage

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

type WalStorage struct {
	data map[string][]byte
}

func NewWalStore() *WalStorage {
	store := &WalStorage{ data: make(map[string][]byte) }
	err := WalRead(store)
	if err != nil {
		fmt.Println("Error creating wal storage")
		return nil
	}
	return store
}

func (m *WalStorage) Set(key string, value []byte) error {
	file, err := os.OpenFile("../../internal/storage/wal.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error reading wal")
		return err
	}
	_, err = fmt.Fprintf(file, "SET %s %s\n",key , value)
	if err != nil {
		return err
	}
	m.data[key] = value
	defer file.Close()
	return nil
}

func (m *WalStorage) Get(key string) ([]byte, error) {
	value, ok := m.data[key]
	if !ok {
		return nil, errors.New("key not found")
	}
	copyValue := make([]byte, len(value))
	copy(copyValue, value)
	return copyValue, nil
}

func (m *WalStorage) Delete(key string) error {
	file, err := os.OpenFile("../../internal/storage/wal.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error reading wal")
		return err
	}
	_, err = fmt.Fprintf(file, "DELETE %s\n", key)
	if err != nil {
		return err
	}
	delete(m.data, key)
	return nil
}

func (m *WalStorage) Exists(key string) bool {
	_, ok := m.data[key]
	return ok
}

func (m *WalStorage) Keys() []string {
	keys := make([]string, len(m.data))
	for key := range m.data {
		keys = append(keys, key)
	}
	return keys
}

func WalRead(store Storage) error {
	file, err := os.OpenFile("../../internal/storage/wal.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	
	if err != nil {
		return errors.New("Error reading wal") 
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	
	walStore, ok := store.(*WalStorage)
	if !ok {
		return errors.New("Storage not of type walstorage")
	}

	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Fields(line)
		if len(tokens) >= 1 {
			switch strings.ToUpper(tokens[0]) {
				case "SET": if (len(tokens) == 3) {
					walStore.data[tokens[1]] = []byte(tokens[2])
				}
				case "DELETE": delete(walStore.data, tokens[0])
			}
		}
	}
	
	if err:= scanner.Err(); err != nil {
		return errors.New("Error reading wal")
	}
	return nil
}