module github.com/apache/servicecomb-service-center

replace github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b => github.com/go-chassis/glog v0.0.0-20180920075250-95a09b2413e9

require (
	github.com/NYTimes/gziphandler v1.0.2-0.20180820182813-253f1acb9d9f
	github.com/Shopify/sarama v1.18.0 // indirect
	github.com/Shopify/toxiproxy v2.1.4+incompatible // indirect
	github.com/apache/thrift v0.0.0-20180125231006-3d556248a8b9 // indirect
	github.com/astaxie/beego v1.8.0
	github.com/cheggaaa/pb v1.0.25
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.22+incompatible
	github.com/coreos/go-semver v0.2.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20180828140353-eee3db372b31 // indirect
	github.com/coreos/pkg v0.0.0-20180108230652-97fdf19511ea // v4
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/eapache/go-resiliency v1.1.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fzipp/gocyclo v0.0.0-20150627053110-6acd4345c835 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-chassis/foundation v0.1.1-0.20191113114104-2b05871e9ec4
	github.com/go-chassis/go-archaius v1.3.2
	github.com/go-chassis/go-chassis v0.0.0-20200716073336-4183d5dd1246
	github.com/go-chassis/paas-lager v1.1.1
	github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/golang/protobuf v1.3.2
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gorilla/websocket v1.4.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.3.0 // indirect
	github.com/hashicorp/mdns v1.0.1 // indirect
	github.com/hashicorp/serf v0.8.3
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/karlseguin/ccache v2.0.3-0.20170217060820-3ba9789cfd2c+incompatible
	github.com/labstack/echo v3.2.2-0.20180316170059-a5d81b8d4a62+incompatible
	github.com/natefinch/lumberjack v0.0.0-20170531160350-a96e63847dc3
	github.com/olekukonko/tablewriter v0.0.0-20180506121414-d4647c9c7a84
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin/zipkin-go-opentracing v0.3.3-0.20180123190626-6bb822a7f15f
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pierrec/lz4 v2.0.5+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.1
	github.com/prometheus/client_model v0.0.0-20190115171406-56726106282f
	github.com/prometheus/procfs v0.0.0-20190117184657-bf6a532e95b1
	github.com/rcrowley/go-metrics v0.0.0-20180503174638-e2704e165165 // indirect
	github.com/rs/cors v0.0.0-20170608165155-8dd4211afb5d // v1.1
	github.com/satori/go.uuid v1.1.0
	github.com/smartystreets/assertions v0.0.0-20190116191733-b6c0e53d7304 // indirect
	github.com/soheilhy/cmux v0.1.4 // indirect
	github.com/spf13/cobra v0.0.0-20170624150100-4d647c8944eb
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	github.com/widuu/gojson v0.0.0-20170212122013-7da9d2cd949b
	github.com/wsxiaoys/terminal v0.0.0-20160513160801-0940f3fc43a0 // indirect
	github.com/xiang90/probing v0.0.0-20160813154853-07dd2e8dfe18 // indirect
	go.etcd.io/bbolt v1.3.4 // indirect
	go.etcd.io/etcd v3.3.22+incompatible
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20200221231518-2aa609cf4a9d
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	google.golang.org/grpc v1.19.0
	gopkg.in/cheggaaa/pb.v1 v1.0.28 // indirect
	gopkg.in/karlseguin/expect.v1 v1.0.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6 // indirect
)

go 1.13
