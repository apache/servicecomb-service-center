Service-Center Design
=============

Service-Center(SC) is a service registry that allows services to
register their instance information and to discover providers of a given
service. Generally, SC uses etcd to store all the information of
micro-service and its instances.

.. image:: aggregator-design.PNG

-  **API Layer**: To expose the RESTful and gRPC service.
-  **Metedata**: The business logic to manage microservice, instance,
   schema, tag, dependency and ACL rules.
-  **Server Core**: Including data model, requests handle chain and so
   on.
-  **Aggregator**: It is the bridge between Core and Registry, includes
   the cache manager and indexer of registry.
-  **Registry Adaptor**: An abstract layer of registry, exposing a
   unified interface for upper layer calls.

Below is the diagram stating the working principles and flow of SC.

On StartUp
^^^^^^^^^^

Here describes a standard client registration process. We assume that
micro-services are written using `java-chassis`_ sdk or `go-chassis`_
sdk. So when micro-service boots up then java-chassis sdk does the
following list of tasks.

1. On startup provider registers the micro-service to SC if not
   registered earlier and also register its instance information like
   its Ip and Port on which instance is running.
2. SC stores the provider information in etcd.
3. On startup consumer retrieves the list of all provider instance from
   SC using the micro-service name of the provider.
4. Consumer sdk stores all the information of provider instances in its
   cache.
5. Consumer sdk creates a web socket connection to SC to watch all the
   provider instance information, if there is any change in the provider
   then sdk updates it’s cache information.

.. image:: onStartup.PNG

Communication
^^^^^^^^^^^^^

Once the bootup is successful then the consumer can communicate with
providers flawlessly, below is the diagram illustrating the
communication between provider and consumer.

.. image:: communication.PNG

| Provider instance regularly sends heartbeat signal every 30 seconds to
  SC, if SC does not receive the heartbeat for particular instance then
  the information in etcd expires and the provider instance information
  is removed.
| Consumer watches the information of provider instances from SC and if
  there is any change then the cache is updated.
| When Consumer needs to communicate to Provider then consumer reads
  endpoints of the provider instances from cache and do loadbalancing to
  communicate to Provider.

Note: Feel free to contribute to this document.

.. _java-chassis: https://github.com/ServiceComb/java-chassis
.. _go-chassis: https://github.com/go-chassis/go-chassis


Storage structure
^^^^^^^^^^^^^^^^^

Backend kind is ETCD

.. code:: yaml

    # services
    # /cse-sr/ms/files/{domain}/{project}/{serviceId}
    /cse-sr/ms/files/default/default/7062417bf9ebd4c646bb23059003cea42180894a:
      {
        "serviceId": "7062417bf9ebd4c646bb23059003cea42180894a",
        "appId": "default",
        "serviceName": "SERVICECENTER",
        "description": "A default service",
        "version": "0.0.1",
        "level": "BACK",
        "schemas": [
          "firstSchema",
          "secondSchema"
        ],
        "paths": [{
                    "path": "/opt/tomcat/webapp",
                    "property": {
                      "allowCrossApp": "true"
                    }
                  }],
        "status": "UP",
        "properties": {
          "allowCrossApp": "true"
        },
        "timestamp": "1592570701",
        "framework": {
          "name": "UNKNOWN",
          "version": "0.0.1"
        },
        "alias": "SERVICECENTER",
        "modTimestamp": "1592570701",
        "environment": "development"
      }

    # /cse-sr/ms/indexes/{domain}/{project}/{environment}/{appId}/{serviceName}/{serviceVersion}
    /cse-sr/ms/indexes/default/default/development/default/SERVICECENTER/0.0.1:
      "7062417bf9ebd4c646bb23059003cea42180894a"

    # /cse-sr/ms/alias/{domain}/{project}/{environment}/{appId}/{serviceName}/{serviceVersion}
    /cse-sr/ms/alias/default/default/development/default/SERVICECENTER/0.0.1:
      "7062417bf9ebd4c646bb23059003cea42180894a"

    # instances
    # /cse-sr/inst/files/{domain}/{project}/{serviceId}/{instanceId}
    /cse-sr/inst/files/default/default/7062417bf9ebd4c646bb23059003cea42180894a/b0ffb9feb22a11eaa76a08002706c83e:
      {
        "instanceId": "b0ffb9feb22a11eaa76a08002706c83e",
        "serviceId": "7062417bf9ebd4c646bb23059003cea42180894a",
        "endpoints": ["rest://127.0.0.1:30100/"],
        "hostName": "tian-VirtualBox",
        "status": "UP",
        "healthCheck": {
          "mode": "push",
          "interval": 30,
          "times": 3
        },
        "timestamp": "1592570701",
        "modTimestamp": "1592570701",
        "version": "0.0.1"
      }

    # /cse-sr/inst/leases/{domain}/{project}/{serviceId}/{instanceId}
    /cse-sr/inst/leases/default/default/7062417bf9ebd4c646bb23059003cea42180894a/b0ffb9feb22a11eaa76a08002706c83e:
      "leaseId"

    # schemas
    # /cse-sr/ms/schemas/{domain}/{project}/{serviceId}/{schemaId}
    /cse-sr/ms/schemas/default/default/7062417bf9ebd4c646bb23059003cea42180894a/first-schema:
      "schema"

    # /cse-sr/ms/schema-sum/{domain}/{project}/{serviceId}/{schemaId}
    /cse-sr/ms/schema-sum/default/default/7062417bf9ebd4c646bb23059003cea42180894a/first-schema:
      "schemaSummary"

    # dependencies
    # /cse-sr/ms/dep-queue/{domain}/{project}/{serviceId}/{uuid}
    /cse-sr/ms/dep-queue/default/default/7062417bf9ebd4c646bb23059003cea42180894a/0:
      {
        "consumer": {
          "tenant": "default/default",
          "project": "project",
          "appId": "appId",
          "serviceName": "ServiceCenter",
          "version": "0.0.1",
          "environment": "development",
          "alias": "serviceCenter"
        },
        "providers": [{
                       "tenant": "default/default",
                       "project": "project",
                       "appId": "appId",
                       "serviceName": "ServiceCenterProvider",
                       "version": "0.0.2",
                       "environment": "development",
                       "alias": "serviceCenterProvider"
                     }],
        "override": true
      }

    # tags
    # /cse-sr/ms/tags/{domain}/{project}/{serviceId}
    /cse-sr/ms/tags/default/default/7062417bf9ebd4c646bb23059003cea42180894a:
      {
        "a": "1"
      }

    # rules
    # /cse-sr/ms/rules/{domain}/{project}/{serviceId}/{ruleId}
    /cse-sr/ms/rules/default/default/7062417bf9ebd4c646bb23059003cea42180894a/Deny:
      {
        "ruleId": "Deny",
        "attribute": "denylist",
        "pattern": "Test*",
        "description": "test BLACK"
      }

    # /cse-sr/ms/rule-indexes/{domain}/{project}/{serviceId}/{attribute}/{pattern}
    /cse-sr/ms/rule-indexes/default/default/7062417bf9ebd4c646bb23059003cea42180894a/denylist/Test:
      "ruleId"

    # auth
    # /cse-sr/accounts/{accountName}
    /cse-sr/accounts/Alice:
      {
        "_id": "xxx",
        "account": "account_name",
        "password": "password",
        "role": "admin",
        "tokenExpirationTime": "1500519927",
        "currentPassword": "password",
        "status": "normal"
      }
    # record role binding to account
    /cse-sr/idx-role-account/{role}/{account}:
      {no value}
    # domain
    # /cse-sr/domains/{domain}
    /cse-sr/domains/default:

    # project
    # /cse-sr/domains/{domain}/{project}
    /cse-sr/projects/default/default:

Backend kind is Mongo

.. code:: yaml

    #type Service struct {
    #  Domain  string            `json:"domain,omitempty"`
    #  Project string            `json:"project,omitempty"`
    #  Tags    map[string]string `json:"tags,omitempty"`
    #  Service *pb.MicroService  `json:"service,omitempty"`
    #}

    #type MicroService struct {
    #  ServiceId    string             `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty" bson:"service_id"`
    #  AppId        string             `protobuf:"bytes,2,opt,name=appId" json:"appId,omitempty" bson:"app"`
    #  ServiceName  string             `protobuf:"bytes,3,opt,name=serviceName" json:"serviceName,omitempty" bson:"service_name"`
    #  Version      string             `protobuf:"bytes,4,opt,name=version" json:"version,omitempty"`
    #  Description  string             `protobuf:"bytes,5,opt,name=description" json:"description,omitempty"`
    #  Level        string             `protobuf:"bytes,6,opt,name=level" json:"level,omitempty"`
    #  Schemas      []string           `protobuf:"bytes,7,rep,name=schemas" json:"schemas,omitempty"`
    #  Paths        []*ServicePath     `protobuf:"bytes,10,rep,name=paths" json:"paths,omitempty"`
    #  Status       string             `protobuf:"bytes,8,opt,name=status" json:"status,omitempty"`
    #  Properties   map[string]string  `protobuf:"bytes,9,rep,name=properties" json:"properties,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
    #  Timestamp    string             `protobuf:"bytes,11,opt,name=timestamp" json:"timestamp,omitempty"`
    #  Providers    []*MicroServiceKey `protobuf:"bytes,12,rep,name=providers" json:"providers,omitempty"`
    #  Alias        string             `protobuf:"bytes,13,opt,name=alias" json:"alias,omitempty"`
    #  LBStrategy   map[string]string  `protobuf:"bytes,14,rep,name=LBStrategy" json:"LBStrategy,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value" bson:"lb_strategy"`
    #  ModTimestamp string             `protobuf:"bytes,15,opt,name=modTimestamp" json:"modTimestamp,omitempty" bson:"mod_timestamp"`
    #  Environment  string             `protobuf:"bytes,16,opt,name=environment" json:"environment,omitempty" bson:"env"`
    #  RegisterBy   string             `protobuf:"bytes,17,opt,name=registerBy" json:"registerBy,omitempty" bson:"register_by"`
    #  Framework    *FrameWork `protobuf:"bytes,18,opt,name=framework" json:"framework,omitempty"`
    #}

    #collection: service
    {
      "_id" : ObjectId("6021fb9527d99d766f82e44f"),
      "domain" : "new_default",
      "project" : "new_default",
      "tags" : null,
      "service" : {
        "service_id" : "6ea4d1c36a8311eba78dfa163e176e7b",
        "app" : "dep_create_dep_group",
        "service_name" : "dep_create_dep_consumer",
        "version" : "1.0.0",
        "description" : "",
        "level" : "FRONT",
        "schemas" : null,
        "paths" : null,
        "status" : "UP",
        "properties" : null,
        "timestamp" : "1612839829",
        "providers" : null,
        "alias" : "",
        "lb_strategy" : null,
        "mod_timestamp" : "1612839829",
        "env" : "",
        "register_by" : "",
        "framework" : null
      }
    }

    #type Instance struct {
    #  Domain      string                   `json:"domain,omitempty"`
    #  Project     string                   `json:"project,omitempty"`
    #  RefreshTime time.Time                `json:"refreshTime,omitempty" bson:"refresh_time"`
    #  Instance    *pb.MicroServiceInstance `json:"instance,omitempty"`
    #}

    #type MicroServiceInstance struct {
    #  InstanceId     string            `protobuf:"bytes,1,opt,name=instanceId" json:"instanceId,omitempty" bson:"instance_id"`
    #  ServiceId      string            `protobuf:"bytes,2,opt,name=serviceId" json:"serviceId,omitempty" bson:"service_id"`
    #  Endpoints      []string          `protobuf:"bytes,3,rep,name=endpoints" json:"endpoints,omitempty"`
    #  HostName       string            `protobuf:"bytes,4,opt,name=hostName" json:"hostName,omitempty"`
    #  Status         string            `protobuf:"bytes,5,opt,name=status" json:"status,omitempty"`
    #  Properties     map[string]string `protobuf:"bytes,6,rep,name=properties" json:"properties,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
    #  HealthCheck    *HealthCheck      `protobuf:"bytes,7,opt,name=healthCheck" json:"healthCheck,omitempty" bson:"health_check"`
    #  Timestamp      string            `protobuf:"bytes,8,opt,name=timestamp" json:"timestamp,omitempty"`
    #  DataCenterInfo *DataCenterInfo   `protobuf:"bytes,9,opt,name=dataCenterInfo" json:"dataCenterInfo,omitempty" bson:"data_center_info"`
    #  ModTimestamp   string            `protobuf:"bytes,10,opt,name=modTimestamp" json:"modTimestamp,omitempty" bson:"mod_timestamp"`
    #  Version        string            `protobuf:"bytes,11,opt,name=version" json:"version,omitempty"`
    #}

    # collection: instance
    {
      "_id" : ObjectId("60222c6f4fe067987f40803e"),
      "domain" : "default",
      "project" : "default",
      "refresh_time" : ISODate("2021-02-09T06:32:15.562Z"),
      "instance" : {
        "instance_id" : "8cde54a46aa011ebab42fa163e176e7b",
        "service_id" : "8cddc7ce6aa011ebab40fa163e176e7b",
        "endpoints" : [
            "find:127.0.0.9:8080"
        ],
        "hostname" : "UT-HOST-MS",
        "status" : "UP",
        "properties" : null,
        "health_check" : {
          "mode" : "push",
          "port" : 0,
          "interval" : 30,
          "times" : 3,
          "url" : ""
        },
        "timestamp" : "1612852335",
        "data_center_info" : null,
        "mod_timestamp" : "1612852335",
        "version" : "1.0.0"
      }
    }

    #type Schema struct {
    #  Domain        string `json:"domain,omitempty"`
    #  Project       string `json:"project,omitempty"`
    #  ServiceId     string `json:"serviceId,omitempty" bson:"service_id"`
    #  SchemaId      string `json:"schemaId,omitempty" bson:"schema_id"`
    #  Schema        string `json:"schema,omitempty"`
    #  SchemaSummary string `json:"schemaSummary,omitempty" bson:"schema_summary"`
    #}

    # collection schema
    {
      "_id" : ObjectId("6021fb9827d99d766f82e4f7"),
      "domain" : "default",
      "project" : "default",
      "service_id" : "70302da16a8311eba7cbfa163e176e7b",
      "schema_id" : "ServiceCombTestTheLimitOfSchemasServiceMS19",
      "schema" : "ServiceCombTestTheLimitOfSchemasServiceMS19",
      "schema_summary" : "ServiceCombTestTheLimitOfSchemasServiceMS19"
    }

    #type Rule struct {
    #  Domain    string          `json:"domain,omitempty"`
    #  Project   string          `json:"project,omitempty"`
    #  ServiceId string          `json:"serviceId,omitempty" bson:"service_id"`
    #  Rule      *pb.ServiceRule `json:"rule,omitempty"`
    #}

    #type ServiceRule struct {
    #  RuleId       string `protobuf:"bytes,1,opt,name=ruleId" json:"ruleId,omitempty" bson:"rule_id"`
    #  RuleType     string `protobuf:"bytes,2,opt,name=ruleType" json:"ruleType,omitempty" bson:"rule_type"`
    #  Attribute    string `protobuf:"bytes,3,opt,name=attribute" json:"attribute,omitempty"`
    #  Pattern      string `protobuf:"bytes,4,opt,name=pattern" json:"pattern,omitempty"`
    #  Description  string `protobuf:"bytes,5,opt,name=description" json:"description,omitempty"`
    #  Timestamp    string `protobuf:"bytes,6,opt,name=timestamp" json:"timestamp,omitempty"`
    #  ModTimestamp string `protobuf:"bytes,7,opt,name=modTimestamp" json:"modTimestamp,omitempty" bson:"mod_timestamp"`
    #}
    # collection rules
    {
      "_id" : ObjectId("6021fb9727d99d766f82e48a"),
      "domain" : "default",
      "project" : "default",
      "service_id" : "7026973b6a8311eba792fa163e176e7b",
      "rule" : {
        "rule_id" : "702897cf6a8311eba79dfa163e176e7b",
        "rule_type" : "BLACK",
        "attribute" : "ServiceName",
        "pattern" : "18",
        "description" : "test white",
        "timestamp" : "1612839831",
        "mod_timestamp" : "1612839831"
      }
    }

    #type ConsumerDep struct {
    #  Domain      string                 `json:"domain,omitempty"`
    #  Project     string                 `json:"project,omitempty"`
    #  ConsumerId  string                 `json:"consumerId,omitempty" bson:"consumer_id"`
    #  UUId        string                 `json:"uuId,omitempty" bson:"uu_id"`
    #  ConsumerDep *pb.ConsumerDependency `json:"consumerDep,omitempty" bson:"consumer_dep"`
    #}

    #type ConsumerDependency struct {
    #  Consumer  *MicroServiceKey   `protobuf:"bytes,1,opt,name=consumer" json:"consumer,omitempty"`
    #  Providers []*MicroServiceKey `protobuf:"bytes,2,rep,name=providers" json:"providers,omitempty"`
    #  Override  bool               `protobuf:"varint,3,opt,name=override" json:"override,omitempty"`
    #}

    #type MicroServiceKey struct {
    #  Tenant      string `protobuf:"bytes,1,opt,name=tenant" json:"tenant,omitempty"`
    #  Environment string `protobuf:"bytes,2,opt,name=environment" json:"environment,omitempty" bson:"env"`
    #  AppId       string `protobuf:"bytes,3,opt,name=appId" json:"appId,omitempty" bson:"app"`
    #  ServiceName string `protobuf:"bytes,4,opt,name=serviceName" json:"serviceName,omitempty" bson:"service_name"`
    #  Alias       string `protobuf:"bytes,5,opt,name=alias" json:"alias,omitempty"`
    #  Version     string `protobuf:"bytes,6,opt,name=version" json:"version,omitempty"`
    #}

    # collection dependencies
    {
      "_id" : ObjectId("6021fb9527d99d766f82e45f"),
      "domain" : "new_default",
      "project" : "new_default",
      "consumer_id" : "6ea4d1c36a8311eba78dfa163e176e7b",
      "uu_id" : "6eaeb1dd6a8311eba790fa163e176e7b",
      "consumer_dep" : {
        "consumer" : {
          "tenant" : "new_default/new_default",
          "env" : "",
          "app" : "dep_create_dep_group",
          "service_name" : "dep_create_dep_consumer",
          "alias" : "",
          "version" : "1.0.0"
        },
        "providers" : null,
        "override" : false
      }
    }

    #type DependencyRule struct {
    #  Type       string                     `json:"type,omitempty"`
    #  Domain     string                     `json:"domain,omitempty"`
    #  Project    string                     `json:"project,omitempty"`
    #  ServiceKey *pb.MicroServiceKey        `json:"serviceKey,omitempty" bson:"service_key"`
    #  Dep        *pb.MicroServiceDependency `json:"dep,omitempty"`
    #}

    #type MicroServiceKey struct {
    #  Tenant      string `protobuf:"bytes,1,opt,name=tenant" json:"tenant,omitempty"`
    #  Environment string `protobuf:"bytes,2,opt,name=environment" json:"environment,omitempty" bson:"env"`
    #  AppId       string `protobuf:"bytes,3,opt,name=appId" json:"appId,omitempty" bson:"app"`
    #  ServiceName string `protobuf:"bytes,4,opt,name=serviceName" json:"serviceName,omitempty" bson:"service_name"`
    #  Alias       string `protobuf:"bytes,5,opt,name=alias" json:"alias,omitempty"`
    #  Version     string `protobuf:"bytes,6,opt,name=version" json:"version,omitempty"`
    #}

    #type MicroServiceDependency struct {
    #  Dependency []*MicroServiceKey `json:"Dependency,omitempty"`
    #}

    # collection dependencies
    {
      "_id" : ObjectId("6022302751a77062a95dd0da"),
      "service_key" : {
        "app" : "create_dep_group",
        "env" : "production",
        "service_name" : "create_dep_consumer",
        "tenant" : "default/default",
        "version" : "1.0.0"
      },
      "type" : "c",
      "dep" : {
        "dependency" : [
          {
            "tenant" : "default/default",
            "env" : "",
            "app" : "service_group_provider",
            "service_name" : "service_name_provider",
            "alias" : "",
            "version" : "latest"
          }
        ]
      }
    }


    #type Account struct {
    #  ID                  string   `json:"id,omitempty"`
    #  Name                string   `json:"name,omitempty"`
    #  Password            string   `json:"password,omitempty"`
    #  Roles               []string `json:"roles,omitempty"`
    #  TokenExpirationTime string   `json:"tokenExpirationTime,omitempty" bson:"token_expiration_time"`
    #  CurrentPassword     string   `json:"currentPassword,omitempty" bson:"current_password"`
    #  Status              string   `json:"status,omitempty"`
    #}

    # collection account
    {
      "_id" : ObjectId("60223e99184f264aee398238"),
      "id" : "6038bf9f6aab11ebbcdefa163e176e7b",
      "name" : "test-account1",
      "password" : "$2a$14$eYyD9DiOA1vGXOyhPTjbhO6CYuGnOVt8VQ8V/sWEmExyvwOQeNI2i",
      "roles" : [
          "admin"
      ],
      "token_expiration_time" : "2020-12-30",
      "current_password" : "tnuocca-tset1",
      "status" : ""
    }
