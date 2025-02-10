# `nagare media engine`

[![Go Report Card](https://goreportcard.com/badge/github.com/nagare-media/engine?style=flat-square)](https://goreportcard.com/report/github.com/nagare-media/engine)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.23-61CFDD.svg?style=flat-square)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/nagare-media/engine)](https://pkg.go.dev/github.com/nagare-media/engine)

`nagare media engine` is a prototypical implementation of [ISO/IEC 23090-8:2020 Network Based Media Processing (NBMP)](https://www.iso.org/standard/77839.html) coming from research. MPEG published NBMP in order to meet today's multimedia workflows. It defines data models, APIs and a reference architecture for network-distributed multimedia workflows. `nagare media engine` builds upon the Kubernetes platform to implement NBMP as a modern open source cloud native solution. Our goal is to extend `nagare media engine` to a production ready implementation and simultaneously innovate on new ideas.

## Documentation

For further documents, see [the `docs` folder](./docs/README.md).

## Quick Start

Install the following tools as prerequisites:

* Docker
* `git`
* `sh` (UNIX shell)
* `bash`
* `curl`
* `make`
* `kind`
* `kubectl`
* `helm`

You can test out `nagare media engine` in a local `kind` cluster by executing the following commands:

```sh
# create kind cluster
$ make kind-up

# deploy nagare media engine
$ make skaffold-run

# ... test nagare media engine ...

# destroy kind cluster
$ make kind-down
```

## License

Apache 2.0 (c) nagare media authors
