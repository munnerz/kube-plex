# Plex Elastic Transcoder

This is a simple tool that should be used in place of the "Plex New Transcoder" binary on your Plex Media Server host in order to distribute the load of Plex transcoding.

It works by scheduling a job on a Kubernetes cluster, with the appropriate media and transcode directories exported via NFS to the transcoding containers.

### Todo list

- Add NFS mount directories to created pods for media and transcoding directories
- Make config load from a file