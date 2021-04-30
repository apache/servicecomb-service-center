module github.com/apache/servicecomb-service-center

replace google.golang.org/grpc => google.golang.org/grpc v1.19.0

require (
	github.com/NYTimes/gziphandler v1.0.2-0.20180820182813-253f1acb9d9f
	github.com/Shopify/sarama v1.27.2 // indirect
	github.com/apache/thrift v0.0.0-20180125231006-3d556248a8b9 // indirect
	github.com/astaxie/beego v1.8.0
	github.com/aws/aws-sdk-go v1.36.30 // indirect
	github.com/cheggaaa/pb v1.0.25
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/go-systemd v0.0.0-20190620071333-e64a0ec8b42a // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // v4
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/elithrar/simple-scrypt v1.3.0
	github.com/fatih/color v1.10.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-chassis/cari v0.3.1-0.20210508100214-a13e083de04e
	github.com/go-chassis/foundation v0.3.0
	github.com/go-chassis/go-archaius v1.5.1
	github.com/go-chassis/go-chassis/v2 v2.1.2-0.20210310004133-c9bc42149a18
	github.com/go-chassis/kie-client v0.1.0
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.4.2
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-version v1.2.1 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/serf v0.8.3
	github.com/iancoleman/strcase v0.1.2
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jinzhu/copier v0.3.0
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/karlseguin/ccache v2.0.3-0.20170217060820-3ba9789cfd2c+incompatible
	github.com/karlseguin/expect v1.0.7 // indirect
	github.com/labstack/echo/v4 v4.1.18-0.20201218141459-936c48a17e97
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/miekg/dns v1.1.35 // indirect
	github.com/natefinch/lumberjack v0.0.0-20170531160350-a96e63847dc3
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin/zipkin-go-opentracing v0.3.3-0.20180123190626-6bb822a7f15f
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/procfs v0.1.3
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rs/cors v1.7.0 // v1.1
	github.com/satori/go.uuid v1.1.0
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20200427203606-3cfed13b9966 // indirect
	github.com/urfave/cli v1.22.4
	github.com/widuu/gojson v0.0.0-20170212122013-7da9d2cd949b
	go.etcd.io/bbolt v1.3.5 // indirect
	go.mongodb.org/mongo-driver v1.4.2
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.13.0
	golang.org/x/crypto v0.0.0-20210314154223-e6e6c4f2bb5b
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/sys v0.0.0-20210507014357-30e306a8bba5 // indirect
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1
	golang.org/x/tools v0.1.1-0.20210302220138-2ac05c832e1a // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/genproto v0.0.0-20200707001353-8e8330bf89df // indirect
	google.golang.org/grpc v1.37.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/cheggaaa/pb.v1 v1.0.28 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
)

go 1.16
