#!/usr/bin/env bash

{{ if .Data.system.pi }}
{{ if .Data.system.pi.crypted_password }}
echo -n "* Changing password for pi ... "
echo 'pi:{{ .Data.system.pi.crypted_password }}' | chpasswd -e
echo "done"
{{ else if .Data.system.pi.password }}
echo -n "* Changing password for pi ... "
echo "pi:{{ .Data.system.pi.password }}" | chpasswd
echo "done"
{{ end }}
{{ if .Data.system.pi.groups }}
echo "* Adding pi to groups: {{ .Data.system.pi.groups | sort | join "," }}"
usermod -G '{{ .Data.system.pi.groups | sort | join "," }}' pi
{{ end }}
mkdir -p ~pi/.ssh
{{ if .Data.system.pi.public_keys }}
rm -f ~pi/.ssh/authorized_keys
echo "* Adding authorized_keys ..."
{{ range $i, $key := .Data.system.pi.public_keys }}
echo '{{ $key }}' >> ~pi/.ssh/authorized_keys
{{ end }}
{{ end }}
[ -f ~pi/.ssh/authorized_keys ] && chmod 600 ~pi/.ssh/authorized_keys
chmod 700 ~pi/.ssh
chown -R pi:pi ~pi/.ssh
{{ end }}
exit 0
