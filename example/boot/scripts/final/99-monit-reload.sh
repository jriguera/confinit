#!/usr/bin/env bash

{{ if .Data.monit }}
monit reload || true
{{ end }}

exit 0

