# RBAC

you can choose to enable RBAC feature, after enable RBAC, 
all request to service center must be authenticated

### Configuration file
Follow steps to enable this feature.

1.get rsa key pairs
```sh
openssl genrsa -out private.key 4096
openssl rsa -in private.key -pubout -out public.key
```

2.edit app.yaml
```yaml
rbac:
  enable: true
  privateKeyFile: ./private.key # rsa key pairs
  publicKeyFile: ./public.key # rsa key pairs
auth:
  kind: buildin # must set to buildin
```
3.root account

before you start server, you need to set env to set your root account password. 
Please note that password must conform to the 
[following set of rules](https://github.com/apache/servicecomb-service-center/blob/63722fadd511c26285e787eb2b4be516eab10b94/pkg/validate/matcher.go#L25): 
have at least 8 characters, have at most 32 characters, have at least one upper alpha, have at least one lower alpha, 
have at least one digit and have at lease one special character.

```sh
export SC_INIT_ROOT_PASSWORD='P4$$word'
```
At the first time service center cluster init, it will use this password to set up rbac module. 
you can revoke password by rest API after a cluster started. but you can not use **SC_INIT_ROOT_PASSWORD** to revoke password after a cluster started.

the initiated account name is fixed as "root"

To securely distribute your root account and private key, 
you can use kubernetes [secret](https://kubernetes.io/zh/docs/tasks/inject-data-application/distribute-credentials-secure/)
### Generate a token 
Token is the only credential to access rest API, before you access any API, you need to get a token from service center
```shell
curl -X POST \
  http://127.0.0.1:30100/v4/token \
  -d '{"name":"root",
"password":"P4$$word"}'
```
will return a token, token will expire after 30m
```json
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTI4MzIxODUsInVzZXIiOiJyb290In0.G65mgb4eQ9hmCAuftVeVogN9lT_jNg7iIOF_EAyAhBU"}
```

### Authentication
in each request you must add token to  http header:
```
Authorization: Bearer {token}
```
for example:
```shell
curl -X GET \
  'http://127.0.0.1:30100/v4/default/registry/microservices/{service-id}/instances' \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTI4OTQ1NTEsInVzZXIiOiJyb290In0.FfLOSvVmHT9qCZSe_6iPf4gNjbXLwCrkXxKHsdJoQ8w' 
```

### Change password
You must supply a current password and token to update to new password
```shell
curl -X POST \
  http://127.0.0.1:30100/v4/account/root/password \
  -H 'Authorization: Bearer {your_token}' \
  -d '{
	"currentPassword":"P4$$word",
	"password":"P4$$word1"
}'
```

### create a new account 
You can create new account named "peter", and his role is developer.
How to add roles and allocate resources please refer to next section.
```shell
curl -X POST \
  http://127.0.0.1:30100/v4/account \
  -H 'Accept: */*' \
  -H 'Authorization: Bearer {your_token}' \
  -H 'Content-Type: application/json' \
  -d '{
	"name":"peter",
	"roles":["developer"],
	"password":"{strong_password}"
}'
```

### Resource 
All APIs of the ServiceComb system is mapping to a **resource type**. resource is list as below:
- service: permission to discover, register service and instance
- governance:  permission to manage traffic control policy, such as rate limiting
- service/schema: permission to register and discover contract
- account: permission to manage accounts and account-locks
- role: permission to manage roles
- ops: permission to access admin API

declare a resource type that account can operate:
```json
 {
  "resources": [
    {
      "type": "service"
    },
    {
      "type": "service/schema"
    }
  ]
}
```
### Label
Define resource(only service resource) scope:
- serviceName: specify service name 
- appId: specify which app that services belongs to
- environment: specify env of the service

```json
{
  "resources": [
    {
      "type": "service",
      "labels": {
        "serviceName": "order-service",
        "environment": "production"
      }
    },
    {
      "type": "service",
      "labels": {
        "serviceName": "order-service",
        "environment": "acceptance"
      }
    }
  ]
}
```
### Verbs
Define what kind of action could be applied to a resource by an account, has 4 kinds:
- get
- delete
- create
- update

declare resource type and action:
```json
{
  "resources": [
    {
      "type": "service"
    },
    {
      "type": "account"
    }
  ],
  "verbs": [
    "get"
  ]
}
```

### Roles
Two default roles are provided after RBAC init:
- admin: can operate account and role resource
- developer: can operate any resource except account and role resource

each role include perms elements to indicates what kind of resource can be operated by this role, for example:

A role "TeamA" can get and create any services but can only delete or update "order-service"
```json
{
  "name": "TeamA",
  "perms": [
    {
      "resources": [
        {
          "type": "service"
        }
      ],
      "verbs": [
        "get",
        "create"
      ]
    },
    {
      "resources": [
        {
          "type": "service",
          "labels": {
            "serviceName": "order-service"
          }
        }
      ],
      "verbs": [
        "update",
        "delete"
      ]
    }
  ]
}
```




### create new role and how to use

You can also create a new role and give perms to this role.

1. You can add new role and allocate resources to new role. For example, a new role named "tester" and allocate resources to "tester". 
```shell
curl -X POST \
  http://127.0.0.1:30100/v4/role \
  -H 'Accept: */*' \
  -H 'Authorization: Bearer {your_token}' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "TeamA",
  "perms": [
    {
      "resources": [
        {
          "type": "service"
        }
      ],
      "verbs": [
        "get",
        "create"
      ]
    },
    {
      "resources": [
        {
          "type": "service",
          "labels": {
            "serviceName": "order-service"
          }
        }
      ],
      "verbs": [
        "update",
        "delete"
      ]
    }
  ]
}'
```
2.then, assigning roles "tester" and "tester2" to user account "peter", "tester2" is a empty role has not any resources.
```shell
curl -X POST \
  http://127.0.0.1:30100/v4/account \
  -H 'Accept: */*' \
  -H 'Authorization: Bearer {your_token}' \
  -H 'Content-Type: application/json' \
  -d '{
	"name":"peter",
	"password":"{strong_password}",
	"roles": ["TeamA"]
}'
```

3.Next, generate a token for the user.
```shell
curl -X POST \
  http://127.0.0.1:30100/v4/token \
  -d '{
  	"name":"peter",
  	"password":"{strong_password}"
  }'
```

4.finally, user "peter" carry token to access resources.

for example 
```shell
curl -X POST \
  http://127.0.0.1:30100/v4/default/registry/microservices \
  -H 'Accept: */*' \
  -H 'Authorization: Bearer {peter_token}' \
  -d '{
        "service": {
          "serviceId": "11111-22222-33333",
          "appId": "test",
          "serviceName": "test",
          "version": "1.0.0"
        }
}'
```
would be ok.

```shell
curl -X DElETE \
  http://127.0.0.1:30100/v4/default/registry/microservices \
  -H 'Accept: */*' \
  -H 'Authorization: Bearer {peter_token}' 
```
has no permission to operate.