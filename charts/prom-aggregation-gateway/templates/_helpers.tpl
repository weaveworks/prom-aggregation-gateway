{{- define "prom-aggregation-gateway.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "prom-aggregation-gateway.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end -}}

{{- define "prom-aggregation-gateway.labels" -}}
helm.sh/chart: {{ include "prom-aggregation-gateway.chart" . }}
{{ include "prom-aggregation-gateway.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
{{- if .Values.commonLabels}}
{{ toYaml .Values.commonLabels }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "prom-aggregation-gateway.selectorLabels" -}}
app.kubernetes.io/name: {{ include "prom-aggregation-gateway.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
