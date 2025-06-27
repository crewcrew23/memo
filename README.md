# Memo this is In-Memory Cache Library

Lightweight thread-safe cache with TTL and metrics.

## 🔧 Features

- Support for arbitrary types (generics)
- Automatic deletion by TTL
- Context support
- Thread safety (sync.RWMutex)
- Usage metrics (hit rate, evictions)
- Callbacks when deleting elements
- Serialization/deserialization (JSON)

## ✅ Install
```bash
go get will be here
```

## 🚀 Quick start
```go
func main() {
	cache := memo.New[int]()
	ctx := context.Background()

    //analog without context
    // cache.Set("key", 2, time.Minute*5)
	if err := cache.SetWithContext(ctx, "key", 2, time.Minute*5); err != nil {
		log.Println(err)
	}

    //analog without context
    // cache.Get("key")
    // If the value does not exist, a null value and an error will be returned
	val, err := cache.GetWithContext(ctx, "key")
	if err != nil {
		log.Println(err)
	}
}
```

## Marshal/Unmarshal
- the cache is stored in memory so if you need to save the state across restarts then use Marshal/Unmarshal

```go
func main() {
	cache := memo.New[int]()
	ctx := context.Background()

	if err := cache.SetWithContext(ctx, "key", 2, time.Minute*5); err != nil {
		log.Println(err)
	}

	//analog without context
	//cache.MarshalJSON()
	bytes, err := cache.MarshalJSONWithContext(ctx)
	if err != nil {
		log.Println(err)
	}

	unmarshalCache := memo.New[int]()

	//analog without context
	//cache.UnmarshalJSON(bytes)
	if err := unmarshalCache.UnmarshalJSONWithContext(ctx, bytes); err != nil {
		log.Println(err)
	}
}
```
## OnEvicted
 - OnEvicted will be called on the element when it is deleted
 - OnEvicted can return error only if cache closed
```go
func main() {
	cache := memo.New[int]()

	if err := cache.OnEvicted(func(key string, value int) {
		log.Printf("%s:%d was be deleted", key, value)
	}); err != nil {
        log.Println(err)
    }
}
```

## Close
- when closing, the internal context will be cancelled and the cleanup goroutine will be stopped
the internal map will be nil and access to methods:
OnEvicted 
Set
SetWithContext
Get
GetWithContext
UnmarshalJSON
MarshalJSONWithContext
UnmarshalJSON
UnmarshalJSONWithContext

if need save state before close you need use MarshalJSON or MarshalJSONWithContext

```go
func main() {
	cache := memo.New[int]()

	cache.Set("key", 2, time.Minute*1)

	cache.Close()

	if _, err := cache.Get("key"); err != nil {
		log.Println(err)
		//will be out "cache is closed"
	}
}
```


## Statistic
can be accessed after closing
```go
type Stats struct {
	Hits      uint64
	Misses    uint64
	Evictions uint64
	HitRate   float64
	SizeBytes int64
}

//return Stats struct
stat := cache.Stat()
```