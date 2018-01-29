### Service-Center Release

#### Making Linux Release

To make a linux release of Service-center please follow the below steps

```
## Clone the code in a proper GOPATH

git clone https://github.com/apache/incubator-servicecomb-service-center.git $GOPATH/src/github.com/apache/incubator-servicecomb-service-center
cd $GOPATH/src/github.com/apache/incubator-servicecomb-service-center

## Donwload all the dependency
go get github.com/FiloSottile/gvt
gvt restore

## Run the release script
bash -x scripts/release/make_linux_release.sh

## Your release is ready
root@SZX1000272432:/home/asif/go/src/github.com/apache/incubator-servicecomb-service-center# ll
total 11848
drwxr-xr-x 15 root root     4096 Jan 29 13:45 ./
drwxr-xr-x  5 root root     4096 Jan 29 13:03 ../
-rw-r--r--  1 root root      564 Jan 29 13:03 DISCLAIMER
drwxr-xr-x  2 root root     4096 Jan 29 13:03 docs/
drwxr-xr-x  3 root root     4096 Jan 29 13:03 etc/
drwxr-xr-x  3 root root     4096 Jan 29 13:03 examples/
drwxr-xr-x  5 root root     4096 Jan 29 13:03 frontend/
drwxr-xr-x  8 root root     4096 Jan 29 13:16 .git/
drwxr-xr-x  2 root root     4096 Jan 29 13:03 .github/
-rw-r--r--  1 root root      414 Jan 29 13:03 .gitignore
-rw-r--r--  1 root root       96 Jan 29 13:03 .gitmodules
drwxr-xr-x  2 root root     4096 Jan 29 13:03 integration/
-rw-r--r--  1 root root    11357 Jan 29 13:03 LICENSE
-rw-r--r--  1 root root     1012 Jan 29 13:03 main.go
-rw-r--r--  1 root root      353 Jan 29 13:03 .mention-bot
-rw-r--r--  1 root root     2590 Jan 29 13:03 NOTICE
drwxr-xr-x 19 root root     4096 Jan 29 13:03 pkg/
-rw-r--r--  1 root root     4358 Jan 29 13:03 README.md
drwxr-xr-x  5 root root     4096 Jan 29 13:03 scripts/
drwxr-xr-x 15 root root     4096 Jan 29 13:03 server/
-rw-r--r--  1 root root 12021226 Jan 29 13:45 servicecomb-service-center-1.0.0-m1-linux-amd64.tar.gz
-rw-r--r--  1 root root     1322 Jan 29 13:03 .travis.yml
drwxr-xr-x  6 root root     4096 Jan 29 13:06 vendor/
drwxr-xr-x  2 root root     4096 Jan 29 13:03 version/

```
