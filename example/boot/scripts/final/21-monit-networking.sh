#!/usr/bin/env bash

{{ if .Data.monit }}
pushd /etc/monit/conf-enabled >/dev/null
{{ if .Data.networking }}
{{ range $key, $value := .Data.networking }}
  echo "* Checking {{ $key }} ..."
  if ip link show {{ $key }} 2>/dev/null
  then
    if [ -r "../conf-available/{{ $key }}" ]
    then
      echo "* Enabling {{ $key }} in monit"
      ln -sf ../conf-available/{{ $key }} {{ $key }}
    fi
  else
    echo "* Disabling {{ $key }} in monit"
    rm -f {{ $key }}
  fi
{{ end }}
{{ end }}
popd
{{ end }}

exit 0
