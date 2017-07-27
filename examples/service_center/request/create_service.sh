#!/usr/bin/env bash
POST /registry/v3/microservices HTTP/1.1
Host: localhost:30100
Content-Type: application/json
x-domain-name: default
Cache-Control: no-cache
Postman-Token: 8378df39-dfac-4f9d-205b-981e367c6b05

{
	"service":
	{
		"serviceName": "Test10",
		"appId": "TestApp10",
		"version":"11.0.0",
		"asdasddescription":"examplfsdsfe",
		"level": "FRONT",
		"schemas": [
			"TestService1212"
		],
		"status": "DOWN",
		"properties": {
			"attr2": "a"
		},
		"description":"中文asdfadads。”"
	}
}
