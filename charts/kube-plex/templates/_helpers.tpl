{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the secret name for the claimToken
*/}}
{{- define "claimTokenSecretName" -}}
{{- if .Values.claimToken.secret -}}
    {{ printf "%s" .Values.claimToken.secret }}
{{- else -}}
    {{ printf "%s-%s" (include "fullname" .) "claim-token" | trunc 63 -}}
{{- end -}}
{{- end -}}
