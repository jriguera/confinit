#!/usr/bin/env bash

{{ if .Data.system.timezone }}
echo -n "* Setting timezone to {{ .Data.system.timezone }} ... "
echo "{{ .Data.system.timezone }}" > /etc/timezone
ln -sf /usr/share/zoneinfo/{{ .Data.system.timezone }} /etc/localtime
echo "done"
{{ end }}

exit 0

