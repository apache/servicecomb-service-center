# Pact Broker Module

Pact broker module enables consumer-driven contract testing in service-center.

## Pact broker services

* Consumer microservices can publish pact.

```
PUT /pacts/provider/:providerId/consumer/:consumerId/version/:number
'{
	"provider" : "",
	"consumer" : "",
	...
	
}'
```

* Provider microservices can retrieve pacts published by consumer microservices, verify the pacts, and publish the verification results.
	
	1. Retrieving the information about the pacts associated with a provider

```
GET /pacts/provider/:providerId/latest

Response:
{
	"_links" : {
		"pacts" : [
			{
				"href" : "/pacts/provider/:providerId/consumer/:consumerId/version/:number"
				"name" : "Pact between consumerId and providerId with version number"
			}
		]
	}
}
```

	2. Retrieving the actual pact between a consumer and provider for verification

```
GET /pacts/provider/:providerId/consumer/:consumerId/version/:number
Response:
{
	"pact" : []byte
}
```

	3. Publishing the pact verification result to the broker

```
POST /pacts/provider/:providerId/consumer/:consumerId/pact-version/:number/verification-results
'{
	"success" : true
	"providerApplicationVersion" : "0.0.1"
}'

Response:
{
	"confirmation" : {
		"providerName" : ""
		"providerApplicationVersion" : "0.0.1"
		"success" : true
		"verificationDate" : ""
	}
}
```

* Consumer microservices can retrieve the verification results published by the providers.

```
GET /verification-results/consumer/:consumerId/version/:version/latest

Response:
{
	"result" : {
		"success" : true
		"providerSummary" : {
			"successful" : [
			]
			"failed" : [
			]
			"unknown" : [
			]
		}
		"_embedded" :  {
			"verificationResults" : [
				{
					"providerName" : ""
					"providerApplicationVersion" : ""
					"success" : true
					"verificationDate" : "" 
				}
			]
		}
	}
}
``` 

## protobuf to golang

```bash
cd brokerpb
protoc --go_out=plugins=grpc:. -I=$GOPATH/src -I=./  broker.proto 
```