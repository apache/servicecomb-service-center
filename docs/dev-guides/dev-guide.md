# Development Guide

This chapter is about how to implement the feature of micro-service discovery with ServiceCenter,
and you can get more detail at [here](https://github.com/apache/servicecomb-service-center/blob/master/server/core/swagger/v3.yaml)

## Micro-service registration
```bash
curl -X POST \
  http://127.0.0.1:30100/registry/v3/microservices \
  -H 'content-type: application/json' \
  -H 'x-domain-name: default' \
  -d '{
	"service":
	{
		"appId": "default",
		"serviceName": "DemoService",
		"version":"1.0.0"
	}
}'
```

and then you can get the 'DemoService' ID like below:

```json
{
    "serviceId": "a3fae679211211e8a831286ed488fc1b"
}
```

## Instance registration

mark down the micro-service ID and call the instance registration API,
according to the ServiceCenter definition: One process should be registered as one instance

```bash
curl -X POST \
  http://127.0.0.1:30100/registry/v3/microservices/a3fae679211211e8a831286ed488fc1b/instances \
  -H 'content-type: application/json' \
  -H 'x-domain-name: default' \
  -d '{
	"instance": 
	{
	    "hostName":"demo-pc",
	    "endpoints": [
		    "rest://127.0.0.1:8080"
	    ]
	}
}'
```

the successful response like below:

```json
{
    "instanceId": "288ad703211311e8a831286ed488fc1b"
}
```

if all are successful, it means you have completed the micro-service registration and instance publish

## Discovery

the next step is that discovery the micro-service instance by service name and version rule

```bash
curl -X GET \
  'http://127.0.0.1:30100/registry/v3/instances?appId=default&serviceName=DemoService&version=latest' \
  -H 'content-type: application/json' \
  -H 'x-consumerid: a3fae679211211e8a831286ed488fc1b' \
  -H 'x-domain-name: default'
```

here, you can get the information from the response

```json
{
    "instances": [
        {
            "instanceId": "b4c9e57f211311e8a831286ed488fc1b",
            "serviceId": "a3fae679211211e8a831286ed488fc1b",
            "version": "1.0.0",
            "hostName": "demo-pc",
            "endpoints": [
                "rest://127.0.0.1:8080"
            ],
            "status": "UP",
            "healthCheck": {
                "mode": "push",
                "interval": 30,
                "times": 3
            },
            "timestamp": "1520322915",
            "modTimestamp": "1520322915"
        }
    ]
}
```