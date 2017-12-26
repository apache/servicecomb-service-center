### Creating Custom Docker Image for [Service-Center](https://github.com/apache/incubator-servicecomb-service-center)

There is already a docker image in our dockerhub [repo](https://hub.docker.com/r/servicecomb/service-center/), you can use this to run the service-center in docker container.
```
docker run -d -p 30100:30100 servicecomb/service-center
```

However if you want to customize the service-center and make your own docker image then follow the steps below. By default we use `quay.io/coreos/etcd` as our base image as we need etcd to be running before starting service-center, if you want to change the base image then you can change [here](scripts/docker/build-image/build.sh#L30).

##### 1. Make Sure your service-center is in correct GOPATH and then download all the dependencies
```
git clone https://github.com/apache/incubator-servicecomb-service-center.git $GOPATH/src/github.com/apache/incubator-servicecomb-service-center
cd $GOPATH/src/github.com/apache/incubator-servicecomb-service-center

go get github.com/FiloSottile/gvt
gvt restore
```

##### 2. Run the build script to make the docker image
```
# cd scripts/docker/build-image/
# bash +x build.sh 
Sending build context to Docker daemon  40.95MB
Step 1/3 : FROM quay.io/coreos/etcd:v3.1.0
 ---> d611dede458a
Step 2/3 : COPY ./service-center ./start.sh /root/
 ---> 8df8858f0720
Removing intermediate container 3ffea40f15b2
Step 3/3 : ENTRYPOINT /root/start.sh
 ---> Running in 6aadc39c2d17
 ---> f9f8fbb0e612
Removing intermediate container 6aadc39c2d17
Successfully built f9f8fbb0e612
Successfully tagged developement/servicecomb/service-center:latest
```
```
# docker images
REPOSITORY                                TAG                 IMAGE ID            CREATED             SIZE
developement/servicecomb/service-center   latest              f9f8fbb0e612        35 seconds ago      74.6MB
```

##### 3. Run the docker image
```
# docker run -d -p 30100:30100 developement/servicecomb/service-center
c140cc4bdc449636cc911584e769bb75d1c973085014da3d530650cc4927ec45

# docker ps
CONTAINER ID        IMAGE                                     COMMAND             CREATED             STATUS              PORTS                                     NAMES
c140cc4bdc44        developement/servicecomb/service-center   "/root/start.sh"    4 seconds ago       Up 3 seconds        2379-2380/tcp, 0.0.0.0:30100->30100/tcp   amazing_leakey
```

##### 4. Verify the status of running container
```
# curl -v http://0.0.0.0:30100/health
* Hostname was NOT found in DNS cache
*   Trying 0.0.0.0...
* Connected to 0.0.0.0 (127.0.0.1) port 30100 (#0)
> GET /health HTTP/1.1
> User-Agent: curl/7.35.0
> Host: 0.0.0.0:30100
> Accept: */*
> 
< HTTP/1.1 200 OK
< Content-Type: application/json;charset=utf-8
* Server SERVICECENTER/3.0.0 is not blacklisted
< Server: SERVICECENTER/3.0.0
< X-Response-Status: 200
< Date: Fri, 04 Aug 2017 06:29:14 GMT
< Content-Length: 296
< 
* Connection #0 to host 0.0.0.0 left intact
{"instances":[{"instanceId":"043e7fa678de11e7b7300242ac110002","serviceId":"043dbea978de11e7b7300242ac110002","endpoints":["rest://0.0.0.0:30100"],"hostName":"service_center_172_17_0_2","status":"UP","healthCheck":{"mode":"push","interval":30,"times":3},"timestamp":"1501828060","environment":"production"}]}
```

