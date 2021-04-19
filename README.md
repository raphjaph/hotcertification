# TODOs and Documentation

## TODO

- [] add [database](https://github.com/gostor/awesome-go-storage) or
  [interface](https://github.com/philippgille/gokv) to database or check out
  [sqlite](https://www.sqlite.org/index.html)
- [] add a Makefile, see [here](https://makefiletutorial.com/) for help
- [] make event-driven architecture? -> more lightweight; look at Flow implementation
- [] instead of ClientID and Sequence Number just use Hash of CSR to identify request (replaces CMDID data structure) -> use HashMap
- [] slow down consensus rounds?
- [] go run with SIGNALS

## Logging and Configuration

- [] find out how to show logs from internal consensus protocol and client facing
server; pipe HS log into my logger
- [] change hotstuff.toml to hotstuff.yml see [here](https://stackoverflow.com/questions/33989612/yaml-equivalent-of-array-of-objects-in-json)
- [x] merge internal/cli into main


## Crypto/Threshold library

- [] define curve parameters somewhere (CURVE, G, N) [wiki](https://en.wikipedia.org/wiki/Elliptic_Curve_Digital_Signature_Algorithm)
- [] recycle PartialCert from Hotstuff for Certification Service crypto
- [] add error checks to threshold.go

## Miscellaneous

- other cool go library [repos](https://github.com/avelino/awesome-go)

## Protocol Buffers (protoc)

- execute compile command in proto folder
- ` protoc -I=/Users/raphael/.go/pkg/mod/github.com/relab/gorums@v0.3.0. --go_out=paths=source_relative:. --gorums_out=paths=source_relative:. certification.proto `
- this command finds the absolute path with versioning for gorums package
- ` go list -m -f {{.Dir}} github.com/relab/gorums `

