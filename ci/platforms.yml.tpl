{{/*
This is the template for the pipeline template input.
*/ -}}
{{- $Values := . -}}
# Generated using "task set-pipeline"
platforms:
{{- range $Values.arch}}
{{- $arch := . }}
{{- range .os}}
{{- $os := . }}
{{- if $arch.variants}}
{{- range $arch.variants}}
{{- $variant := . }}
- os: {{$os}}
  arch: {{$arch.name}}
  var: "{{$variant}}"
{{- range $Values.os}}
{{- if eq .name $os}}
{{- if .docker}}
  docker: true
{{- end }}
{{- end }}
{{- end }}
{{- end }}
{{- else }}
- os: {{$os}}
  arch: {{$arch.name}}
{{- range $Values.os}}
{{- if eq .name $os}}
{{- if .docker}}
  docker: true
{{- end }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}
