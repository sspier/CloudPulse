{{/* Helper template for cloudpulse-api chart */}}
{{- define "cloudpulse-api.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/* Helper template to create a fullname for resources */}}
{{- define "cloudpulse-api.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name (include "cloudpulse-api.name" .) | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/* Helper template for common labels */}}
{{- define "cloudpulse-api.labels" -}}
app.kubernetes.io/name: {{ include "cloudpulse-api.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/part-of: cloudpulse
{{- end }}

{{/* Helper template for selector labels */}}
{{- define "cloudpulse-api.selectorLabels" -}}
app.kubernetes.io/name: {{ include "cloudpulse-api.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
