#!/usr/bin/env bash
curl -X POST -H "Content-Type: application/json" -d '{
  "service_uuid": "822dc214-3e62-4bae-8411-7e1261e2c328",
  "attribute": "ServiceName",
  "pattern": "user.*8",
  "description": "test",
  "rule_type": "white"
}' "http://10.162.197.95:30200/service_center/v2/rules"