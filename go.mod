module github.com/apache/servicecomb-service-center

replace (
	github.com/apache/thrift => github.com/apache/thrift v0.0.0-20180125231006-3d556248a8b9
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/astaxie/beego v1.12.2
	github.com/cheggaaa/pb v1.0.25
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // v4
	github.com/deckarep/golang-set v1.7.1
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/elithrar/simple-scrypt v1.3.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-chassis/cari v0.5.1-0.20210723060050-4a4f119d64ff
	github.com/go-chassis/foundation v0.3.1-0.20210513015331-b54416b66bcd
	github.com/go-chassis/go-archaius v1.5.1
	github.com/go-chassis/go-chassis/v2 v2.2.1-0.20210630123055-6b4c31c5ad02
	github.com/go-chassis/kie-client v0.1.0
	github.com/golang/protobuf v1.4.3
	github.com/gorilla/websocket v1.4.3-0.20210424162022-e8629af678b7
	github.com/hashicorp/serf v0.8.3
	github.com/iancoleman/strcase v0.1.2
	github.com/jinzhu/copier v0.3.0
	github.com/karlseguin/ccache v2.0.3-0.20170217060820-3ba9789cfd2c+incompatible
	github.com/labstack/echo/v4 v4.1.18-0.20201218141459-936c48a17e97
	github.com/natefinch/lumberjack v0.0.0-20170531160350-a96e63847dc3
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin/zipkin-go-opentracing v0.3.3-0.20180123190626-6bb822a7f15f
	github.com/orcaman/concurrent-map v0.0.0-20210501183033-44dafcb38ecc
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.8.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/procfs v0.2.0
	github.com/rs/cors v1.7.0 // v1.1
	github.com/satori/go.uuid v1.1.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli v1.22.4
	github.com/widuu/gojson v0.0.0-20170212122013-7da9d2cd949b
	go.mongodb.org/mongo-driver v1.4.2
	go.uber.org/zap v1.13.0
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1
	google.golang.org/grpc v1.37.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.19.5
	k8s.io/apimachinery v0.19.5
	k8s.io/client-go v0.19.5
	k8s.io/kube-openapi v0.0.0-20210527164424-3c818078ee3d
)

go 1.16
