# TODOs and Documentation

## TODO

- [] find out how to show logs from internal consensus protocol and client facing
  server; pipe HS log into my logger
- [] add [database](https://github.com/gostor/awesome-go-storage) or
  [interface](https://github.com/philippgille/gokv) to database or check out
  [sqlite](https://www.sqlite.org/index.html)
- [] add a Makefile, see [here](https://makefiletutorial.com/) for help
- [] change hotstuff.toml to hotstuff.yml see [here](https://stackoverflow.com/questions/33989612/yaml-equivalent-of-array-of-objects-in-json)
- [] make event-driven architecture? -> more lightweight; look at Flow implementation
- [] recycle PartialCert from Hotstuff for Certification Service crypto
- [] instead of ClientID and Sequence Number just use Hash of CSR to identify request (replaces CMDID data structure) -> use HashMap
- [] slow down consensus rounds?
- [] merge internal/cli into main
- [] go run with SIGNALS
- [] add error checks to threshold.go

## Miscellaneous

- other cool go library [repos](https://github.com/avelino/awesome-go)

## Protocol Buffers (protoc)

- execute compile command in proto folder
- ` protoc -I=/Users/raphael/.go/pkg/mod/github.com/relab/gorums@v0.2.3-0.20210213125733-f04667f97266. --go_out=paths=source_relative:. --gorums_out=paths=source_relative:. certification.proto `
- this command finds the absolute path with versioning for gorums package
- ` go list -m -f {{.Dir}} github.com/relab/gorums `

