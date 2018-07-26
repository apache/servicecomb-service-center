## Apache-Incubator-ServiceComb-Service-Center-Frontend

Service-Center UI enables user to view the list of MicroServices registered in SC. Users can view the detailed information of their MicroServices, Instances and Schemas.
Service-Center UI also offers a unique feature of testing the Schemas of their MicroServices from UI, Users can also download the html client for their Schemas.

### QuickStart Guide

Easiest way to get started with Service-Center UI is to download the release from [here](https://dist.apache.org/repos/dist/dev/incubator/servicecomb/incubator-servicecomb-service-center/) and then untar/unzip it based on your OS and run start-frontend.sh/start-frontend.bat.
This will bring up the Service-Center UI on [http://127.0.0.1:30103](http://127.0.0.1:30103).

Windows(apache-incubator-servicecomb-service-center-XXX-windows-amd64.zip):
```
start-frontend.bat
```

Linux(apache-incubator-servicecomb-service-center-XXXX-linux-amd64.tar.gz):
```sh
./start-frontend.sh
```

Note: By default frontend runs on 127.0.0.1, if you want to change this then you can change it in `conf/app.conf`. 
```
frontend_host_ip=127.0.0.1
frontend_host_port=30103
```

##### Running UI from source code
However if you want to try our latest code then you can follow the below steps
```
#Make sure your GOPATH is set correctly as the UI runs on GO Backend Server
git clone https://github.com/apache/incubator-servicecomb-service-center.git $GOPATH/src/github.com/apache/incubator-servicecomb-service-center
cd $GOPATH/src/github.com/apache/incubator-servicecomb-service-center

cp -r etc/conf frontend/

cd frontend/app

bower install

cd ..

go run main.go
```
This will bring up the Service-Center UI on [http://127.0.0.1:30103](http://127.0.0.1:30103).
If you want to change the listening ip/port, you can modify it in the configuration file (service-center/frontend/conf/app.conf : frontend_host_ip, frontend_host_port).

### Preview of Service-Center UI
![Service-Center Preview](/docs/Service-Center-UI-Preview.gif)

### Feature List and RoadMap of Service-Center UI
Below is the comprehensive list of features which Service-Center UI offers, we are working constantly to improve the user experience and offer more useful features to leverage the features offered by Service-Center.
We Welcome our community members to come forward and help us to build this UI together.

|Sl|Feature|Status|
|--|-------|------|
|1|Dashobard to display the overall MicroService Statistics|Done|
|2|Service List with basic Information| Done|
|3|Instance List for MicroServices| Done|
|4|Provider List for MicroServices|Done|
|5|Consumer List for MicroServices|Done|
|6|Schema List for MicroServices|Done|
|7|Test Schema for MicroServices|Done|
|8|Pact Broker Support|Done|
|9|Topology View for Microservices Relationship|TBD|

Any Contribution(issues,PR,Documentation,Translation) will be highly appreciated.
