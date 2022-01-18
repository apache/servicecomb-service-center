module github.com/apache/servicecomb-service-center

replace (
	github.com/apache/servicecomb-service-center/api => ./api
	github.com/apache/servicecomb-service-center/eventbase => ./eventbase
	github.com/apache/thrift => github.com/apache/thrift v0.0.0-20180125231006-3d556248a8b9
)

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/astaxie/beego v1.12.2
	github.com/cheggaaa/pb v1.0.25
	github.com/deckarep/golang-set v1.7.1
	github.com/elithrar/simple-scrypt v1.3.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-chassis/cari v0.5.1-0.20220115081247-ed7ae4111a1b
	github.com/go-chassis/foundation v0.4.0
	github.com/go-chassis/go-archaius v1.5.1
	github.com/go-chassis/go-chassis-extension/protocol/grpc v0.0.0-20210902082902-eb5df922afcd
	github.com/go-chassis/go-chassis/v2 v2.3.0
	github.com/go-chassis/kie-client v0.1.1-0.20210926011742-97eed4281056
	github.com/go-chassis/openlog v1.1.3
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/gorilla/websocket v1.4.3-0.20210424162022-e8629af678b7
	github.com/iancoleman/strcase v0.1.2
	github.com/jinzhu/copier v0.3.0
	github.com/karlseguin/ccache v2.0.3-0.20170217060820-3ba9789cfd2c+incompatible
	github.com/labstack/echo/v4 v4.1.18-0.20201218141459-936c48a17e97
	github.com/little-cui/etcdadpt v0.3.2
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin/zipkin-go-opentracing v0.3.3-0.20180123190626-6bb822a7f15f
	github.com/orcaman/concurrent-map v0.0.0-20210501183033-44dafcb38ecc
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/procfs v0.6.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/rs/cors v1.7.0 // v1.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli v1.22.4
	github.com/widuu/gojson v0.0.0-20170212122013-7da9d2cd949b
	go.etcd.io/etcd/api/v3 v3.5.0
	go.etcd.io/etcd/client/v3 v3.5.0
	go.mongodb.org/mongo-driver v1.4.2
	go.uber.org/zap v1.17.0
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	google.golang.org/grpc v1.40.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	k8s.io/api v0.19.5
	k8s.io/apimachinery v0.19.5
	k8s.io/client-go v0.19.5
	k8s.io/kube-openapi v0.0.0-20210527164424-3c818078ee3d
)

require (
	github.com/apache/servicecomb-service-center/eventbase v0.0.0-00010101000000-000000000000
	github.com/spf13/cast v1.3.0
)

require (
	github.com/Shopify/sarama v1.19.0 // indirect
	github.com/Shopify/toxiproxy v2.1.4+incompatible // indirect
	github.com/apache/thrift v0.13.0 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/eapache/go-resiliency v1.1.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/pierrec/lz4 v2.0.5+incompatible // indirect
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a // indirect
	google.golang.org/protobuf v1.27.1
	gopkg.in/cheggaaa/pb.v1 v1.0.25 // indirect
)

go 1.17
