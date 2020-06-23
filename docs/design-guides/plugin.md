# Plug-in mechanism
## Required
1. Go version 1.8(+)
1. Compile service-center with GO_EXTLINK_ENABLED=1 and CGO_ENABLED=1
1. The plugin file name must has suffix '_plugin.so'
1. All plugin interface files are in [plugin](https://github.com/apache/servicecomb-service-center/blob/master/server/plugin) package

## Plug-in names
1. **auth**: Customize authentication of service-center.
1. **uuid**: Customize micro-service/instance id format.
1. **auditlog**: Customize audit log for any change done to the service-center.
1. **cipher**: Customize encryption and decryption of TLS certificate private key password.
1. **quota**: Customize quota for instance registry.
1. **tracing**: Customize tracing data reporter.
1. **tls**: Customize loading the tls certificates in server

## Example: an authentication plug-in

### Step 1: code auth.go

auth.go is the implement from [auth interface](https://github.com/apache/servicecomb-service-center/blob/master/server/plugin/pkg/auth/auth.go)

```go
package main

import (
    "fmt"
    "net/http"
)

func Identify(*http.Request) error {
	// do something
	return nil
}
```

### Step 2: compile auth.go

```bash
GOPATH=$(pwd) go build -o auth_plugin.so -buildmode=plugin auth.go
```

### Step 3: move the plug-in in plugins directory

```bash
mkdir ${service-center}/plugins
mv auth_plugin.so ${service-center}/plugins
```

### Step 4: run service-center

```bash
cd ${service-center}
./servicecenter
```