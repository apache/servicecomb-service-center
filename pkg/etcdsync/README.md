# CSE Distributed Etcd lock

## example

```go
lock, _ := etcdsync.Lock("/test", true)
defer lock.Unlock()
//do something
g += 1
fmt.Println(g)
```