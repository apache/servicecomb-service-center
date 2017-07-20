#!/usr/bin/env bash
PUT /registry/v3/microservices/2/properties HTTP/1.1
Host: localhost:30100
Content-Type: application/json
x-domain-name: default
Cache-Control: no-cache
Postman-Token: 06a8e72d-caaf-34b3-89bc-710aa418d5f3

{
	"properties": {
		"attr1": "b"
	}
}
