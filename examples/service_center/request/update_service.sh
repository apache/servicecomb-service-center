#!/usr/bin/env bash
curl -X PUT -H "Content-Type: application/json"  -d '{
  "service_level": "front",
  "desc": "this is a test content",
  "status": "In Service",
  "service_type": "billing system",
  "service_name": "billingsystem",
  "namespace": "default",
  "version": "v1",
  "domain": "",
  "app_id": "1"
}' "http://127.0.0.1:9980/service_center/v2/mservices/08a8db3c-c267-48f1-9582-f314fb8a2f4d"