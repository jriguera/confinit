#!/usr/bin/env bash

# Docker
{{ if .Data.docker.Compose }}
echo -n "* Enable Docker Compose ... "
rm -f /etc/docker-compose/docker-compose-disabled
{{ else }}
echo -n "* Disable Docker Compose ... "
echo "Disabled by confinit, $(date)" > /etc/docker-compose/docker-compose-disabled
{{ end }}

echo "done"

