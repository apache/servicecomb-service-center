#!/usr/bin/env sh

set -e
umask 027

cd /opt/service-center

if [ ! -z "${ETCD_ADDRESS}" ]; then
    sed -i "s|^registry_plugin.*=.*$|registry_plugin = etcd|g" conf/app.conf
    sed -i "s|^# manager_cluster.*=.*$|manager_cluster = ${ETCD_ADDRESS}|g" conf/app.conf
fi

./service-center