# kube-plex

kube-plex is a scalable Plex Media Server solution for Kubernetes. It
distributes transcode jobs by creating jobs in a Kubernetes cluster to perform
transcodes, instead of running transcodes on the Plex Media Server instance
itself.

[ressu/kube-plex](https://github.com/ressu/kube-plex) is a fork from
[munnerz/kube-plex](https://github.com/munnerz/kube-plex).

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

This guide describes a way to configure Kube-plex with a Plex Media Server instance on a
Kubernetes cluster. Provided helm chart is intended as both a reference and an
easy way of creating a kube-plex deployment. Advanced users are encouraged to
explore deployment options and to contribute additional deployment examples.

### Plex claim token

Create a Plex Claim Token by visiting [plex.tv/claim](https://plex.tv/claim). 
Plex claim token is used to associate a Plex instance with a Plex account. Claim
token expires automatically within minutes of creating the token. If Plex
instance fails to register with Plex Account when creating a new deployment, a
good troubleshooting step is to recreate the claim token and deployment.

Claim token isn't the only way to register an instance to a Plex Account. See
"Access Plex Dashboard" below.

### Helm chart deployment

Register the helm chart for this repository by running

```bash
helm repo add kube-plex https://ressu.github.io/kube-plex
```

Claim token and other chart configuration is defined with the `--set` flags.
Available configuration options are described in the
[chart](charts/kube-plex/README.md) and in more detail in
[values.yaml](charts/kube-plex/values.yaml).

If you have pre-existing persistent volume claims for your
media, you can specify its name with `--set persistence.data.claimName`. If not
specified, a persistent volume will be automatically provisioned for you.

In order for the transcoding to work, a shared transcode persistent volume claim
needs to be defined with `--set persistence.transcode.claimName` or by defining
the relevant parameters separately.

As an example, the following command would install or upgrade an existing plex
deployment with the given values:

```bash
helm upgrade plex kube-plex \
    --namespace plex \
    --install \
    --set claimToken=[insert claim token here] \
    --set persistence.data.claimName=[existing-pms-data-pvc] \
    --set persistence.transcode.enabled=true \
    --set persistence.transcode.claimName=[shared-pms-transcode-pvc] \
    --set ingress.enabled=true
```

This will deploy a scalable Plex Media Server instance that uses Kubernetes as
a backend for executing transcode jobs.

### Access Plex Dashboard

If you used claim token, the plex instance should be visible in [Plex Web
App](https://app.plex.tv). If the token registration failed, access the instance
using using `kubectl -n plex port-forward <pod name> 32400`.

Once the registration has been completed, the instance can be accessed via the
load balancer IP (via `kubectl get service`) or the ingress (if provisioned with
`--set ingress.enabled=true`).

### Set pod network allowlist

Visit Settings->Server->Network and add your pod network subnet to the
`List of IP addresses and networks that are allowed without auth` (near the
bottom). For example, `10.100.0.0/16` is the subnet that pods in my cluster are
assigned IPs from, so I enter `10.100.0.0/16` in the box.

You should now be able to play media from your PMS instance

## Internal operations

Kube-plex will automatically create transcoding jobs within the Kubernetes
instance. The jobs have shared transcode and data mounts with the main kube-plex
pod. Kube-plex replaces the `Plex Transcoder` binary with a launcher on Plex
startup. Kube-plex launcher processes the arguments from Plex and creates a
transcoding job to handle the final transcoding.

```bash
$ kubectl get pod,job
NAME                                        READY   STATUS    RESTARTS   AGE
pod/kube-plex-694d659b64-7wg2b              1/1     Running   0          6d23h
pod/pms-elastic-transcoder-tqw5s-8w2bc      1/1     Running   0          4s

NAME                                     COMPLETIONS   DURATION   AGE
job.batch/pms-elastic-transcoder-tqw5s   0/1           4s         5s
```

Transcoder pod will run a shim which will

* Download codecs from main kube-plex pod
* Relay transcoder callbacks from `Plex Transcoder` to main kube-plex

Logging from kube-plex processes is written to Plex process and can be viewed in `Settings->Manage->Console`.