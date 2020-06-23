# Report trace data

## Edit the configuration of the tracing plugin
```bash
trace_plugin='buildin' # or empty
```

## To zipkin server

```eval_rst
.. image:: tracing-server.PNG
```

### Add the zipkin server endpoint
```
# Export the environments
export TRACING_COLLECTOR=server
export TRACING_SERVER_ADDRESS=http://127.0.0.1:9411 # zipkin server endpoint

# Start the Service-center
./servicecenter
```

## To file

```eval_rst
.. image:: tracing-file.PNG
```

### Customize the path of trace data file
```
# Export the environments
export TRACING_COLLECTOR=file
export TRACING_FILE_PATH=/tmp/servicecenter.trace # if not set, use ${work directory}/SERVICECENTER.trace

# Start the Service-center
./servicecenter
```
