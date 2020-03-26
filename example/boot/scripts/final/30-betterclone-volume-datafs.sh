#!/usr/bin/env bash

VOLUME="/media/volume-data"
DATA_SUBVOL="${VOLUME}/data"

{{ if .Data.betterclone }}
systemctl enable --now "`systemd-escape -p --template=betterclone-restore@.service ${DATA_SUBVOL}`"
systemctl enable --now "`systemd-escape -p --template=betterclone-backup@.timer ${DATA_SUBVOL}`"
{{ else }}
systemctl disable --now "`systemd-escape -p --template=betterclone-restore@.service ${DATA_SUBVOL}`"
systemctl disable --now "`systemd-escape -p --template=betterclone-backup@.timer ${DATA_SUBVOL}`"
{{ end }}

exit 0
