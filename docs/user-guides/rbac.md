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
before you start server, you need to set env to set your root account password.  

```sh
export SC_INIT_ROOT_PASSWORD=rootpwd
```
at the first time service center cluster init, it will use this env to setup rbac module. you can not revoke password after cluster start

the root account name is "root"

To securely distribute your root account and private key, 
you can use kubernetes [secret](https://kubernetes.io/zh/docs/tasks/inject-data-application/distribute-credentials-secure/)
### Generate a token 
token is the only credential to access rest API, before you access any API, you need to get a token
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
in each request you must add token to  http header:
```
Authorization: Bear {token}
```
for example:
```shell script
curl -X GET \
  'http://127.0.0.1:30100/v4/default/registry/microservices/{service-id}/instances' \
  -H 'Authorization: Bear eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTI4OTQ1NTEsInVzZXIiOiJyb290In0.FfLOSvVmHT9qCZSe_6iPf4gNjbXLwCrkXxKHsdJoQ8w' 
```

### Change password
You must supply current password and token to update to new password
```shell script
curl -X PUT \
  http://127.0.0.1:30100/v4/account-password \
  -H 'Authorization: Bear eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50Ijoicm9vdCIsImV4cCI6MTU5MzMyOTE3OSwicm9sZSI6IiJ9.OR_uruuLds1wz10_J4gDEA-L9Ma_1RrHiKEA6CS-Nilv6hHB5KyhZ9_4qqf_c0iia4uKryAGHKsXUvrjOE51tz4QCXlgCmddrkYuLQsnDezXhV3TIqzdl4R_cy8h2cZo8O_b_q7eU2Iemd6x7BJE49SLgNiP5LTXCVct5Qm_GiXYTaM4dbHIJ01V-EPmNQuBr1vKdfNa8cqWtASSp9IEkFx1YpzhFacQgmfoiSGHvxQYZldQXuAh60ZXLBDexGu6jGnG39MqVNRysvHTpZRqxZWBhmEn5DeXpgKu-zlakJMjeEma4zcN-H0MumE-nMlBT5kjKWVr1DOdtOyJI6i786ZpS0wWHV4VOxpSursoKsW_XuTZCMM8LTBgdy5icCuHUXvvWXYJxPks9Pq3DcFjPlY3IuXyfokEWxGvrAF6jzglgSrNTiRkoNBKVktEapDyrpyWfktp22mhvWF6GuNoUzztxFPJblH-TXdudzWeqx-gV1lsRPSMsW8-oq6pxJfeb-b0PNM8vAIbwvv8an4T5iNMBZMz7J9NbpVCaj5eLcgfUXktyb8eWSfANhYMxY9kQN9dHZlkASAW-sjehi-rBXYJ8aCL4EbLzrYlmFWoN0z25dxvAxmWaPRQED3METYyZHvV_G4DSQf0cB2Oer_YdoRa6HWmxnTlz0HwPEq55PM' \
  -d '{
	"currentPassword":"rootpwd",
	"password":"123"
}'
```

### Roles TODO
currently, you can not custom and manage any role and role policy. there is only 1 build in roles
- admin: able to do anything, including manage account, even change other account password
- service: able to call most of API except account management.