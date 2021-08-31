module github.com/pachyderm/pachyderm/v2

go 1.16

require (
	cloud.google.com/go v0.84.0
	cloud.google.com/go/storage v1.10.0
	github.com/Azure/azure-sdk-for-go v36.1.0+incompatible
	github.com/Azure/go-autorest/autorest/to v0.3.1-0.20191028180845-3492b2aff503 // indirect
	github.com/aws/aws-lambda-go v1.13.3
	github.com/aws/aws-sdk-go v1.27.0
	github.com/c-bata/go-prompt v0.2.3
	github.com/cevaris/ordered_map v0.0.0-20190319150403-3adeae072e73
	github.com/chmduquesne/rollinghash v4.0.0+incompatible
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.13+incompatible
	github.com/coreos/go-etcd v2.0.0+incompatible // indirect
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f
	github.com/cpuguy83/go-md2man v1.0.10 // indirect
	github.com/dexidp/dex v0.0.0-20210629090108-0780edbcbe43
	github.com/dexidp/dex/api/v2 v2.0.0
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/dlmiddlecote/sqlstats v1.0.2
	github.com/docker/go-units v0.4.0
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/elazarl/goproxy v0.0.0-20191011121108-aa519ddbe484 // indirect
	github.com/evanphx/json-patch v4.11.0+incompatible
	github.com/fatih/camelcase v1.0.0
	github.com/fatih/color v1.9.0
	github.com/fsouza/go-dockerclient v1.7.4
	github.com/go-openapi/validate v0.19.5 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/google/pprof v0.0.0-20190723021845-34ac40c74b70 // indirect
	github.com/gophercloud/gophercloud v0.1.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20191106031601-ce3c9ade29de // indirect
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.1-0.20191002090509-6af20e3a5340
	github.com/hanwen/go-fuse/v2 v2.0.3
	github.com/hashicorp/golang-lru v0.5.4
	github.com/itchyny/gojq v0.11.2
	github.com/jackc/pgconn v1.10.0
	github.com/jackc/pgerrcode v0.0.0-20201024163028-a0d42d470451
	github.com/jackc/pgx/v4 v4.13.0
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jmoiron/sqlx v1.2.0
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/json-iterator/go v1.1.11
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/juju/ansiterm v0.0.0-20180109212912-720a0952cc2a
	github.com/lib/pq v1.10.2
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-isatty v0.0.12
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/mattn/go-tty v0.0.3 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/minio/minio-go/v6 v6.0.55
	github.com/modern-go/reflect2 v1.0.1
	github.com/opentracing-contrib/go-grpc v0.0.0-20180928155321-4b5a12d3ff02
	github.com/opentracing/opentracing-go v1.1.1-0.20200124165624-2876d2018785
	github.com/pachyderm/ohmyglob v0.0.0-20210308211843-d5b47775fc36
	github.com/pachyderm/s2 v0.0.0-20200609183354-d52f35094520
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.9.1
	github.com/pkg/term v0.0.0-20190109203006-aa71e9d9e942 // indirect
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.26.0
	github.com/remyoudompheng/bigfft v0.0.0-20170806203942-52369c62f446 // indirect
	github.com/robfig/cron v1.2.0
	github.com/satori/go.uuid v1.2.0
	github.com/segmentio/analytics-go v0.0.0-20160426181448-2d840d861c32
	github.com/segmentio/backo-go v0.0.0-20160424052352-204274ad699c // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/smartystreets/assertions v1.0.1 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/uber/jaeger-client-go v2.20.1+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	github.com/ugorji/go/codec v0.0.0-20181204163529-d75b2dcb6bc8 // indirect
	github.com/vbauerster/mpb/v6 v6.0.2
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	go.uber.org/automaxprocs v1.4.0
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	golang.org/x/net v0.0.0-20210503060351-7fd8e65b6420
	golang.org/x/oauth2 v0.0.0-20210615190721-d04028783cf1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b
	google.golang.org/api v0.49.0
	google.golang.org/grpc v1.38.0
	gopkg.in/pachyderm/yaml.v3 v3.0.0-20200130061037-1dd3d7bd0850
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	k8s.io/klog v1.0.0 // indirect
	modernc.org/mathutil v1.0.0
	sigs.k8s.io/controller-runtime v0.9.6 // indirect
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0 // indirect
)

replace github.com/sercand/kuberesolver => github.com/sercand/kuberesolver v1.0.1-0.20200204133151-f60278fd3dac

// Dex pulls in a newer grpc and protobuf, but our etcd client can't work with the newer version.
// The following pin grpc, protobuf and everything else that would otherwise rely on the newer version.
// See https://github.com/etcd-io/etcd/pull/12000
replace google.golang.org/grpc => google.golang.org/grpc v1.29.1

replace github.com/golang/protobuf => github.com/golang/protobuf v1.3.5

replace cloud.google.com/go => cloud.google.com/go v0.49.0

replace cloud.google.com/go/storage => cloud.google.com/go/storage v1.10.0

//replace github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.5.0

//replace github.com/prometheus/common => github.com/prometheus/common v0.9.1

replace google.golang.org/genproto => google.golang.org/genproto v0.0.0-20191115194625-c23dd37a84c9

replace github.com/dexidp/dex => github.com/pachyderm/dex v0.0.0-20210811182333-56fc504b721f

//For controller runtime
replace github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1
