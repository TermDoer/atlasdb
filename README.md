# Implementation
## 1. Go map (In-Memory datastore)
### Architecture
```
      Client
        │
     atlasctl
        │
      atlasd
        │
┌───────────────┐
│    Go map     │
└───────────────┘
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
           │
        atlasctl
           │
         atlasd
           │
      ┌──────────┐
      │ MemTable │
      └────┬─────┘
           │
           ▼
     Write Ahead Log
           │
     append-only file
```
### Drawbacks
- SSD memory delete, or file delete
