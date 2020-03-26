#!/usr/bin/env bash

sshkeygen() {
    # Generate entropy
    dd if=/dev/hwrng of=/dev/urandom count=1 bs=4096
    echo "* Deleting old ssh host keys ..."
    rm -f /etc/ssh/ssh_host_*
    echo "* Generate new host keys ..."
    ssh-keygen -A -v
}

{{ if .Data.sshd }}
# SSH is enabled
{{ if .Data.sshd.KeygenVersion }}
if [ ! -r /etc/ssh/keygen.env ]
then
    sshkeygen && {
        echo "# Automatically generated host keys at $(date) by confinit"
        echo "STATE={{ .Data.sshd.KeygenVersion }}"
    } > /etc/ssh/keygen.env
else
    if [ -r /etc/ssh/keygen.env ]
    then
        (
            . /etc/ssh/keygen.env
            if [ -z "$STATE" ]
            then
                sshkeygen && {
                    echo "# Automatically generated host keys at $(date) by confinit"
                    echo "STATE={{ .Data.sshd.KeygenVersion }}"
                } > /etc/ssh/keygen.env
            else
                [ "$STATE" != "{{.Data.sshd.KeygenVersion }}" ] && sshkeygen & {
                    echo "# Automatically generated host keys at $(date) by confinit"
                    echo "STATE={{ .Data.sshd.KeygenVersion }}"
                } > /etc/ssh/keygen.env
            fi
        )
    fi
fi
{{ end }}
{{ end }}

exit 0

