curl -X POST -H "Content-Type: application/json" -d '{
  "instance": {
    "hostName": "lfgy11",
    "serviceUuid": "3c9b7d12-57d1-4105-b6ff-df7b14c1c53b",
    "ipAddr": "10.130.158.159",
    "status": "UP",
    "overriddenstatus": "UNKNOWN",
    "port": {
      "$": 9980,
      "@enabled": "true"
    },
    "securePort": {
      "$": 0,
      "@enabled": "false"
    },
    "countryId": 0,
    "dataCenterInfo": {
      "name": "MyOwn"
    },
    "metadata": {
        "test":"value"
    },
    "homePageUrl": "",
    "statusPageUrl": "",
    "healthCheckUrl": "v2/version",
    "vipAddress": "",
    "secureVipAddress": "",
    "isCoordinatingDiscoveryServer": "true",
    "lastUpdatedTimestamp": "1471394003477",
    "lastDirtyTimestamp": "1471394002952"
  }
}' "http://localhost:9980/registry/v2/mservices/3c9b7d12-57d1-4105-b6ff-df7b14c1c53b"