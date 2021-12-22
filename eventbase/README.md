# eventbase

eventbase provides the crud interface of task and tombstone.

### package

**bootstrap**：used to start initial loading.

**datasource**: realize the dao operation of etcd and mongo on task and tombstone.

**domain**: task and tombstone request.

**service**: Interfaces exposed by task and tombstone.

**test**: test parameters.

### how to use

```go
import (
	_ "github.com/apache/servicecomb-service-center/eventbase/bootstrap"
	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	)

func Init(){
    dbCfg := db.Config{
    	Kind: "etcd",
    	URI: "http://127.0.0.1:2379",
    	Timeout: 10 * time.Second,
    }
    err := datasource.Init(dbCfg)
    ...
    datasource.GetDataSource().TaskDao()
    datasource.GetDataSource().TombstoneDao()
    ...
}
```