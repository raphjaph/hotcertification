# HotCertification: A distributed Certificate Authority

A byzantine fault-tolerant state machine replication algorithm and threshold signature scheme that in conjunction perform the basic functionalities of a Certificate Authority (CA).
It adds these layers of complexity in order to be more resilient against process failures and malicious attacks.
The certification process can be abstracted to sign any piece of data (not just a X509 Certificate) like access tokens (macaroons) or JWTs and through that act as a sort of Authentication Server.
 
## Using it
This is really just a prototype I built as part of my Bachelor's thesis so it still has bugs and major refactoring.
An example configuration of a cluster of four HotCertification nodes can be found in `hotcertification.toml`.
Compile the binaries by calling `make`.
Create the cryptographic material like private keys and TLS certificates with:
```bash
mkdir keys
./cmd/keygen/keygen -n $num_nodes -t $threshold --key-size 512 keys
```
Run the cluster of four nodes locally by executing `run_servers_localhost.sh`.
Test the cluster with an example client with:
```bash
./cmd/client/client client.crt
```









++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
## TODO

- [] rename coordinator struct
- [x] merge/put in same dir coordinator.go and options.go (Because they outline how hotcertification works)
- [] add [database](https://github.com/gostor/awesome-go-storage) or
  [interface](https://github.com/philippgille/gokv) to database or check out
  [sqlite](https://www.sqlite.org/index.html)
- [x] add a Makefile, see [here](https://makefiletutorial.com/) for help
- [] make event-driven architecture? -> more lightweight; look at Flow implementation
- [] instead of ClientID and Sequence Number just use Hash of CSR to identify request (replaces CMDID data structure) -> use HashMap
- [] slow down consensus rounds?
- [x] go run with SIGNALS

## Testing

- https://github.com/alexei-led/pumba

## Logging and Configuration

- [] add a custom level to the log -> APPLICATION/CERTIFICATION; refactor code accordingly
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

