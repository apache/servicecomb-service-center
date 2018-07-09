## Service-Center support for build docker image

This script helps you to make the docker image of service-center present in your current working directory.

Here quay.io/coreos/etcd:v3.1.0 used as etcd base image. start.sh will be the entrypoint for the docker container to start.

### How To Run

This script assumes you have already downloaded all the dependencies using 'glide install'.

    bash -x build.sh
    
Once the script finishes you will see image service-center-dev.tgz in the same directory. 
Load this image to the docker and start using.

    docker run -d -p 30100:30100 developement/servicecomb/service-center
