#!/usr/bin/env bash
PUT /v4/default/registry/microservices/serviceID/instances/instanceID/status?value=UP HTTP/1.1
Host: localhost:30100
Content-Type: application/json
x-domain-name: default
Cache-Control: no-cache
Postman-Token: ceec0671-e415-804d-ce80-886c64159317

