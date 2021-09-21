# Example configuration for Kustomize based kube-plex setup

This directory contains an example Kustomize based configuraiton for kube-plex.
The configuration doesn't create PersistentVolume (PV) or PersistentVolumeClaim
(PVC) objects. Creating PV and PVC definitions is left as a task to the cluster
maintainer.

## Using the templates with a two directory layout

Templates can be used directly from this repository or the templates can be
copied to a centralized storage. [Base](base/) directory contains a basic
kube-plex installation. [Myplex](myplex/) directory is an overlay directory and
contains an example on how to customise the base deployment.  All modifications
to the deployment are made to the overlay directory.

Two directory based setup is updated by replacing the base directory with the
new version from upstream.

## Configuration

Overlay directory contains two evxample patches to the base deployment:
- [deployment.yaml](myplex/deployment.yaml)
- [service.yaml](myplex/service.yaml)

Patch files update the full deployment in base directory. More details on how
to use patches can be found in [kubectl
documentation](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/patches/).

In addition to patches, kustomize templates contain shared definitions:

```yaml
commonLabels:
  release: my-plex
namePrefix: my-
namespace: plex
```

Namespace defines which namespace is used for deployment and overrides the
definitions in yaml files. Common labels is an easy way of defining a label
that is applied to all defined resources. Nameprefix and namesuffix can be used
to change the names of deployed resources.

## Applying and inspecting final configuration

`Kubectl` can be used to apply the final configuration.

```bash
kubectl apply -k myplex/
```

To inspect what would be changed in a deployment, the `diff` command can be used:

```bash
kubectl diff -k myplex/
```

To inspect the complete template, a separate tool called `kustomize` can be used instead:

```bash
kustomize build myplex/
```
