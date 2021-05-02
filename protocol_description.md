
# The HotCertification Protocol

## System Overview

The key idea behind the novel certification process is having **no single point of failure** by distributing
both the validation and signing of a certificate. From the client's perspective it looks like its
interacting with a monolithic system but internally (and from an adversaries perspective) we have a cluster
of nodes. By combining the validation and signing process into one node we hope to have higher cohesion than
the Hyperledger Fabric implementation and through that better performance.

![system-overview](/Users/raphael/bachelor_arbeit/lightweight-consensus/diagram_files/system_overview.png)

## A Client's View

For client-to-server communication this protocol uses gRPC. It has many useful
features, one of them being that the client does not have be written in the same
language as the server. This means that the client implementation can be
integrated easily with any existing code base.  Another feature is the
simplification of the control flow because gRPC handles most of the networking
stack including (de)serialization. Here is the gRPC interface definition (API)
for getting a certificate from a certificate signing request (CSR).

```gRPC
service Certification {
    rpc GetCertificate(CSR) returns (Certificate) {}
}

message CSR {
    uint32 ClientID = 1;
    bytes CertificateRequest = 2;
    bytes ValidationInfo = 3;
}

message Certificate {
    bytes Certificate = 1;
}

```

The server exposes only the `GetCertificate()` function and there are only two
types of messages (`CSR` and `Certificate`) which means that to the client it
looks like it's interacting with a normal Certification Authority (CA). When a
client wishes to request a new certificate it can use any language's X.509
library to create a `CertificateRequest`, which should be an ASN.1 DER encoded
byte array. Using gRPC (in combination with Protocol Buffers) the client then
wraps its `CertificateRequest` into the `CSR` message type and adds its
`ClientID` and `ValidationInfo` (tbd). The message is then sent to the server.

The server then validates, signs and creates a certificate with the other nodes
of the HotCertification cluster/configuration. The server that received the
client's `CSR` then returns the ASN.1 DER encoded `Certificate` byte array,
which can easily be parsed by the client's X.509 library.

## A Node's View

From a very high-level a HotCertification node has three processes that run
independently and continuously: `ClientServer`, `Replication` and `Signing`.
There is of course other processes in the background that coordinates all these
asynchronously running processes but they are omitted here for simplicity.

![high-level view](/Users/raphael/bachelor_arbeit/lightweight-consensus/diagram_files/high-level-overview.svg)

`ClientServer`: Handles clients' requests and tags incoming `CSR` messages
before passing them on the the replication process. It receives a X.509
`Certificate` from the signing process and returns it to the client.

`Replication`: Since a client's CSR enters the cluster at one node all the other
nodes have to be made aware of this request. The goal is to replicate one node's
state to all other nodes in an extremely tamper-resistent way. The technical term
is byzantine fault-tolerant state-machine replication(BFT-SMR). To achieve this
goal the system uses a protocol called HotStuff (that's the reason for the name
HotCertification). The replication process also validates the `CSR` before
passing it on the the signing process.

`Signing`: The input to the signing process is a validated and replicated `CSR`,
which means that the signing process knows that the other nodes have the same
state as itself. It can now proceed to partially sign the certificate with its
threshold key. If the signing process is part of the node that handles the
client request then it collects the partial signatures from the other nodes and
computes the full `SHA256WithRSA` signature. Finally it passes the ASN.1 DER
encoded `Certificate` to the client server, which then returns it to the client.

The message types are the same as defined above only that the client server tags the
`CSR` with additional fields to keep track of its status while it travels
through the system.
