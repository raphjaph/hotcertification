module github.com/raphasch/hotcertification

go 1.16

require (
	github.com/mattn/go-isatty v0.0.12
	github.com/niclabs/tcrsa v0.0.5
	github.com/relab/gorums v0.5.0
	github.com/relab/hotstuff v0.2.2
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	go.uber.org/zap v1.16.0
	google.golang.org/grpc v1.37.0
	google.golang.org/protobuf v1.26.0
)

replace github.com/relab/hotstuff v0.2.2 => ./hotstuff
