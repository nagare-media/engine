# `nagare media engine`

[![Go Report Card](https://goreportcard.com/badge/github.com/nagare-media/engine?style=flat-square)](https://goreportcard.com/report/github.com/nagare-media/engine)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.23-61CFDD.svg?style=flat-square)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/nagare-media/engine)](https://pkg.go.dev/github.com/nagare-media/engine)

`nagare media engine` is a prototypical implementation of [ISO/IEC 23090-8:2020 Network Based Media Processing (NBMP)](https://www.iso.org/standard/77839.html) coming from research. MPEG published NBMP in order to meet today's multimedia workflows. It defines data models, APIs and a reference architecture for network-distributed multimedia workflows. `nagare media engine` builds upon the Kubernetes platform to implement NBMP as a modern open source cloud native solution. Our goal is to extend `nagare media engine` to a production ready implementation and simultaneously innovate on new ideas.

Original repo: <https://github.com/nagare-media/engine>

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
* `skaffold`

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

## Components

![Architecture](./docs/architecture.drawio.svg)

`nagare media engine` consists of multiple components as depicted above. All components are assumed to be executed within Kubernetes clusters and take advantage of Kubernetes primitives.

### NBMP Gateway

The NBMP Gateway implements the NBMP Workflow API and translates to and from the NBMP and Kubernetes data models.

### Workflow Manager

This is the central component and implements a NBMP workflow manager as a collection of [Kubernetes controllers](https://kubernetes.io/docs/concepts/architecture/controller/). NBMP concepts are reflected as custom resources in Kubernetes that are reconciled with the current cluster state.

### Workflow Manager Helper

The Workflow Manager Helper is executed once for each NBMP Task and implements the necessary interactions with the NBMP Task API. It also provides a report API for media functions. All reported events are stored in a [NATS](https://nats.io/) [JetStream](https://docs.nats.io/nats-concepts/jetstream).

### Task Shim

The Task Shim provides a generic implementation of the NBMP Task API. This allows media functions to focus on the core logic instead of dealing with API interactions.

### Function

Several media functions are implemented as prototype in `nagare media engine`:

* **data-copy**: The input stream is copied to output streams without any modifications. Can be used as fan-out.
* **data-discard**: The input stream is discarded.
* **generic-noop**: This function does nothing.
* **generic-sleep**: This function sleeps for the configured amount.
* **media-encode**: The input stream is encoded and send to output streams.
* **media-generate-testpattern**: A test stream is generated and send to output streams.
* **script-lua**: A Lua script is executed that can self-adapt the currently running workflow.

## Development

The included `Makefile` provides helpful functions to ease development.

`make generate` will automatically generate code and configuration files, e.g. after updating the custom resource definitions.

`make kind-up` and `make kind-down` will start and stop two [kind](https://kind.sigs.k8s.io/) clusters already provisioned with all necessary dependencies ready for local testing.

`make skaffold-run` and `make skaffold-debug` use [Skaffold](https://skaffold.dev/) to build and deploy a development version of `nagare media engine` to the previously provisioned kind clusters.

## Building

`nagare media engine` does not depend on specific operating systems or architectures and should compile to many of Go's compile targets. However, since all components run within Kubernetes clusters, it mainly focuses on `linux`. With the included `Makefile`, it is easy to build binaries or container images for various platforms:

```sh
# build binaries / container images for local system architecture
$ make build
$ make image

# build binaries / container images for linux/amd64
$ make build OS=linux ARCH=amd64
$ make image OS=linux ARCH=amd64

# build workflow-manager binary / container image for linux/amd64
$ make build-workflow-manager OS=linux ARCH=amd64
$ make image-workflow-manager OS=linux ARCH=amd64

# build multi-platform container images pushed to a container registry
$ make image IMAGE_PLATFORMS=linux/amd64,linux/arm64 BUILDX_OUTPUT="--push"

# build container images with custom registry and tag
$ make image IMAGE_REGISTRY=my-reg.local IMAGE_TAG=my-version
```

There are also `make` targets to generate YAML manifests:

```sh
# generate CRDs as YAML
$ make output-crds

# generate YAML resources to deploy nagare media engine
$ make output-deployment
```

Finally, the `clean` target resets the build environment:

```sh
# delete all binaries, YAML manifests and temporary files
$ make clean
```

## Documentation

For further documents, see [the `docs` folder](./docs/README.md).

## Publications

Parts of this software were presented at various conferences. Please cite this project in your research projects as follows:

> Matthias Neugebauer. 2023. Nagare Media Engine: Towards an Open-Source Cloud- and Edge-Native NBMP Implementation. In *International Conference on Software Technologies (ICSOFT ’23), July 10-12, 2023, Rome, Italy.* SciTePress, Setúbal, Portugal, 8 pages. <https://doi.org/10.5220/0012087200003538>

> Matthias Neugebauer. 2024. Nagare Media Engine: Task Error Recovery in MPEG NBMP Workflows Through Event Sourcing. In *ACM Multimedia Systems Conference 2024 (MMSys ’24), April 15-18, 2024, Bari, Italy.* ACM, New York, NY, USA, 7 pages. <https://doi.org/10.1145/3625468.3652167>

> Matthias Neugebauer. 2025. Nagare Media Engine: Towards Self-Adapting MPEG NBMP Multimedia Workflows. In *ACM Multimedia Systems Conference 2025 (MMSys ’25), March 31-April 4, 2025, Stellenbosch, South Africa.* ACM, New York, NY, USA, 7 pages. <https://doi.org/10.1145/3712676.3718340>

```
@inproceedings{10.5220/0012087200003538,
  author = {Neugebauer, Matthias},
  title = {Nagare Media Engine: Towards an Open-Source Cloud- and Edge-Native NBMP Implementation},
  year = {2023},
  isbn = {978-989-758-665-1},
  issn={2184-2833},
  organization={INSTICC},
  publisher = {SciTePress - Science and Technology Publications},
  address = {Setúbal, Portugal},
  url = {https://doi.org/10.5220/0012087200003538},
  doi = {10.5220/0012087200003538},
  abstract = {Making efficient use of cloud and edge computing resources in multimedia workflows that span multiple providers poses a significant challenge. Recently, MPEG published ISO/IEC 23090-8 Network-Based Media Processing (NBMP), which defines APIs and data models for network-distributed multimedia workflows. This standardized way of describing workflows over various cloud providers, computing models and environments will benefit researchers and practitioners alike. A wide adoption of this standard would enable users to easily optimize the placement of tasks that are part of the multimedia workflow, potentially leading to an increase in the quality of experience (QoE). As a first step towards a modern open-source cloud- and edge-native NBMP implementation, we have developed the NBMP workflow manager Nagare Media Engine based on the Kubernetes platform. We describe its components in detail and discuss the advantages and challenges involved with our approach. We evaluate Nagare Media Engine in a test scenario and show its scalability.},
  booktitle = {Proceedings of the 18th International Conference on Software Technologies - ICSOFT},
  pages = {404-411},
  numpages = {8},
  keywords = {nbmp, network-distributed multimedia processing},
  location = {Rome, Italy}
}
```

```
@inproceedings{10.1145/3625468.3652167,
  author = {Neugebauer, Matthias},
  title = {Nagare Media Engine: Task Error Recovery in MPEG NBMP Workflows Through Event Sourcing},
  year = {2024},
  isbn = {979-8-4007-0412-3/24/04},
  publisher = {Association for Computing Machinery},
  address = {New York, NY, USA},
  url = {https://doi.org/10.1145/3625468.3652167},
  doi = {10.1145/3625468.3652167},
  abstract = {Multimedia workflows have become complex distributed systems that are deployed in a multi-cloud and multi-edge fashion. Such systems are prone to errors either in hardware or software. Modern multimedia workflows therefore need to design an appropriate error-handling strategy. Because tasks are often computationally expensive and run for a long time, simply restarting and redoing prior work is an inadequate solution. Instead, tasks should create regular checkpoints and continue from the last good state after a restart. We propose to use an event sourcing approach for recording and potentially replaying state changes in the form of published events. In this paper, we adopt this approach for Network-Based Media Processing (NBMP), an MPEG standard published as ISO/IEC 23090-8 that defines APIs and data models for network-distributed multimedia workflows. Additionally, we developed our approach as an extension to Nagare Media Engine, our existing open source NBMP implementation based on the Kubernetes platform. We evaluated our approach in scene detection and video encoding scenarios with simulated disruptions and observed significant speedups.},
  booktitle = {Proceedings of the 15th ACM Multimedia Systems Conference},
  pages = {257-263},
  numpages = {7},
  keywords = {encoding, error recovery, event sourcing, nbmp, network-distributed multimedia processing},
  location = {Bari, Italy},
  series = {MMSys '24}
}
```

```
@inproceedings{10.1145/3712676.3718340,
  author = {Neugebauer, Matthias},
  title = {Nagare Media Engine: Towards Self-Adapting MPEG NBMP Multimedia Workflows},
  year = {2025},
  isbn = {979-8-4007-1467-2/2025/03},
  publisher = {Association for Computing Machinery},
  address = {New York, NY, USA},
  url = {https://doi.org/10.1145/3712676.3718340},
  doi = {10.1145/3712676.3718340},
  abstract = {With ISO/IEC 23090-8 Network-Based Media Processing (NBMP), MPEG published a standard for today's multimedia workflows: complex distributed systems deployed in multi-cloud and multi-edge environments. It defines data models, APIs and a reference architecture for implementing and operating multimedia workflows as well as the corresponding workflow system. Next to the increased architectural complexity, multimedia workflows are subject to changing surroundings or objectives. As such, they are forced to adapt in order to still meet the workflow goals. In this paper, we explore self-adaptability in the context of NBMP to automate this process. We give an overview of adaptation types encountered in multimedia workflows. Moreover, we propose a general design for self-adapting NBMP workflows. Finally, we implement our approach as part of Nagare Media Engine, our existing open source NBMP implementation. The evaluation of our prototype demonstrates self-adaptability in an HTTP Adaptive Streaming (HAS) scenario resulting in decreased computing resource usage.},
  booktitle = {Proceedings of the 16th ACM Multimedia Systems Conference},
  pages = {263-269},
  numpages = {7},
  keywords = {nbmp, network-based media processing, multimedia, workflow, self-adapting, streaming},
  location = {Stellenbosch, South Africa},
  series = {MMSys '25}
}
```

## License

Apache 2.0 (c) nagare media authors
