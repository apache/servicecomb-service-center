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
rbac_rsa_pub_key_file = ./public.key
```

before you start server, you need to set env to set your root account.  
can revoke private.key after each cluster restart,can not revoke root name and password
```sh
export SC_INIT_ROOT_USERNAME=root  
export SC_INIT_ROOT_PASSWORD=rootpwd
export SC_INIT_PRIVATE_KEY=`cat private.key`
```
at the first time service center cluster init, it will use this env to setup rbac module.

To securely distribute your root account and private key, 
you can use kubernetes [secret](https://kubernetes.io/zh/docs/tasks/inject-data-application/distribute-credentials-secure/)
### Generate a token 
```shell script
curl -X POST \
  http://127.0.0.1:30100/v4/token \
  -d '{"name":"root",
"password":"rootpwd"}'
```
will return a token, token will expired after 30m
```json
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTI4MzIxODUsInVzZXIiOiJyb290In0.G65mgb4eQ9hmCAuftVeVogN9lT_jNg7iIOF_EAyAhBU"}
```

### Authentication
in each request you must add token to  http header
Authorization: Bear {token}

for example:
```shell script
curl -X GET \
  'http://127.0.0.1:30100/v4/default/registry/microservices/{service-id}/instances' \
  -H 'Authorization: Bear eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTI4OTQ1NTEsInVzZXIiOiJyb290In0.FfLOSvVmHT9qCZSe_6iPf4gNjbXLwCrkXxKHsdJoQ8w' 
```