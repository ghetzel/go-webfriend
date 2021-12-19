module github.com/ghetzel/go-webfriend

go 1.17

replace (
	k8s.io/api => k8s.io/api v0.19.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.1
	k8s.io/client-go => k8s.io/client-go v0.19.1
)

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.7.0
	github.com/ghetzel/argonaut v0.0.0-20180428155514-51604c68ce30
	github.com/ghetzel/cli v1.17.0
	github.com/ghetzel/diecast v1.21.0
	github.com/ghetzel/friendscript v0.8.4
	github.com/ghetzel/go-defaults v1.2.0
	github.com/ghetzel/go-stockutil v1.10.3
	github.com/ghetzel/testify v1.4.1
	github.com/gobwas/glob v0.2.3
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.3 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/husobee/vestigo v1.1.0
	github.com/jdxcode/netrc v0.0.0-20201119100258-050cafb6dbe6
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/mafredri/cdp v0.28.0
	github.com/mitchellh/go-ps v0.0.0-20170309133038-4fdf99ab2936
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/urfave/negroni v1.0.1-0.20191011213438-f4316798d5d3
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	k8s.io/api v0.19.1
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009 // indirect
)

require (
	cloud.google.com/go v0.54.0 // indirect
	github.com/PuerkitoBio/goquery v1.5.0 // indirect
	github.com/alecthomas/chroma v0.7.3 // indirect
	github.com/andybalholm/cascadia v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.36.24 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/beevik/etree v1.1.0 // indirect
	github.com/biessek/golang-ico v0.0.0-20180326222316-d348d9ea4670 // indirect
	github.com/c-bata/go-prompt v0.2.2 // indirect
	github.com/cenkalti/backoff v2.1.1+incompatible // indirect
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/dsnet/compress v0.0.0-20171208185109-cc9eb1d7ad76 // indirect
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/fatih/structs v1.0.0 // indirect
	github.com/ghetzel/ratelimit v0.0.0-20200513232932-b28727c55ae1 // indirect
	github.com/ghetzel/uuid v0.0.0-20171129191014-dec09d789f3d // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/logr v0.2.0 // indirect
	github.com/go-shiori/dom v0.0.0-20200325044552-dcb2bfb8d4d8 // indirect
	github.com/go-shiori/go-readability v0.0.0-20200413080041-05caea5f6592 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/gomodule/redigo v2.0.0+incompatible // indirect
	github.com/gorilla/css v1.0.0 // indirect
	github.com/grandcat/zeroconf v0.0.0-20190118114326-c2d1b4121200 // indirect
	github.com/grokify/html-strip-tags-go v0.0.0-20180530080503-3f8856873ce5 // indirect
	github.com/h2non/filetype v1.0.13-0.20200520201155-df519de6e270 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.0.0 // indirect
	github.com/imdario/mergo v0.3.5 // indirect
	github.com/jackpal/gateway v1.0.5-0.20180407163008-cbcf4e3f3bae // indirect
	github.com/jbenet/go-base58 v0.0.0-20150317085156-6237cf65f3a6 // indirect
	github.com/jdkato/prose v1.1.0 // indirect
	github.com/jlaffaye/ftp v0.0.0-20190126081051-8019e6774408 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jsummers/gobmp v0.0.0-20151104160322-e2ba15ffa76e // indirect
	github.com/jszwec/s3fs v0.3.1 // indirect
	github.com/juliangruber/go-intersect v1.0.0 // indirect
	github.com/kellydunn/golang-geo v0.7.0 // indirect
	github.com/kelvins/sunrisesunset v0.0.0-20170601204625-14f1915ad4b4 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/kylelemons/go-gypsy v0.0.0-20160905020020-08cad365cd28 // indirect
	github.com/kyokomi/emoji v2.1.0+incompatible // indirect
	github.com/lib/pq v1.1.0 // indirect
	github.com/martinlindhe/unit v0.0.0-20190604142932-3b6be53d49af // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/mattn/go-shellwords v1.0.9 // indirect
	github.com/mattn/go-tty v0.0.0-20180219170247-931426f7535a // indirect
	github.com/mcuadros/go-defaults v1.1.0 // indirect
	github.com/melbahja/goph v1.2.1 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/microcosm-cc/bluemonday v1.0.16 // indirect
	github.com/miekg/dns v1.1.43 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/montanaflynn/stats v0.0.0-20151014174947-eeaced052adb // indirect
	github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852 // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7 // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.12.0 // indirect
	github.com/pkg/term v0.0.0-20180730021639-bffc007b7fd5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/signalsciences/tlstext v1.2.0 // indirect
	github.com/sirupsen/logrus v1.5.0 // indirect
	github.com/sj14/astral v0.1.1 // indirect
	github.com/spaolacci/murmur3 v0.0.0-20170819071325-9f5d223c6079 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/tg123/go-htpasswd v0.0.0-20190305225429-d38e564730bf // indirect
	github.com/uber/jaeger-client-go v2.24.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	github.com/yosssi/gohtml v0.0.0-20180130040904-97fbf36f4aa8 // indirect
	github.com/ziutek/mymysql v1.5.4 // indirect
	go.uber.org/atomic v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/net v0.0.0-20211015210444-4f30a5c0130f // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20211019181941-9d821ace8654 // indirect
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.8 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/neurosnap/sentences.v1 v1.0.6 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/klog/v2 v2.2.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.0.1 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)
