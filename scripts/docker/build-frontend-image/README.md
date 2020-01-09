## Service Center Frontend

This script helps you to make the docker image of service-center frontend present in "PROJECT_ROOT/frontend" folder. Where PROJECT_ROOT is project's top level folder. The script file "PROJECT_ROOT/scripts/frontend/start_linux.sh" will be the entrypoint for docker container.

### How To Run

This script assumes you have already downloaded all the dependencies using 'glide install'. Make sure service-center application is running and get service-center applications IP and PORT addresses.

Then build docker image of frontend from folder "PROJECT_ROOT/scripts/docker/build-frontend-image/".

    bash -x build.sh
    
Once the script finishes you will see image scfrontend-dev.tgz in the same directory. You can use this this image to load the docker and start using. Run following command, open Web Browser and connect to URL "http://<HOST_IP>:30103" to view frontend UI.

    docker run -d -p 30103:30103 -e SC_ADDRESS=${SC_ADDRESS} -t servicecomb/scfrontend

Where HOST_IP is the ip of host machine where docker container for frontend is executing.

Note: 

1. The same image can be used to deploy in Huawei Public Cloud using CFE/CCE or just deploy in a VM.
1. Configuring SC_ADDRESS will enable you to setup the EIP of the service-center.
