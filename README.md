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
### Drawbacks
- SSD memory delete, or file delete

## 3. SSTables
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
    -----------
    | MemTable |
     -----------
         |
      (Flush)
   SSTable-001.db
   SSTable-002.db
```
### Drawbacks
- Read overhead
