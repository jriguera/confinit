#!/usr/bin/env bash

VOLUME="/media/volume-data"

{{ if .Data.monit }}
pushd /etc/monit/conf-enabled >/dev/null
  if btrfs filesystem show ${VOLUME} 2>/dev/null
  then
    echo "* Enabling volume-datafs in monit"
    ln -sf ../conf-available/volume-datafs volume-datafs
  else
    echo "* Disabling volume-datafs in monit"
    rm -f volume-datafs
  fi
popd
{{ end }}

exit 0

