# Setup SSL/TLS

## Requirement
Service center(SC) takes several files related SSL/TLS options.

1. Environment variable 'SSL_ROOT': The directory contains certificates. If not set, uses 'etc/ssl' under the SC work directory.
1. `$SSL_ROOT`/trust.cer: Trusted certificate authority.
1. `$SSL_ROOT`/server.cer: Certificate used for SSL/TLS connections to SC.
1. `$SSL_ROOT`/server_key.pem: Key for the certificate. If key is encrypted, 'cert_pwd' must be set.
1. `$SSL_ROOT`/cert_pwd(optional): The password used to decrypt the private key.

## Configuration
Please modify the conf/app.conf before start up SC

1. ssl_mode: Enabled SSL/TLS mode. [0, 1]
1. ssl_verify_client: Whether the SC verify client(including etcd server). [0, 1]
1. ssl_min_version: Minimal SSL/TLS protocol version. ["TLSv1.0", "TLSv1.1", "TLSv1.2", "TLSv1.3"], based on Go version
1. ssl_ciphers: A list of cipher suite. By default, uses `TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256`, `TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384`, `TLS_RSA_WITH_AES_256_GCM_SHA384`, `TLS_RSA_WITH_AES_128_GCM_SHA256`
