#!/usr/bin/env bash

cd /etc/monit/conf-enabled

echo "* Checking eth0 ..."
if ip link show eth0 2>/dev/null
then
  echo "* Enabling eth0 in monit"
  ln -sf ../conf-available/eth0 eth0
else
  echo "* Disabling eth0 in monit"
  rm -f eth0
fi

echo "* Checking wlan0 ..."
if ip link show wlan0 2>/dev/null
then
  echo "* Enabling wlan0 in monit"
  ln -sf ../conf-available/wlan0 wlan0
else
  echo "* Disabling wlan0 in monit"
  rm -f wlan0
fi

echo "Checking /dev/sda ..."
disable=0
if ! lsblk -d -n -o name,size /dev/sda 2>/dev/null
then
  echo "Checking /dev/sdb ..."
  if ! lsblk -d -n -o name,size /dev/sdb 2>/dev/null
  then
    echo "* Disabling volume-datafs in monit"
    rm -f volume-datafs
    disable=1
  fi
fi
if [ "${disable}" == "0" ]
then
  echo "* Enabling volume-datafs in monit"
  ln -sf ../conf-available/volume-datafs volume-datafs
fi

exit 0
