#!/usr/bin/env bash

{{ if .Data.system.users }}
{{ range $i, $user := .Data.system.users }}
### START {{ $user.name }}
getent passwd {{ $user.name }} > /dev/null || useradd --user-group --create-home {{ $user.name }}
{{ if $user.shell }}usermod --shell '{{ $user.shell }}' {{ $user.name }} {{ else }}usermod --shell '/bin/bash' {{ $user.name }} {{ end }}
{{ if $user.comment }}usermod --comment '{{ $user.comment }}' {{ $user.name }} {{ end }}
{{ if $user.uid }}usermod --uid {{ $user.uid }} {{ $user.name }} {{ end }}
{{ if $user.groups }}usermod -G '{{ $user.groups | sort | join "," }}' {{ $user.name }} {{ end }}
chmod 700 ~{{ $user.name }}
passwd -d {{ $user.name }}
{{ if $user.crypted_password }}
echo '{{ $user.name }}:{{ $user.crypted_password }}' | chpasswd -e
{{ else if $user.password }}
echo '{{ $user.name }}:{{ $user.password }}' | chpasswd
{{ end }}
mkdir -p ~{{ $user.name }}/.ssh
rm -f ~{{ $user.name }}/.ssh/authorized_keys
{{ if $user.disabled }}
# Disabling user
usermod --shell '/sbin/nologin' {{ $user.name }}
passwd -l {{ $user.name }}
{{ else if $user.public_keys }}
{{ range $i, $key := $user.public_keys }}
echo '{{ $key }}' >> ~{{ $user.name }}/.ssh/authorized_keys
{{ end }}
[ -f ~{{ $user.name }}/.ssh/authorized_keys ] && chmod 600 ~{{ $user.name }}/.ssh/authorized_keys
{{ end }}
chmod 700 ~{{ $user.name }}/.ssh
chown -R {{ $user.name }}:{{ $user.name }} ~{{ $user.name }}/.ssh
if ! [ -f ~{{ $user.name }}/.profile ]
then
	[ -f /etc/skel/.profile ] && cp /etc/skel/.profile ~{{ $user.name }}/.profile
	chown {{ $user.name }}:{{ $user.name }} ~{{ $user.name }}/.profile
fi
### END {{ $user.name }}
{{ end }}
{{ end }}