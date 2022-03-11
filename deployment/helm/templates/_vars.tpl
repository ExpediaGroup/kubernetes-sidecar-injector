{{- define "common.labels" }}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/component: webhook
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "certs.secret.name" }}
{{- .Chart.Name }}
{{- end }}

{{- define "service.name" }}
{{- .Chart.Name }}
{{- end }}