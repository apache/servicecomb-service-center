module github.com/apache/servicecomb-service-center

replace (
	github.com/apache/servicecomb-service-center/api => ./api
	github.com/apache/servicecomb-service-center/eventbase => ./eventbase
	// github.com/openzipkin/zipkin-go-opentracing depends on
	github.com/apache/thrift => github.com/apache/thrift v0.0.0-20180125231006-3d556248a8b9
	github.com/go-logr/logr => github.com/go-logr/logr v0.2.0
	// beego depends on
	go.opentelemetry.io/otel => go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/sdk => go.opentelemetry.io/otel/sdk v0.20.0
	go.opentelemetry.io/otel/trace => go.opentelemetry.io/otel/trace v0.20.0
)

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/apache/servicecomb-kie v0.2.1-0.20230916085640-4d917719ec79
	github.com/apache/servicecomb-service-center/eventbase v0.0.0-20220120070230-26997eb876ca
	github.com/beego/beego/v2 v2.0.4
	github.com/cheggaaa/pb v1.0.25
	github.com/cloudflare/gokey v0.1.2
	github.com/deckarep/golang-set v1.8.0
	github.com/elithrar/simple-scrypt v1.3.0
	github.com/go-chassis/cari v0.9.0
	github.com/go-chassis/foundation v0.4.0
	github.com/go-chassis/go-archaius v1.5.6
	github.com/go-chassis/go-chassis-extension/codec/gojson v0.0.0-20220816060440-fe98a5615d3f
	github.com/go-chassis/go-chassis-extension/protocol/grpc v0.0.0-20220208081606-003611df45da
	github.com/go-chassis/go-chassis/v2 v2.7.1
	github.com/go-chassis/kie-client v0.2.1-0.20230916082929-a48f84588280
	github.com/go-chassis/openlog v1.1.3
	github.com/gofiber/fiber/v2 v2.36.0
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/gorilla/websocket v1.4.3-0.20210424162022-e8629af678b7
	github.com/iancoleman/strcase v0.2.0
	github.com/jinzhu/copier v0.3.5
	github.com/karlseguin/ccache v2.0.3-0.20170217060820-3ba9789cfd2c+incompatible
	github.com/labstack/echo/v4 v4.9.0
	github.com/little-cui/etcdadpt v0.4.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	github.com/opentracing/opentracing-go v1.2.0
	github.com/openzipkin/zipkin-go-opentracing v0.3.3-0.20180123190626-6bb822a7f15f
	github.com/orcaman/concurrent-map v1.0.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.13.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/procfs v0.8.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/rs/cors v1.7.0 // v1.1
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.0
	github.com/urfave/cli v1.22.4
	github.com/widuu/gojson v0.0.0-20170212122013-7da9d2cd949b
	go.etcd.io/etcd/api/v3 v3.5.4
	go.etcd.io/etcd/client/v3 v3.5.4
	go.mongodb.org/mongo-driver v1.5.1
	go.uber.org/zap v1.20.0
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	google.golang.org/grpc v1.46.0
	google.golang.org/protobuf v1.28.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.19.5
	k8s.io/apimachinery v0.19.5
	k8s.io/client-go v0.19.5
	k8s.io/kube-openapi v0.0.0-20210527164424-3c818078ee3d
)

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-chassis/go-chassis-extension/protocol/fiber4r v0.0.0-20220825091211-99d5e9810fd7
)

require (
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/Shopify/sarama v1.19.0 // indirect
	github.com/Shopify/toxiproxy v2.1.4+incompatible // indirect
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/apache/thrift v0.13.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a // indirect
	github.com/aws/aws-sdk-go v1.34.28 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/eapache/go-resiliency v1.1.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/emicklei/go-restful v2.16.0+incompatible // indirect
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-chassis/go-restful-swagger20 v1.0.4-0.20220704025524-9243cbee26b7 // indirect
	github.com/go-chassis/sc-client v0.6.1-0.20220728072125-dacdd0c834bf // indirect
	github.com/go-chassis/seclog v1.3.1-0.20210917082355-52c40864f240 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.3 // indirect
	github.com/go-openapi/jsonreference v0.19.3 // indirect
	github.com/go-openapi/swag v0.19.5 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/goccy/go-json v0.9.10 // indirect
	github.com/gofiber/adaptor/v2 v2.1.25 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.5 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/karlseguin/ccache/v2 v2.0.8 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/labstack/gommon v0.3.1 // indirect
	github.com/mailru/easyjson v0.0.0-20190626092158-b2ccc519800e // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nxadm/tail v1.4.4 // indirect
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/pierrec/lz4 v2.0.5+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18 // indirect
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20201229170055-e5319fda7802 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.38.0 // indirect
	github.com/valyala/fasttemplate v1.2.1 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.0.2 // indirect
	github.com/xdg-go/stringprep v1.0.2 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.4 // indirect
	go.etcd.io/etcd/client/v2 v2.305.4 // indirect
	go.etcd.io/etcd/pkg/v3 v3.5.4 // indirect
	go.etcd.io/etcd/raft/v3 v3.5.4 // indirect
	go.etcd.io/etcd/server/v3 v3.5.4 // indirect
	go.opentelemetry.io/contrib v0.20.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.20.0 // indirect
	go.opentelemetry.io/otel v1.7.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp v0.20.0 // indirect
	go.opentelemetry.io/otel/metric v0.20.0 // indirect
	go.opentelemetry.io/otel/sdk v1.7.0 // indirect
	go.opentelemetry.io/otel/sdk/export/metric v0.20.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.20.0 // indirect
	go.opentelemetry.io/otel/trace v1.7.0 // indirect
	go.opentelemetry.io/proto/otlp v0.7.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/oauth2 v0.0.0-20220223155221-ee480838109b // indirect
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f // indirect
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa // indirect
	gopkg.in/cheggaaa/pb.v1 v1.0.28 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.2.0 // indirect
	k8s.io/utils v0.0.0-20200729134348-d5654de09c73 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.0.2 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

go 1.18
