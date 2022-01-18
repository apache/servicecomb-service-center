# eventbase

eventbase provides the crud interface of task and tombstone.

### package

**bootstrap**：used to start initial loading.

**datasource**: realize the dao operation of etcd and mongo on task and tombstone.

**model**: task and tombstone request.

**service**: Interfaces exposed by task and tombstone.

**test**: test parameters.

### how to use

1.First you should initialize db

```go
import (
    _ "github.com/go-chassis/cari/db/bootstrap"
    "github.com/go-chassis/cari/db"
)

func Init(){
    cfg := config.Config{
    Kind: "etcd",
    URI: "http://127.0.0.1:2379",
    Timeout: 10 * time.Second,
    }
    err := db.Init(&cfg)
}
```


2.Second you should import the eventbase's bootstrap
```go
import (
    _ "github.com/apache/servicecomb-service-center/eventbase/bootstrap"
)
```
3.Third you should do Init func

```go

import (
    "github.com/go-chassis/cari/db/config"
    "github.com/go-chassis/cari/db"
    
    _ "github.com/apache/servicecomb-service-center/eventbase/bootstrap"
    "github.com/apache/servicecomb-service-center/eventbase/datasource"
    "github.com/apache/servicecomb-service-center/eventbase/service/task"
    "github.com/apache/servicecomb-service-center/eventbase/service/tombstone"
)

func main(){
	cfg := config.Config{
		Kind: "etcd",
		URI: "http://127.0.0.1:2379",
		Timeout: 10 * time.Second,
	}
	err := db.Init(&cfg)
	err = datasource.Init("etcd")
	...
	tasksvc.List(...)
	tombstonesvc.List(...)
	...
}
```