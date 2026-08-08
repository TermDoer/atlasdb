package storage

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

type SSTableStorage struct {
	data     map[string][]byte
	maxlen int
}

func NewSSTableStore(maxBytes int) *SSTableStorage {
	store := &SSTableStorage{data: make(map[string][]byte, maxBytes), maxlen: maxBytes}
	err := SSTableWalRead(store)
	if err != nil {
		fmt.Println("Error creating wal storage")
		return nil
	}
	return store
}

func (m *SSTableStorage) Set(key string, value []byte) error {	
	if len(m.data) == m.maxlen {
		sortKeys := make([]string, 0, len(m.data))
		for key := range m.data {
			sortKeys = append(sortKeys, key)
		}
		sort.Strings(sortKeys)
		
		if err := os.MkdirAll("../../internal/storage/data", 0755); err != nil {
			return err
		}
		sstable, err := os.OpenFile(fmt.Sprintf("../../internal/storage/data/sst-%d.db", time.Now().Unix()), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			if os.IsExist(err) {
				fmt.Println("Sstable immutable")
				} else {
					fmt.Println("failure creating sstable")
				}
				return err
			}
		
			for _, key := range sortKeys {
				_, err = fmt.Fprintf(sstable, "%s -> %s\n", key, m.data[key])
				if err != nil {
					fmt.Println("failure creating sstable")
					defer os.Remove(sstable.Name())
					return err
				}
		}
		clear(m.data)
		sstable.Close()
	}
	walfile, err := os.OpenFile("../../internal/storage/wal.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return errors.New("Error reading wal")
	}
	_, err = fmt.Fprintf(walfile, "SET %s %s\n", key, value)
	if err != nil {
		return err
	}
	m.data[key] = value
	walfile.Close()
	return nil
}

func (m *SSTableStorage) Get(key string) ([]byte, error) {
	value, ok := m.data[key]
	if string(value) == "TOMBSTONE" {
		return nil, errors.New("key not found")
	}
	if !ok {
		entries, err := os.ReadDir("../../internal/storage/data")
		if err != nil {
			return nil, errors.New("key not found")
		}
		var sstables []os.DirEntry
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			if !strings.HasPrefix(entry.Name(), "sst-") ||
			!strings.HasSuffix(entry.Name(), ".db") {
				continue
			}
			sstables = append(sstables, entry)
		}
		sort.Slice(sstables, func (i, j int) bool {
			return strings.Compare(sstables[i].Name(), sstables[j].Name()) > 0
		})

		for _, sstable := range sstables {
			file, err := os.Open("../../internal/storage/data/" + sstable.Name())
			if err != nil {
				return nil, err
			}
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				tokens := strings.SplitN(line, "->", 2)
				if len(tokens) == 2 && strings.TrimSpace(key) == strings.TrimSpace(tokens[0]) && strings.TrimSpace(tokens[1]) == "TOMBSTONE" {
					return nil, errors.New("key not found")
				} else if len(tokens) == 2 && strings.TrimSpace(key) == strings.TrimSpace(tokens[0]) {
					return []byte(strings.TrimSpace(tokens[1])), nil
				}
			}
			
			if err:= scanner.Err(); err != nil {
				return nil, errors.New("Error reading wal")
			}
		}
		return nil, errors.New("key not found")
	}
	return value, nil
}

func (m *SSTableStorage) Delete(key string) error {	
	if len(m.data) == m.maxlen {
		sortKeys := make([]string, 0, len(m.data))
		for key := range m.data {
			sortKeys = append(sortKeys, key)
		}
		sort.Strings(sortKeys)
		
		sstable, err := os.OpenFile(fmt.Sprintf("../../internal/storage/data/sst-%d.db", time.Now().Unix()), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			if os.IsExist(err) {
				fmt.Println("Sstable immutable")
			} else {
				fmt.Println("failure creating sstable")
			}
			return err
		}
		
		for _, key := range sortKeys {
			_, err = fmt.Fprintf(sstable, "%s -> TOMBSTONE\n", key)
			if err != nil {
				fmt.Println("failure creating sstable")
				defer os.Remove(sstable.Name())
				return err
			}
		}
		clear(m.data)
		sstable.Close()
	}
	walfile, err := os.OpenFile("../../internal/storage/wal.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return errors.New("Error reading wal")
	}
	_, err = fmt.Fprintf(walfile, "DELETE %s\n", key)
	if err != nil {
		return err
	}
	m.data[key] = []byte("TOMBSTONE")
	walfile.Close()
	return nil
}

func (m *SSTableStorage) Exists(key string) bool {
	_, err := m.Get(key)
	return err != nil
}

func SSTableWalRead(store Storage) error {
	file, err := os.OpenFile("../../internal/storage/wal.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	
	if err != nil {
		return errors.New("Error reading wal") 
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	
	sstableWalStore, ok := store.(*SSTableStorage)
	if !ok {
		return errors.New("Storage not of type walstorage")
	}

	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Fields(line)
		if len(tokens) >= 1 {
			switch strings.ToUpper(tokens[0]) {
				case "SET": if (len(tokens) == 3) {
					sstableWalStore.data[tokens[1]] = []byte(tokens[2])
				}
				case "DELETE": delete(sstableWalStore.data, tokens[0])
			}
		}
	}
	
	if err:= scanner.Err(); err != nil {
		return errors.New("Error reading wal")
	}
	return nil
}