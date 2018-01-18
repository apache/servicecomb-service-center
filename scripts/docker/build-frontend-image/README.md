## Service-Center support for build docker image

This script helps you to make the docker image of service-center frontend  present in your current working directory.

The start_linux.sh will be the entrypoint for the docker container to start.

### How To Run

This script assumes you have already downloaded all the dependencies using 'gvt restore'.

    bash -x build.sh
    
Once the script finishes you will see image scfrontend-dev.tgz in the same directory. 
Load this image to the docker and start using.

    docker run -d -p 30101:30103 -t servicecomb/scfrontend
