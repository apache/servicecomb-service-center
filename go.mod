module github.com/apache/servicecomb-service-center

replace (
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b => github.com/go-chassis/glog v0.0.0-20180920075250-95a09b2413e9
	github.com/apache/thrift => github.com/apache/thrift v0.0.0-20180125231006-3d556248a8b9
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/astaxie/beego v1.12.2
	github.com/cheggaaa/pb v1.0.25
	github.com/coreos/etcd v3.3.22+incompatible
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // v4
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/elithrar/simple-scrypt v1.3.0
	github.com/go-chassis/cari v0.3.0
	github.com/go-chassis/foundation v0.3.0
	github.com/go-chassis/go-archaius v1.3.2
	github.com/go-chassis/go-chassis v0.0.0-20200826064053-d90be848aa10
	github.com/go-chassis/paas-lager v1.1.1
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.4.3
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/serf v0.8.3
	github.com/karlseguin/ccache v2.0.3-0.20170217060820-3ba9789cfd2c+incompatible
	github.com/labstack/echo v3.2.2-0.20180316170059-a5d81b8d4a62+incompatible
	github.com/labstack/echo/v4 v4.1.17
	github.com/natefinch/lumberjack v0.0.0-20170531160350-a96e63847dc3
	github.com/olekukonko/tablewriter v0.0.0-20180506121414-d4647c9c7a84
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin/zipkin-go-opentracing v0.3.3-0.20180123190626-6bb822a7f15f
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.8.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/procfs v0.2.0
	github.com/rs/cors v1.7.0
	github.com/satori/go.uuid v1.1.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/widuu/gojson v0.0.0-20170212122013-7da9d2cd949b
	go.uber.org/zap v1.13.0
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	google.golang.org/grpc v1.27.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.5
	k8s.io/apimachinery v0.19.5
	k8s.io/client-go v0.19.5
)

go 1.13
