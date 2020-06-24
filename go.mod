module github.com/ghetzel/go-webfriend

go 1.12

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
	github.com/ghetzel/diecast v1.18.3
	github.com/ghetzel/friendscript v0.7.5
	github.com/ghetzel/go-defaults v1.2.0
	github.com/ghetzel/go-stockutil v1.8.72
	github.com/ghetzel/testify v1.4.1
	github.com/gobwas/glob v0.2.3
	github.com/gorilla/websocket v1.4.2
	github.com/husobee/vestigo v1.1.0
	github.com/jdxcode/netrc v0.0.0-20180207092346-e1a19c977509
	github.com/mafredri/cdp v0.28.0
	github.com/mitchellh/go-ps v0.0.0-20170309133038-4fdf99ab2936
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/urfave/negroni v1.0.1-0.20191011213438-f4316798d5d3
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
)

// replace github.com/ghetzel/friendscript v0.7.5 => ../friendscript
