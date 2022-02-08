# Quota management

### Resources

- service: microservice version quotas.
- instance: instance quotas.
- schema: schema quotas for each microservice.
- tag: tag quotas for each microservice.
- account: account quotas.
- role: role quotas.

### How to configure

#### 1. Use configuration file

edit conf/app.yaml
```yaml
quota:
  kind: buildin
  cap:
    service:
      limit: 50000
    instance:
      limit: 150000
    schema:
      limit: 100
    tag:
      limit: 100
    account:
      limit: 1000
    role:
      limit: 100
```

#### 2. Use environment variable

- QUOTA_SERVICE: the same as the config key `quota.cap.service.limit`
- QUOTA_INSTANCE: the same as the config key `quota.cap.instance.limit`
- QUOTA_SCHEMA: the same as the config key `quota.cap.schema.limit`
- QUOTA_TAG: the same as the config key `quota.cap.tag.limit`
- QUOTA_ACCOUNT: the same as the config key `quota.cap.account.limit`
- QUOTA_ROLE: the same as the config key `quota.cap.role.limit`
