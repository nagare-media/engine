# Example for mmsys-test-encode operation

This example shows task execution of the `mmsys-test-encode` function demonstrating MPEG NBMP task recovery using event sourcing.

```sh
# create cluster
make kind-up
kubectl create namespace mmsys-test-encode

# upload media to S3
kubectl -n mmsys-test-encode apply -f config/samples/k8s/mmsys-test-encode/job-upload-media.yaml
kubectl -n mmsys-test-encode wait --timeout=-1s --for=condition=complete job/upload-media
kubectl -n mmsys-test-encode delete job upload-media

# create secrets
kubectl -n mmsys-test-encode create secret generic workflow-manager-helper-data-mmsys-test-encode-baseline-simple "--from-literal=data.yaml=$(
  cat config/samples/nagare-media/workflow-manager-helper-data_mmsys-test-encode-baseline-simple.yaml \
  | sed 's/localhost:9000/minio.s3.svc.cluster.local:9000/g' \
  | sed 's/localhost:4222/nats.nats.svc.cluster.local:4222/g'
)"
kubectl -n mmsys-test-encode create secret generic workflow-manager-helper-data-mmsys-test-encode-baseline-split-merge "--from-literal=data.yaml=$(
  cat config/samples/nagare-media/workflow-manager-helper-data_mmsys-test-encode-baseline-split-merge.yaml \
  | sed 's/localhost:9000/minio.s3.svc.cluster.local:9000/g' \
  | sed 's/localhost:4222/nats.nats.svc.cluster.local:4222/g'
)"
kubectl -n mmsys-test-encode create secret generic workflow-manager-helper-data-mmsys-test-encode-test-no-recovery-simple "--from-literal=data.yaml=$(
  cat config/samples/nagare-media/workflow-manager-helper-data_mmsys-test-encode-test-no-recovery-simple.yaml \
  | sed 's/localhost:9000/minio.s3.svc.cluster.local:9000/g' \
  | sed 's/localhost:4222/nats.nats.svc.cluster.local:4222/g'
)"
kubectl -n mmsys-test-encode create secret generic workflow-manager-helper-data-mmsys-test-encode-test-no-recovery-split-merge "--from-literal=data.yaml=$(
  cat config/samples/nagare-media/workflow-manager-helper-data_mmsys-test-encode-test-no-recovery-split-merge.yaml \
  | sed 's/localhost:9000/minio.s3.svc.cluster.local:9000/g' \
  | sed 's/localhost:4222/nats.nats.svc.cluster.local:4222/g'
)"
kubectl -n mmsys-test-encode create secret generic workflow-manager-helper-data-mmsys-test-encode-test-recovery-split-merge "--from-literal=data.yaml=$(
  cat config/samples/nagare-media/workflow-manager-helper-data_mmsys-test-encode-test-recovery-split-merge.yaml \
  | sed 's/localhost:9000/minio.s3.svc.cluster.local:9000/g' \
  | sed 's/localhost:4222/nats.nats.svc.cluster.local:4222/g'
)"

# create jobs
kubectl -n mmsys-test-encode apply -f config/samples/k8s/mmsys-test-encode/job-task-mmsys-test-encode-baseline-simple.yaml
kubectl -n mmsys-test-encode wait --timeout=-1s --for=condition=complete job/task-mmsys-test-encode-baseline-simple

kubectl -n mmsys-test-encode apply -f config/samples/k8s/mmsys-test-encode/job-task-mmsys-test-encode-baseline-split-merge.yaml
kubectl -n mmsys-test-encode wait --timeout=-1s --for=condition=complete job/task-mmsys-test-encode-baseline-split-merge

kubectl -n mmsys-test-encode apply -f config/samples/k8s/mmsys-test-encode/job-task-mmsys-test-encode-test-no-recovery-simple.yaml
kubectl -n mmsys-test-encode wait --timeout=-1s --for=condition=complete job/task-mmsys-test-encode-test-no-recovery-simple

kubectl -n mmsys-test-encode apply -f config/samples/k8s/mmsys-test-encode/job-task-mmsys-test-encode-test-no-recovery-split-merge.yaml
kubectl -n mmsys-test-encode wait --timeout=-1s --for=condition=complete job/task-mmsys-test-encode-test-no-recovery-split-merge

kubectl -n mmsys-test-encode apply -f config/samples/k8s/mmsys-test-encode/job-task-mmsys-test-encode-test-recovery-split-merge.yaml
kubectl -n mmsys-test-encode wait --timeout=-1s --for=condition=complete job/task-mmsys-test-encode-test-recovery-split-merge
```
