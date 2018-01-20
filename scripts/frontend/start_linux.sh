#!/bin/bash

cd /opt/frontend/
cd conf/
. app.conf
echo $SC_HOST_IP
echo $SC_HOST_PORT
echo $SC_HOST_MODE
cd ../
sed -i '/ip/c\ip:"'$SC_HOST_MODE'://'$SC_HOST_IP'",' ./app/apiList/apiList.js
sed -i '/port/c\port:"'$SC_HOST_PORT'"' ./app/apiList/apiList.js
./scfrontend > start-sc-frontend.log 2>&1
