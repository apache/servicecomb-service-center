#!/usr/bin/env bash
POST /registry/v3/microservices/2/instances HTTP/1.1
Host: localhost:30100
Content-Type: application/json
x-domain-name: default
X-Project-Name: default
Cache-Control: no-cache
Postman-Token: bf33f47f-acfe-76fe-8c53-d1f79a46b246

{
	"instance": 
	{
	    "endpoints": [
			"ase://127.0.0.1:99841"
		],
		"hostName":"ase",
		"status":"UP",
		"environment":"production",
		"properties": {
			"_TAGS": "A, B",
			"attr1": "a",
			"nodeIP": "one"
		},
		"healthCheck": {
			"mode": "push",
			"interval": 30,
			"times": 2
		}
	}
}
