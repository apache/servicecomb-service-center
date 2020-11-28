# RBAC
alpha feature. now the feature is very simple in early stage. only has root account authentication

you can choose to enable RBAC feature, after enable RBAC, all request to service center must be authenticated

### Configuration file
follow steps to enable this feature.

1.get rsa key pairs
```sh
openssl genrsa -out private.key 4096
openssl rsa -in private.key -pubout -out public.key
```

2.edit app.conf
```ini
rbac_enabled = true
rbac_rsa_public_key_file = ./public.key # rsa key pairs
rbac_rsa_private_key_file = ./private.key # rsa key pairs
auth_plugin = buildin # must set to buildin
```
3.root account

before you start server, you need to set env to set your root account password. Please note that password must conform to the [following set of rules](https://github.com/apache/servicecomb-service-center/blob/63722fadd511c26285e787eb2b4be516eab10b94/pkg/validate/matcher.go#L25): have more than 8 characters, have at least one upper alpha, have at least one lower alpha, have at least one digit and have at lease one special character.

```sh
export SC_INIT_ROOT_PASSWORD='P4$$word'
```
at the first time service center cluster init, it will use this password to setup rbac module. 
you can revoke password by rest API after cluster started. but you can not use this env to revoke password after cluster started.

the root account name is "root"

To securely distribute your root account and private key, 
you can use kubernetes [secret](https://kubernetes.io/zh/docs/tasks/inject-data-application/distribute-credentials-secure/)
### Generate a token 
token is the only credential to access rest API, before you access any API, you need to get a token
```shell script
curl -X POST \
  http://127.0.0.1:30100/v4/token \
  -d '{"name":"root",
"password":"P4$$word"}'
```
will return a token, token will expired after 30m
```json
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTI4MzIxODUsInVzZXIiOiJyb290In0.G65mgb4eQ9hmCAuftVeVogN9lT_jNg7iIOF_EAyAhBU"}
```

### Authentication
in each request you must add token to  http header:
```
Authorization: Bearer {token}
```
for example:
```shell script
curl -X GET \
  'http://127.0.0.1:30100/v4/default/registry/microservices/{service-id}/instances' \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTI4OTQ1NTEsInVzZXIiOiJyb290In0.FfLOSvVmHT9qCZSe_6iPf4gNjbXLwCrkXxKHsdJoQ8w' 
```

### Change password
You must supply current password and token to update to new password
```shell script
curl -X POST \
  http://127.0.0.1:30100/v4/account/root/password \
  -H 'Authorization: Bearer {your_token}' \
  -d '{
	"currentPassword":"P4$$word",
	"password":"123"
}'
```

### create a new account 
You can create new account named "peter", now peter has no any roles, also has no permission to operate resources. How to add roles and allocate resources please refer to next.
```shell script
curl -X POST \
  http://127.0.0.1:30100/v4/account \
  -H 'Accept: */*' \
  -H 'Authorization: Bearer {your_token}' \
  -H 'Content-Type: application/json' \
  -d '{
	"name":"peter",
	"password":"{strong_password}"    
}'
```
### Roles 
Currently, two default roles are provided. You can also add new roles and assign resources.

### API and resources
All APIs of the system are divided according to their attributes. For example, resource account has the permission to create or update or delete user account when assign the corresponding permissions, resource service has all permission to create, get, add or delete microservices when permissions equal to "*". For more details to see [https://github.com/apache/servicecomb-service-center/blob/master/server/service/rbac/resource.go]()
A new role named "tester" owns resources "service", "instance" and "rule".
 ```json
{
 "name": "tester",
 "perms": [
         { 
            "resources": ["service","instance"],
            "verbs":     ["get", "create", "update"]
         },
         { 
             "resources": ["rule"],
             "verbs":     ["get"]
         }
    ]
}
```

### create new role and how to use
1. You can add new role and allocate resources to new role. For example, a new role named "tester" and allocate resources to "tester". 
```shell script
curl -X POST \
  http://127.0.0.1:30100/v4/role \
  -H 'Accept: */*' \
  -H 'Authorization: Bearer {your_token}' \
  -H 'Content-Type: application/json' \
  -d '{
	  "name": "tester",
      "perms": [
              { 
                  "resources": ["service","instance"],
                  "verbs":     ["get", "create", "update"]
              },
              { 
                  "resources": ["rule"],
                  "verbs":     ["get"]
              }
        ]
}'
```
2.then, assigning roles "tester" and "tester2" to user account "peter", "tester2" is a empty role has not any resources.
```shell script
curl -X POST \
  http://127.0.0.1:30100/v4/account \
  -H 'Accept: */*' \
  -H 'Authorization: Bearer {your_token}' \
  -H 'Content-Type: application/json' \
  -d '{
	"name":"peter",
	"password":"{strong_password}",
	"roles": ["tester", "tester2"]
}'
```

3.Next, generate a token for the user.
```shell script
curl -X POST \
  http://127.0.0.1:30100/v4/token \
  -d '{
  	"name":"peter",
  	"password":"{strong_password}"
  }'
```

4.finally, user "peter" carry token to access the above allocated API resources would be permit, but access others API is not allowed.

for example 
```shell script
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

```shell script
curl -X DElETE \
  http://127.0.0.1:30100/v4/default/registry/microservices \
  -H 'Accept: */*' \
  -H 'Authorization: Bearer {peter_token}' 
```
has no permission to operate.