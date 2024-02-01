# Example for mmsys-test-encode operation

This example shows task execution of the `mmsys-test-encode` function demonstrating MPEG NBMP task error recovery using event sourcing. The NBMP tasks are created directly without going through the NBMP Workflow API and the full workflow manager. However, the nagare media engine `workflow-manager-helper` is deployed as a sidecar container next to the task container. The `workflow-manager-helper` will read the mounted secret and configure the task using the NBMP Task API. It then monitors the execution status.

The `mmsys-test-encode` function can be configured with the following options:

| Config                                                                  | Description                                                         |
| ----------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `mmsys-test-encode.engine.nagare.media/test`                            | Execution test/mode to use (s. below). (required)                   |
| `mmsys-test-encode.engine.nagare.media/chunk-seconds`                   | Duration of a chunk for split-merge execution modes. (required)     |
| `mmsys-test-encode.engine.nagare.media/max-number-of-simulated-crashes` | Number crashes, i.e. hard terminations, to simulate (defaults to 2) |
| `mmsys-test-encode.engine.nagare.media/simulated-crash-wait-duration`   | Time to wait until simulated crash is triggered. (defaults to 120s) |

The following execution tests/modes exist:

| Test                         | Description                                                           |
| ---------------------------- | --------------------------------------------------------------------- |
| baseline-simple              | Simple encode of the whole file.                                      |
| baseline-split-merge         | Simple split-and-merge encode. The chunks are processes sequentially. |
| test-no-recovery-simple      | Encode of the whole file with simulated crashes without recovery.     |
| test-no-recovery-split-merge | Split-and-merge encode with simulated crashes without recovery.       |
| test-recovery-split-merge    | Split-and-merge encode with simulated crashed with recovery.          |

## Host environment

The following sections assume that certain software is installed on the host environment:

* Docker
* `sh` (UNIX shell)
* `bash`
* `make`
* `kind`
* `kubectl`
* `helm`

Alternatively, the commands can be executed in a Docker container that serves as the test environment. Use the following commands to initialize a new container and install the necessary tools. Then follow the guide normally but execute the commands within the container environment.

```sh
# create kind network; ignore if it already exist
$ docker network create \
    --driver bridge \
    --ipv6 --subnet fc00:f853:ccd:e793::/64 \
    --opt com.docker.network.bridge.enable_ip_masquerade=true \
    --opt com.docker.network.driver.mtu=65535 \
    kind

$ docker run --rm -it --privileged \
    --name nagare-media-engine-test-env \
    --network kind \
    -v "/var/run/docker.sock:/var/run/docker.sock" \
    -v "$PWD:/src" \
    -w /src \
    docker.io/library/alpine:edge

# within the container:

# install necessary tools
docker$ echo "https://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories
docker$ apk add \
          bash \
          docker-cli \
          helm \
          kind \
          kubectl \
          make

# continue with the guide
docker$ ...

# exit (with automatic cleanup)
docker$ exit
```

## Create Kubernetes Cluster

```sh
$ make kind-up
$ kubectl create namespace mmsys-test-encode
```

## Upload Test Media to In-Cluster S3

```sh
$ kubectl -n mmsys-test-encode apply -f config/samples/k8s/mmsys-test-encode/job-upload-media.yaml
# wait until job is complete
$ kubectl logs -f -n mmsys-test-encode jobs/upload-media
```

## Create Tasks (`Secret` + `Job`)

In every test case, the output encoding can be accessed at <http://s3.local.gd/nagare-media-engine-tests/outputs/caminandes-1-llama-drama.mp4> (`s3.local.gd` resolves to 127.0.0.1 assuming the local DNS resolver allows that; the Kubernetes cluster has an Ingress that routes `s3.local.gd` requests to the in-cluster S3).

```sh
# baseline-simple
$ kubectl -n mmsys-test-encode create secret generic workflow-manager-helper-data-mmsys-test-encode-baseline-simple \
    --from-file=data.yaml=config/samples/nagare-media/workflow-manager-helper-data_mmsys-test-encode-baseline-simple.yaml

$ kubectl -n mmsys-test-encode apply -f config/samples/k8s/mmsys-test-encode/job-task-mmsys-test-encode-baseline-simple.yaml
$ kubectl logs -f -n mmsys-test-encode jobs/task-mmsys-test-encode-baseline-simple -c function

# baseline-split-merge
$ kubectl -n mmsys-test-encode create secret generic workflow-manager-helper-data-mmsys-test-encode-baseline-split-merge \
    --from-file=data.yaml=config/samples/nagare-media/workflow-manager-helper-data_mmsys-test-encode-baseline-split-merge.yaml

$ kubectl -n mmsys-test-encode apply -f config/samples/k8s/mmsys-test-encode/job-task-mmsys-test-encode-baseline-split-merge.yaml
$ kubectl logs -f -n mmsys-test-encode jobs/task-mmsys-test-encode-baseline-split-merge -c function

# test-no-recovery-simple
$ kubectl -n mmsys-test-encode create secret generic workflow-manager-helper-data-mmsys-test-encode-test-no-recovery-simple \
    --from-file=data.yaml=config/samples/nagare-media/workflow-manager-helper-data_mmsys-test-encode-test-no-recovery-simple.yaml

$ kubectl -n mmsys-test-encode apply -f config/samples/k8s/mmsys-test-encode/job-task-mmsys-test-encode-test-no-recovery-simple.yaml
$ kubectl logs -f -n mmsys-test-encode jobs/task-mmsys-test-encode-test-no-recovery-simple -c function

# test-no-recovery-split-merge
$ kubectl -n mmsys-test-encode create secret generic workflow-manager-helper-data-mmsys-test-encode-test-no-recovery-split-merge \
    --from-file=data.yaml=config/samples/nagare-media/workflow-manager-helper-data_mmsys-test-encode-test-no-recovery-split-merge.yaml

$ kubectl -n mmsys-test-encode apply -f config/samples/k8s/mmsys-test-encode/job-task-mmsys-test-encode-test-no-recovery-split-merge.yaml
$ kubectl logs -f -n mmsys-test-encode jobs/task-mmsys-test-encode-test-no-recovery-split-merge -c function

# test-recovery-split-merge
$ kubectl -n mmsys-test-encode create secret generic workflow-manager-helper-data-mmsys-test-encode-test-recovery-split-merge \
    --from-file=data.yaml=config/samples/nagare-media/workflow-manager-helper-data_mmsys-test-encode-test-recovery-split-merge.yaml

$ kubectl -n mmsys-test-encode apply -f config/samples/k8s/mmsys-test-encode/job-task-mmsys-test-encode-test-recovery-split-merge.yaml
$ kubectl logs -f -n mmsys-test-encode jobs/task-mmsys-test-encode-test-recovery-split-merge -c function
```

## Delete Kubernetes Cluster

```sh
$ make kind-down
```
