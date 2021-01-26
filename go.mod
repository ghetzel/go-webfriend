module github.com/ghetzel/go-webfriend

go 1.12

require (
	cloud.google.com/go v0.51.0 // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.7.0
	github.com/ghetzel/argonaut v0.0.0-20180428155514-51604c68ce30
	github.com/ghetzel/cli v1.17.0
	github.com/ghetzel/diecast v1.20.0
	github.com/ghetzel/friendscript v0.8.4
	github.com/ghetzel/go-defaults v1.2.0
	github.com/ghetzel/go-stockutil v1.9.5
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
	k8s.io/api v0.16.15
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009 // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.16.15
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.16-rc.0
	k8s.io/client-go => k8s.io/client-go v0.16.15
)
