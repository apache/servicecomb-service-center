#!/usr/bin/env bash
PUT /registry/v3/dependencies HTTP/1.1
Host: localhost:30100
Content-Type: application/json
x-domain-name: default
Cache-Control: no-cache
Postman-Token: e924c839-df99-5526-5c30-f7ca9a671316

{
    "dependencies": [
        {
            "consumer": {
                "appId": "TestApp8",
                "serviceName": "Test8",
                "version": "1.0.0"
            },
            "providers": [
                {
                     "serviceName": "*",
                     "appId": "TestApp3",
                     "version":"1.0.0"
                }
            ]
        }
    ]
}
