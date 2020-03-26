#!/usr/bin/env bash

{{ if .Data.system.hostname }}
echo "* Setting hostname ..."
hostname {{ .Data.system.hostname }}{{ if .Data.system.domain }}.{{ .Data.system.domain }}{{ end }}
echo "{{ .Data.system.hostname }}{{ if .Data.system.domain }}.{{ .Data.system.domain }}{{ end }}" > /etc/hostname
hostnamectl set-hostname {{ .Data.system.hostname }}{{ if .Data.system.domain }}.{{ .Data.system.domain }}{{ end }} || true
hostnamectl --transient set-hostname {{ .Data.system.hostname }} || true
{{ else }}
exit 0
{{ end }}
