module github.com/raphasch/hotcertification

go 1.16

require (
	github.com/go-delve/delve v1.6.0 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.12
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/niclabs/tcrsa v0.0.5
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/peterh/liner v1.2.1 // indirect
	github.com/pkg/profile v0.0.0-20170413231811-06b906832ed0 // indirect
	github.com/relab/gorums v0.4.0
	github.com/relab/hotstuff v0.2.2
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday v1.6.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/cobra v1.1.3 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	go.dedis.ch/kyber/v3 v3.0.13
	go.starlark.net v0.0.0-20210416142453-1607a96e3d72 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/arch v0.0.0-20210422031329-8ee3ab241ede // indirect
	golang.org/x/sys v0.0.0-20210426080607-c94f62235c83 // indirect
	google.golang.org/grpc v1.36.1
	google.golang.org/protobuf v1.26.0
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
)

replace github.com/relab/hotstuff => ../hotstuff
