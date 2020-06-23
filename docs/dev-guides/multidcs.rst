Multiple Datacenters
====================

ServiceCenter Aggregate Architecture
------------------------------------

Now, service center has supported multiple datacenters deployment. Its
architecture likes below.

.. figure:: multidcs.PNG
   :alt: architecture

   architecture

As shown in the figure, we deploy an SC(Service-Center) cluster
independently under each DC(datacenter). Each SC cluster manages the
micro-service instances under the DC under which it belongs, and the DCs
are isolated from each other. Another implementation of the discovery
plug-in, ``Service-Center Aggregate`` service, can access multiple SC
instances and periodically pull up micro-service instance information so
that if some micro-services can request aggregate, cross-DCs can be
implemented using the same API as SC cluster.

If SC aggregate is not deployed globally, SC also supports another way
to implement multiple DCs discovery, as shown below.

.. figure:: multidcs2.PNG
   :alt: architecture

   architecture

The difference between the two approaches is that global deployment
aggregate can divert service discovery traffic, the whole architecture
is more like a read-write separation architecture, and the SC of each DC
manage microservice information independently, which reduces the
complexity. So we recommend the first architecture.

Quick Start
-----------

Let’s assume you want to install 2 clusters of Service-Center in
different DCs with following details.

======= ========== =========
Cluster Datacenter Address
======= ========== =========
sc-1    dc-1       10.12.0.1
sc-2    dc-2       10.12.0.2
======= ========== =========

you can follow `this`_ guide to install Service-Center in cluster mode.
After that, we can deploy a Service-Center Aggregate service now.

Start Service-Center Aggregate
''''''''''''''''''''''''''''''

Edit the configuration of the ip/port on which SC aggregate will run, we
assume you launch it at 10.12.0.3.

.. code:: bash

   vi conf/app.conf
   # Replace the below values
   httpaddr = 10.12.0.3
   discovery_plugin = servicecenter
   registry_plugin = buildin
   self_register = 0
   manager_cluster = "sc-1=http://10.12.0.1:30100,sc-2=http://10.12.0.2:30100"

   # Start the Service-center
   ./service-center

Note: Please don’t run start.sh as it will also start the etcd.

Confirm the service is OK
'''''''''''''''''''''''''

We recommend that you use `scctl`_, and using `cluster command`_
which makes it very convenient to verify OK.

.. code:: bash

   scctl --addr http://10.12.0.3:30100 get cluster
   #   CLUSTER |        ENDPOINTS
   # +---------+-------------------------+
   #   sc-1    | http://10.12.0.1:30100
   #   sc-2    | http://10.12.0.2:30100

Example
-------

Here we show a ``golang`` example of multiple datacenters access, where
we use an `example`_ of the `go-chassis`_ project, assuming that below.

============ ========== =========
Microservice Datacenter Address
============ ========== =========
Client       dc-1       10.12.0.4
Server       dc-2       10.12.0.5
============ ========== =========

Notes: ``go-chassis`` application can run perfectly in the above 2
architectures. If you are using `java-chassis`_, there are only support
the service center with the second architecture at the moment. You can
ref to `here`_ for more details of the second architecture.

Start Server
''''''''''''

Edit the configuration of the ip/port on which Server will register.

::

   vi examples/discovery/server/conf/chassis.yaml

Replace the below values

.. code:: yaml

   cse:
     service:
       registry:
         type: servicecenter
         address: http://10.12.0.2:30100 # the address of SC in dc-2

Run the Server

.. code:: bash

   go run examples/discovery/server/main.go

Confirm the multiple datacenters discovery is OK
''''''''''''''''''''''''''''''''''''''''''''''''

Since client is not a service, we check its running log.

::

   2018-09-29 10:30:25.556 +08:00 INFO registry/bootstrap.go:69 Register [Client] success
   ...
   2018-09-29 10:30:25.566 +08:00 WARN servicecenter/servicecenter.go:324 55c783c5c38e11e8951f0a58ac00011d Get instances from remote, key: default Server
   2018-09-29 10:30:25.566 +08:00 INFO client/client_manager.go:86 Create client for highway:Server:127.0.0.1:8082
   ...
   2018/09/29 10:30:25 AddEmploy ------------------------------ employList:<name:"One" phone:"15989351111" >


.. _this: ../user-guides/sc-cluster.rst
.. _scctl: ../intro/scctl.md
.. _cluster command: https://github.com/apache/servicecomb-service-center/blob/master/scctl/pkg/plugin/README.md#cluster-options
.. _example: https://github.com/go-chassis/go-chassis/tree/master/examples/discovery
.. _go-chassis: https://github.com/go-chassis/go-chassis
.. _java-chassis: https://github.com/apache/servicecomb-java-chassis
.. _here: multidcs2.rst

