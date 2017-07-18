# CSE Distributed Etcd lock

## example

```bash
lock, _ := etcdsync.Lock("/test")
defer lock.Unlock()
//do something
g += 1
fmt.Println(g)
```