# kube-plex

kube-plex is a scalable Plex Media Server solution for Kubernetes. It
distributes transcode jobs by creating pods in a Kubernetes cluster to perform
transcodes, instead of running transcodes on the Plex Media Server instance
itself.

## How it works

kube-plex works by replacing the Plex Transcoder program on the main PMS
instance with our own little shim. This shim intercepts calls to Plex
Transcoder, and creates Kubernetes pods to perform the work instead. These
pods use shared persistent volumes to store the results of the transcode (and
read your media!).

## Prerequisites

* A persistent volume type that supports ReadWriteMany volumes (e.g. NFS,
Amazon EFS)
* Your Plex Media Server *must* be configured to allow connections from
unauthorized users for your pod network, else the transcode job is unable to
report information back to Plex about the state of the transcode job. At some
point in the future this may change, but it is a required step in order to make
transcodes work right now.

## Setup

This guide will go through setting up a Plex Media Server instance on a
Kubernetes cluster, configured to launch transcode jobs on the same cluster
in pods created in the same 'plex' namespace.

1) Obtain a Plex Claim Token by visiting [plex.tv/claim](https://plex.tv/claim).
This will be used to bind your new PMS instance to your own user account
automatically.

2) Deploy the Helm chart included in this repository using the claim token
obtained in step 1. If you have pre-existing persistent volume claims for your
media, you can specify its name with `--set persistence.data.claimName`. If not
specified, a persistent volume will be automatically provisioned for you.

```bash
➜  helm install ./charts/kube-plex --name plex \
    --namespace plex \
    --set claimToken=[insert claim token here] \
    --set persistence.data.claimName=existing-pms-data-pvc \
    --set ingress.enabled=true
```

This will deploy a scalable Plex Media Server instance that uses Kubernetes as
a backend for executing transcode jobs.

3) Access the Plex dashboard, either using `kubectl port-forward`, or using
the services LoadBalancer IP (via `kubectl get service`), or alternatively use
the ingress provisioned in the previous step (with `--set ingress.enabled=true`).

4) Visit Settings->Server->Network and add your pod network subnet to the
`List of IP addresses and networks that are allowed without auth` (near the
bottom). For example, `10.100.0.0/16` is the subnet that pods in my cluster are
assigned IPs from, so I enter `10.100.0.0/16` in the box.

You should now be able to play media from your PMS instance - pods will be
created to handle transcodes, and data automatically mounted in appropriately:

```bash
➜  kubectl get po -n plex
NAME                              READY     STATUS    RESTARTS   AGE
plex-kube-plex-75b96cdcb4-skrxr   1/1       Running   0          14m
pms-elastic-transcoder-7wnqk      1/1       Running   0          8m
```



# Plex Media Server helm chart

## Configuration

The following tables lists the configurable parameters of the Plex chart and their default values.

| Parameter                  | Description                         | Default                                                 |
|----------------------------|-------------------------------------|---------------------------------------------------------|
| `image.repository`         | Image repository | `plexinc/pms-docker` |
| `image.tag`                | Image tag. Possible values listed [here](https://hub.docker.com/r/plexinc/pms-docker/tags/).| `1.10.1.4602-f54242b6b`|
| `image.pullPolicy`         | Image pull policy | `IfNotPresent` |
| `operator.enabled`         | Enable operator transcoder | `true` |
| `operator.image.repository`         | Image repository | `mcadm/plex-operator` |
| `operator.image.tag`                | Image tag. | `v0.0.1`|
| `operator.image.pullPolicy`         | Image pull policy | `IfNotPresent` |
| `claimToken`                 | Plex Claim Token to authenticate your acount | `` |
| `timezone`                 | Timezone plex instance should run as, e.g. 'America/New_York' | `Europe/London` |
| `service.type`          | Kubernetes service type for the plex GUI/API | `ClusterIP` |
| `service.port`          | Kubernetes port where the plex GUI/API is exposed| `32400` |
| `service.annotations`   | Service annotations for the Plex GUI | `{}` |
| `service.labels`        | Custom labels | `{}` |
| `service.loadBalancerIP` | Loadbalance IP for the Plex GUI | `{}` |
| `service.loadBalancerSourceRanges` | List of IP CIDRs allowed access to load balancer (if supported)      | None
| `ingress.enabled`              | Enables Ingress | `false` |
| `ingress.annotations`          | Ingress annotations | `{}` |
| `ingress.labels`               | Custom labels                       | `{}`
| `ingress.path`                 | Ingress path | `/` |
| `ingress.hosts`                | Ingress accepted hostnames | `chart-example.local` |
| `ingress.tls`                  | Ingress TLS configuration | `[]` |
| `rbac.create`                  | Create RBAC roles? | `true` |
| `nodeSelector`             | Node labels for pod assignment | `beta.kubernetes.io/arch: amd64` |
| `persistence.transcode.enabled`      | Use persistent volume for transcoding | `false` |
| `persistence.transcode.size`         | Size of persistent volume claim | `20Gi` |
| `persistence.transcode.claimName`| Use an existing PVC to persist data | `nil` |
| `persistence.transcode.subPath` | SubPath to use for existing Claim | `nil` |
| `persistence.transcode.storageClass` | Type of persistent volume claim | `-` |
| `persistence.data.size`         | Size of persistent volume claim | `40Gi` |
| `persistence.data.existingClaim`| Use an existing PVC to persist data | `nil` |
| `persistence.data.subPath` | SubPath to use for existing Claim | `nil` |
| `persistence.data.storageClass` | Type of persistent volume claim | `-` |
| `persistence.config.size`         | Size of persistent volume claim | `20Gi` |
| `persistence.config.existingClaim`| Use an existing PVC to persist data | `nil` |
| `persistence.config.subPath` | SubPath to use for existing Claim | `nil` |
| `persistence.config.storageClass` | Type of persistent volume claim | `-` |
| `resources`                | CPU/Memory resource requests/limits | `{}` |
| `podAnnotations`           | Key-value pairs to add as pod annotations  | `{}` |


Read through the [values.yaml](values.yaml) file. It has several commented out suggested values.
