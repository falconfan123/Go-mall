{{- /*
  Helm Chart helpers for go-mall
*/ -}}

{{- /*
  生成 Service 名称
*/ -}}
{{- define "go-mall.serviceName" -}}
{{- .name -}}-rpc
{{- end -}}

{{- /*
  标签选择器
*/ -}}
{{- define "go-mall.labels" -}}
app: {{ .name }}
tier: backend
{{- end -}}