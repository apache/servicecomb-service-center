module github.com/apache/servicecomb-service-center

replace google.golang.org/grpc => google.golang.org/grpc v1.19.0

require (
	github.com/NYTimes/gziphandler v1.0.2-0.20180820182813-253f1acb9d9f
	github.com/Shopify/sarama v1.27.2 // indirect
	github.com/apache/thrift v0.0.0-20180125231006-3d556248a8b9 // indirect
	github.com/astaxie/beego v1.8.0
	github.com/cheggaaa/pb v1.0.25
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.22+incompatible
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20180828140353-eee3db372b31 // indirect
	github.com/coreos/pkg v0.0.0-20180108230652-97fdf19511ea // v4
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/go-chassis/foundation v0.1.1-0.20200825060850-b16bf420f7b3
	github.com/go-chassis/go-archaius v1.3.6-0.20201103103813-43dd1680ebfb
	github.com/go-chassis/go-chassis/v2 v2.0.5-0.20201114043352-d6ab1f0a882d
	github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d
	github.com/golang/protobuf v1.4.2
	github.com/gorilla/websocket v1.4.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hashicorp/serf v0.8.3
	github.com/iancoleman/strcase v0.1.2
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/karlseguin/ccache v2.0.3-0.20170217060820-3ba9789cfd2c+incompatible
	github.com/labstack/echo v3.2.2-0.20180316170059-a5d81b8d4a62+incompatible
	github.com/labstack/echo/v4 v4.1.17
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/natefinch/lumberjack v0.0.0-20170531160350-a96e63847dc3
	github.com/olekukonko/tablewriter v0.0.0-20180506121414-d4647c9c7a84
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.1
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin/zipkin-go-opentracing v0.3.3-0.20180123190626-6bb822a7f15f
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v0.9.1
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4
	github.com/prometheus/procfs v0.0.0-20190117184657-bf6a532e95b1
	github.com/rs/cors v0.0.0-20170608165155-8dd4211afb5d // v1.1
	github.com/satori/go.uuid v1.1.0
	github.com/soheilhy/cmux v0.1.4 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/tmc/grpc-websocket-proxy v0.0.0-20200427203606-3cfed13b9966 // indirect
	github.com/urfave/cli v1.22.4
	github.com/widuu/gojson v0.0.0-20170212122013-7da9d2cd949b
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.mongodb.org/mongo-driver v1.4.2
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	google.golang.org/grpc v1.33.1
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
)

go 1.13
