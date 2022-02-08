# Quota plugins

### Standard Plugins

- buildin: standard quota management implement, read local quota configuration and limit the resource quotas.

### How to extend

1. Implement the interface `Manager` in server/plugin/quota/quota.go
```go
type Manager interface {
	RemandQuotas(ctx context.Context, t ResourceType)
	GetQuota(ctx context.Context, t ResourceType) int64
	Usage(ctx context.Context, req *Request) (int64, error)
}
```

2. Declare new instance func and register it to plugin manager
```go
import "github.com/apache/servicecomb-service-center/pkg/plugin"

plugin.RegisterPlugin(plugin.Plugin{Kind: quota.QUOTA, Name: "your plugin name", New: NewPluginInstanceFunc})
```

3. edit conf/app.yaml
```yaml
quota:
  kind: ${your plugin name}
```
