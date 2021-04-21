module github.com/raphasch/hotcertification

go 1.16

require (
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/niclabs/tcrsa v0.0.5
	github.com/relab/gorums v0.4.0
	github.com/relab/hotstuff v0.2.2
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	go.dedis.ch/kyber/v3 v3.0.13
	go.uber.org/zap v1.16.0 // indirect
	google.golang.org/grpc v1.36.1
	google.golang.org/protobuf v1.26.0
)

replace github.com/relab/hotstuff => ../hotstuff
