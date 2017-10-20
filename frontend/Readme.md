## Service-Center UI

Service-Center UI enables user to view the list of MicroServices registered in SC. Users can view the detailed information of their MicroServices, Instances and Schemas.
Service-Center UI also offers a unique feature of testing the Schemas of their MicroServices from UI, Users can also download the html client for their Schemas.

### QuickStart Guide

Easiest way to get started with Service-Center UI is to download the release from [here](https://github.com/ServiceComb/service-center/releases) and then untar/unzip it based on your OS and run start.sh/start.bat.
This will bring up the Service-Center UI on [http://127.0.0.1:30101](http://127.0.0.1:30101).

##### Running UI from source code
However if you want to try our latest code then you can follow the below steps
```
#Make sure your GOPATH is set correctly as the UI runs on GO Backend Server
git clone https://github.com/ServiceComb/service-center.git $GOPATH/src/github.com/ServiceComb/service-center
cd $GOPATH/src/github.com/ServiceComb/service-center

cd frontend
go run main.go
```
This will bring up the Service-Center UI on [http://127.0.0.1:30101](http://127.0.0.1:30101).

### Preview of Service-Center UI


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
|7|Test Schema for MicroServices|InProgress|
|8|Generate Client from Schema |TBD|
|9|Generate Server from Schema|TBD|

Any Contribution(issues,PR,Documentation,Translation) will be highly appreciated.
