# Implementation
## 1. Go map (In-Memory datastore)
### Architecture
```
      Client
        |
        V
     atlasctl
        │
      atlasd
        │
        V
----------------
|    Go map     |
-----------------
```
### Drawbacks
- No failure or memory recovery
```
$ atlasctl
SET name John
OK

Ctrl+C

$ atlasctl
GET name
key not found
```

## 2. WAL (Write Ahead Log)
### Architecture
```
        Client
           |
           V
        atlasctl
           |
         atlasd
           |
           V
      ------------
      | MemTable |
      ------------
           |
           V
     Write Ahead Log
           |
     append-only file
```
### Features
- Failure recovery (Durability)
- Sequential writes
### Drawbacks
- SSD memory delete, or file delete

## 3. SSTables
### Architecture
```
      Client
        |
        V
     atlasctl--------|
        |            |----WAL
      atlasd---------|
        |
        V
    -----------
    | MemTable |
     -----------
         |
      (Flush)
   SSTable-001.db
   SSTable-002.db
```
### Features
- Failure recovery
- Data overflow stored in sstables
- Sequential writes
### Drawbacks
- SSD memory delete, or file delete
- Read overhead (Linear Read)

## 4. LSM Tree
### Architecture
```
      Client
        |
        V
     atlasctl--------|
        |            |----WAL
      atlasd---------|
        |
        V
    -----------
    | MemTable |
     -----------
         |
      (Flush)
   SSTable-001.db
   SSTable-002.db
```
### Features
- Failure recovery
- Binary Search + Sparse Indexing + Bloom filters
- SStable Compaction (Junk memory removed)
- Sequential writes
### Drawbacks
- SSD memory delete, or file delete

### Compaction 
- Batch compaction of sstables & Duplicate sparse index sstable compaction
- Each sstable has key values sorted increasing order
- the sstables are sorted based on increasing order of sstable file name (sparse index + file creation timestamp)
- latest value only is considered

### How binary search works ?
- Each sstable has key values sorted increasing order
- The smallest key acts as the sparse index and is in the sstable file name
- During search for key, the sstables are sorted based on increasing order of sstable file name (sparse index + file creation timestamp)
- Duplicate spares index sstables are merged during write when threshold meets
1. If key greater than last sparse index, key is searched in sstable for the value and return
2. If greater than middle sparse index, search in middle sstable and write to memtable if found
  sstable[mid index + 1 : last index - 1] go to step 1
3. If smaller than middle sparse index, search in first sstable and write to memtable if found
  sstable[1 : mid index] go to step 1
