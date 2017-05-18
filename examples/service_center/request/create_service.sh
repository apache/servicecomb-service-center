curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache"  -d '{
  "service_level": "front",
  "desc": "this is a test content",
  "status": "In Service",
  "service_type": "billing system",
  "service_name": "billingsystem",
  "namespace": "default",
  "version": "v1",
  "domain": "",
  "app_id": "1"
}' "http://127.0.0.1:9980/service_center/v2/mservices"