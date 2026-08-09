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

type LSMStorage struct {
	data     map[string][]byte
	maxlen   int
}

func NewLSMStore(maxLen int) *LSMStorage {
	store := &LSMStorage{data: make(map[string][]byte, maxLen), maxlen: maxLen}
	err := LSMWalRead(store)
	if err != nil {
		fmt.Println("Error creating wal storage")
		return nil
	}
	return store
}

func (m *LSMStorage) Set(key string, value []byte) error {	
	if len(m.data) == m.maxlen {
		sortKeys := make([]string, 0, len(m.data))
		for key := range m.data {
			sortKeys = append(sortKeys, key)
		}
		sort.Strings(sortKeys)
		
		if err := os.MkdirAll("../../internal/storage/data", 0755); err != nil {
			return err
		}
		sstableName := fmt.Sprintf("../../internal/storage/data/sst-%s-%d.db", sortKeys[0], time.Now().Unix())
		sstable, err := os.OpenFile(sstableName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
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

		entries, err := os.ReadDir("../../internal/storage/data")
		groups := make(map[string][]os.DirEntry)
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			parts := strings.SplitN(entry.Name(), "-", 3)
			if len(parts) < 3 {
				continue
			}
			groups[parts[1]] = append(groups[parts[1]], entry)
		}
			DuplicateIndexKeyCompaction(groups)
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

func (m *LSMStorage) Get(key string) ([]byte, error) {
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
			return strings.Compare(sstables[i].Name(), sstables[j].Name()) < 0
		})

		for len(sstables) >= 1 {
			lastSstableIndex := strings.SplitN(sstables[len(sstables)-1].Name(), "-", 3)[1]
			midSstableIndex := strings.SplitN(sstables[len(sstables)/2].Name(), "-", 3)[1]
			firstSstableIndex := strings.SplitN(sstables[0].Name(), "-", 3)[1]
			if key >= lastSstableIndex {
				value, err := SSTableSearch(sstables[len(sstables)-1], key)
				if err != nil {
					return nil, err
				}
				if value != nil {
					return value, nil
				}
			}
			if key >= midSstableIndex {
				value, err := SSTableSearch(sstables[len(sstables)/2], key)
				if err != nil {
					return nil, err
				}
				if value != nil {
					m.data[key] = value
				}
				sstables = sstables[len(sstables)/2+1:len(sstables)-2]
			} else if key >= firstSstableIndex {
				value, err := SSTableSearch(sstables[0], key)
				if err != nil {
					return nil, err
				}
				if value != nil {
					m.data[key] = value
				}
				sstables = sstables[1: len(sstables)/2]
			}
		}
		value, ok = m.data[key]
		if !ok {
			return nil, errors.New("key not found")
		}
	}
	return value, nil
}

func (m *LSMStorage) Delete(key string) error {	
	if len(m.data) == m.maxlen {
		sortKeys := make([]string, 0, len(m.data))
		for key := range m.data {
			sortKeys = append(sortKeys, key)
		}
		sort.Strings(sortKeys)
		
		sstable, err := os.OpenFile(fmt.Sprintf("../../internal/storage/data/sst-%s-%d.db", sortKeys[0], time.Now().Unix()), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
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

func (m *LSMStorage) Exists(key string) bool {
	_, err := m.Get(key)
	return err != nil
}

func LSMWalRead(store Storage) error {
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

func SSTableSearch(sstable os.DirEntry, key string) ([]byte, error) {
	file, err := os.Open("../../internal/storage/data/" + sstable.Name())
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.SplitN(line, "->", 2)
		if len(tokens) == 2 && strings.TrimSpace(key) == strings.TrimSpace(tokens[0]) && strings.TrimSpace(tokens[1]) == "TOMBSTONE" {
			return nil, nil
		} else if len(tokens) == 2 && strings.TrimSpace(key) == strings.TrimSpace(tokens[0]) {
			return []byte(strings.TrimSpace(tokens[1])), nil
		}
	}
	
	if err:= scanner.Err(); err != nil {
		return nil, errors.New("Error reading wal")
	}

	return nil, nil
}

func BatchCompaction(sstables []os.DirEntry, batchSize int) error{
	sort.Slice(sstables, func (i, j int) bool {
		return strings.Compare(sstables[i].Name(), sstables[j].Name()) > 0
	})
	merged := make(map[string][]byte)
	for _, sstable := range sstables {
		file, err := os.Open("../../internal/storage/data/" + sstable.Name())
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.SplitN(line, "->", 2)
			if len(parts) != 2 {
				continue
			}
			_, ok := merged[parts[0]]
			if !ok {
				merged[parts[0]] = []byte(parts[1])
			}
		}

		if err := scanner.Err(); err != nil {
			file.Close()
			return err
		}

		file.Close()
		os.Remove(file.Name())
	}

	var sortedKeys []string
	for key := range merged {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Slice(sortedKeys, func (i, j int) bool {
		return strings.Compare(sortedKeys[i], sortedKeys[j]) < 0
	})

	newSstable, err := os.OpenFile("../../internal/storage/data/" + sstables[0].Name(), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return errors.New("Error reading sstable") 
	}
	for _, key := range sortedKeys {
		_, err = fmt.Fprintf(newSstable, "%s -> %s\n", key, merged[key])
		if err != nil {
			return err
		}
	}
	defer newSstable.Close()
	return nil
}

func DuplicateIndexKeyCompaction(groups map[string][]os.DirEntry) error {
	for _, group := range groups {
		if len(group) > 1 {
			sort.Slice(group, func (i, j int) bool {
				return strings.Compare(group[i].Name(), group[j].Name()) > 0
			})
			merged := make(map[string][]byte)
			for _, sstable := range group {
				file, err := os.Open("../../internal/storage/data/" + sstable.Name())
				if err != nil {
					return err
				}

				scanner := bufio.NewScanner(file)

				for scanner.Scan() {
					line := scanner.Text()
					parts := strings.SplitN(line, "->", 2)
					if len(parts) != 2 {
						continue
					}
					_, ok := merged[parts[0]]
					if !ok {
						merged[parts[0]] = []byte(parts[1])
					}
				}

				if err := scanner.Err(); err != nil {
					file.Close()
					return err
				}

				file.Close()
				os.Remove(file.Name())
			}

			var sortedKeys []string
			for key := range merged {
				sortedKeys = append(sortedKeys, key)
			}
			sort.Slice(sortedKeys, func (i, j int) bool {
				return strings.Compare(sortedKeys[i], sortedKeys[j]) < 0
			})

			newSstable, err := os.OpenFile("../../internal/storage/data/" + group[0].Name(), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
			if err != nil {
				return errors.New("Error reading sstable") 
			}
			for _, key := range sortedKeys {
				_, err = fmt.Fprintf(newSstable, "%s -> %s\n", key, merged[key])
				if err != nil {
					return err
				}
			}
			defer newSstable.Close()
		}
	}

	return nil
}

