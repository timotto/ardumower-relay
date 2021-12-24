{{/*
Expand the build variant.

The pipeline builds executable binaries for all the given OS and processor architecture
*/}}
{{- define "variant" -}}
    {{ .os }}-{{ .arch }}
    {{- if .var -}}
        {{ .var }}
    {{- end -}}
{{- end -}}

{{/*
Expand the build resource name
*/}}
{{- define "build" -}}
  build-{{template "variant" .}}
{{- end -}}

{{/*
Expand the docker image resource name
*/}}
{{- define "image" -}}
  image-{{template "variant" .}}
{{- end -}}

{{/*
Expand the oci image IMAGE_PLATFORM parameter value
*/}}
{{- define "imagePlatform" -}}
    {{ .os }}/{{ .arch }}{{template "imagePlatformVariant" .}}
{{- end -}}

{{/*
Expand the optional oci-build-task parameter value for "IMAGE_PLATFORM"
Adds "/v..." suffix for ARM architectures
*/}}
{{- define "imagePlatformVariant" -}}
    {{- if eq .arch "arm" -}}
      /v{{.var}}
    {{- else if eq .arch "arm64" -}}
      /v8
    {{- end -}}
{{- end -}}

{{/*
Expand the optional docker manifest resource parameter "variant" for ARM architectures
*/}}
{{- define "manifestPlatformVariant" -}}
    {{- if eq .arch "arm" -}}
      variant: v{{.var}}
    {{- else if eq .arch "arm64" -}}
      variant: v8
    {{- end -}}
{{- end -}}

{{/*
Expand the optional go build parameter "GOARM"
*/}}
{{- define "goArmParam" -}}
    {{- if eq .arch "arm" -}}
      GOARM: "{{.var}}"
    {{- end -}}
{{- end -}}
