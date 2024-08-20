# Example for `mmsys-test-scene-detection` function

This example shows task execution of the `mmsys-test-scene-detection` function demonstrating MPEG NBMP task error recovery using event sourcing. The NBMP tasks are created directly without going through the NBMP Workflow API and the full workflow manager. However, the nagare media engine `workflow-manager-helper` is deployed as a sidecar container next to the task container. The `workflow-manager-helper` will read the mounted secret and configure the task using the NBMP Task API. It then monitors the execution status.

The `mmsys-test-scene-detection` function can be configured with the following options:

| Config                                                                           | Description                                                         |
| -------------------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `mmsys-test-scene-detection.engine.nagare.media/test`                            | Execution test/mode to use (s. below). (required)                   |
| `mmsys-test-scene-detection.engine.nagare.media/max-number-of-simulated-crashes` | Number crashes, i.e. hard terminations, to simulate (defaults to 1) |
| `mmsys-test-scene-detection.engine.nagare.media/simulated-crash-wait-duration`   | Time to wait until simulated crash is triggered. (defaults to 20s) |

The following execution tests/modes exist:

| Test             | Description                                                                   |
| ---------------- | ----------------------------------------------------------------------------- |
| baseline         | Simple scene detection.                                                       |
| test-no-recovery | Scene detection with simulated crashes and no recovery.                       |
| test-recovery    | Scene detection with simulated crashes and recovery from last detected scene. |

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
$ kubectl create namespace mmsys-test-scene-detection
```

## Upload Test Media to In-Cluster S3

```sh
$ kubectl -n mmsys-test-scene-detection apply -f config/samples/k8s/mmsys-test-scene-detection/job-upload-media.yaml
# wait until job is complete
$ kubectl logs -f -n mmsys-test-scene-detection jobs/upload-media
```

## Create Tasks (`Secret` + `Job`)

In every test case, the output can be accessed at <http://s3.localtest.me/nagare-media-engine-tests/outputs/caminandes-1-llama-drama.txt> (`s3.localtest.me` resolves to 127.0.0.1 assuming the local DNS resolver allows that; the Kubernetes cluster has an Ingress that routes `s3.localtest.me` requests to the in-cluster S3).

```sh
# baseline
$ kubectl -n mmsys-test-scene-detection create secret generic workflow-manager-helper-data-mmsys-test-scene-detection-baseline \
    --from-file=data.yaml=config/samples/nagare-media/workflow-manager-helper-data_mmsys-test-scene-detection-baseline.yaml
$ kubectl -n mmsys-test-scene-detection apply -f config/samples/k8s/mmsys-test-scene-detection/job-task-mmsys-test-scene-detection-baseline.yaml
$ kubectl logs -f -n mmsys-test-scene-detection jobs/task-mmsys-test-scene-detection-baseline -c function

# test-no-recovery
$ kubectl -n mmsys-test-scene-detection create secret generic workflow-manager-helper-data-mmsys-test-scene-detection-test-no-recovery \
    --from-file=data.yaml=config/samples/nagare-media/workflow-manager-helper-data_mmsys-test-scene-detection-test-no-recovery.yaml
$ kubectl -n mmsys-test-scene-detection apply -f config/samples/k8s/mmsys-test-scene-detection/job-task-mmsys-test-scene-detection-test-no-recovery.yaml
$ kubectl logs -f -n mmsys-test-scene-detection jobs/task-mmsys-test-scene-detection-test-no-recovery -c function

# test-recovery
$ kubectl -n mmsys-test-scene-detection create secret generic workflow-manager-helper-data-mmsys-test-scene-detection-test-recovery \
    --from-file=data.yaml=config/samples/nagare-media/workflow-manager-helper-data_mmsys-test-scene-detection-test-recovery.yaml
$ kubectl -n mmsys-test-scene-detection apply -f config/samples/k8s/mmsys-test-scene-detection/job-task-mmsys-test-scene-detection-test-recovery.yaml
$ kubectl logs -f -n mmsys-test-scene-detection jobs/task-mmsys-test-scene-detection-test-recovery -c function
```

## Delete Kubernetes Cluster

```sh
$ make kind-down
```
