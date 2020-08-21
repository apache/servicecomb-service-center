module github.com/apache/servicecomb-service-center

replace github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b => github.com/go-chassis/glog v0.0.0-20180920075250-95a09b2413e9

require (
	github.com/NYTimes/gziphandler v1.0.2-0.20180820182813-253f1acb9d9f
	github.com/astaxie/beego v1.8.0
	github.com/cheggaaa/pb v1.0.25
	github.com/coreos/etcd v3.3.22+incompatible
	github.com/coreos/pkg v0.0.0-20180108230652-97fdf19511ea // v4
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-chassis/foundation v0.1.1-0.20191113114104-2b05871e9ec4
	github.com/go-chassis/go-archaius v1.3.2
	github.com/go-chassis/go-chassis v0.0.0-20200820073047-a1c634b7bdd1
	github.com/go-chassis/paas-lager v1.1.1
	github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d
	github.com/golang/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.0
	github.com/hashicorp/serf v0.8.3
	github.com/karlseguin/ccache v2.0.3-0.20170217060820-3ba9789cfd2c+incompatible
	github.com/labstack/echo v3.2.2-0.20180316170059-a5d81b8d4a62+incompatible
	github.com/natefinch/lumberjack v0.0.0-20170531160350-a96e63847dc3
	github.com/olekukonko/tablewriter v0.0.0-20180506121414-d4647c9c7a84
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin/zipkin-go-opentracing v0.3.3-0.20180123190626-6bb822a7f15f
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.1
	github.com/prometheus/client_model v0.0.0-20190115171406-56726106282f
	github.com/prometheus/procfs v0.0.0-20190117184657-bf6a532e95b1
	github.com/rs/cors v0.0.0-20170608165155-8dd4211afb5d // v1.1
	github.com/satori/go.uuid v1.1.0
	github.com/spf13/cobra v0.0.0-20170624150100-4d647c8944eb
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	github.com/widuu/gojson v0.0.0-20170212122013-7da9d2cd949b
	go.etcd.io/etcd v3.3.22+incompatible
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20190820162420-60c769a6c586
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	google.golang.org/grpc v1.19.0
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
)

go 1.13
