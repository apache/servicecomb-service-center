Deploying Service-Center
========================

Deploying Service-Center in Cluster Mode
----------------------------------------

As Service-center is a stateless application so it can be seamlessly
deployed in cluster mode to achieve HA. SC is dependent on the etcd to
store the microservices information so you can opt for running etcd
standalone or in `cluster`_ mode. Once you are done with installing the
etcd either in cluster or standalone mode then you can follow the below
steps to run the Service-Center.

Let’s assume you want to install 2 instances of Service-Center on VM
with following details

==== =========
Name Address
==== =========
VM1  10.12.0.1
VM2  10.12.0.2
==== =========

Here we assume your etcd is running on http://10.12.0.4:2379 (you can
follow `this`_ guide to install etcd in cluster mode.)

Step 1
~~~~~~

Download the SC release from `here`_ on all the VM’s.

::

   # Untar the release
   # tar -xvf service-center-X.X.X-linux-amd64.tar.gz

Note: Please don’t run start.sh as it will also start the etcd.

Step 2
~~~~~~

Edit the configuration of the ip/port on which SC will run and etcd ip
#### VM1

::

   # vi conf/app.conf
   #Replace the below values
   httpaddr = 10.12.0.1
   manager_cluster = "10.12.0.4:2379"

   # Start the Service-center
   ./service-center

VM2
^^^

::

   # vi conf/app.conf
   #Replace the below values
   httpaddr = 10.12.0.2
   manager_cluster = "10.12.0.4:2379"

   # Start the Service-center
   ./service-center

Note: In ``manger_cluster`` you can put the multiple instances of etcd
in the cluster like

::

   manager_cluster= "10.12.0.4:2379,10.12.0.X:2379,10.12.0.X:2379"

Step 3
~~~~~~

Verify your instances

::

   # curl http://10.12.0.1:30101/v4/default/registry/health
   {
       "instances": [
           {
               "instanceId": "d6e9e976f9df11e7a72b286ed488ff9f",
               "serviceId": "d6e99f4cf9df11e7a72b286ed488ff9f",
               "endpoints": [
                   "rest://10.12.0.1:30100"
               ],
               "hostName": "service_center_10_12_0_1",
               "status": "UP",
               "healthCheck": {
                   "mode": "push",
                   "interval": 30,
                   "times": 3
               },
               "timestamp": "1516012543",
               "modTimestamp": "1516012543"
           },
           {
               "instanceId": "16d4cb35f9e011e7a58a286ed488ff9f",
               "serviceId": "d6e99f4cf9df11e7a72b286ed488ff9f",
               "endpoints": [
                   "rest://10.12.0.2:30100"
               ],
               "hostName": "service_center_10_12_0_2",
               "status": "UP",
               "healthCheck": {
                   "mode": "push",
                   "interval": 30,
                   "times": 3
               },
               "timestamp": "1516012650",
               "modTimestamp": "1516012650"
           }
       ]
   }

As we can see here the Service-Center can auto-discover all the
instances of the Service-Center running in cluster, this auto-discovery
feature is used by the `Java-Chassis SDK`_ to auto-discover all the
instances of the Service-Center by knowing atleast 1 IP of
Service-Center running in cluster.

In your microservice.yaml you can provide the SC IP of both the instance
or any one instance, sdk can auto-discover other instances and use the
other instances to get microservice details in case of failure of the
first one.

::

   cse:
     service:
       registry:
         address: "http://10.12.0.1:30100,http://10.12.0.2:30100"
         autodiscovery: true

In this case sdk will be able to discover all the instances of SC in
cluster.


.. _cluster: https://github.com/coreos/etcd/blob/master/Documentation/op-guide/container.md
.. _this: https://github.com/coreos/etcd/blob/master/Documentation/op-guide/container.md
.. _here: https://github.com/apache/servicecomb-service-center/releases
.. _Java-Chassis SDK: https://github.com/apache/servicecomb-java-chassis