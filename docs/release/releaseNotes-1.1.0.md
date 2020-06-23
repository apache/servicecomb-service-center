# Release Notes
        Release Notes - Apache ServiceComb - Version service-center-1.1.0
            
<h2>        Bug
</h2>
<ul>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-744'>SCB-744</a>] -         Wrong error code returned in Find API
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-851'>SCB-851</a>] -         Can not get providers if consumer have * dependency rule
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-857'>SCB-857</a>] -         Provider rule of consumer can not be removed
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-863'>SCB-863</a>] -         build script for docker image gives an error
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-890'>SCB-890</a>] -         Lost changed event when bootstrap with embedded etcd
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-912'>SCB-912</a>] -         rest client still verify peer host when verifyPeer flag set false
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-924'>SCB-924</a>] -         Etcd cacher should re-list etcd in fixed time interval
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-927'>SCB-927</a>] -         The latest Lager is not compatible
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-929'>SCB-929</a>] -         Concurrent error in update resource APIs
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-930'>SCB-930</a>] -         Service Center Frontend stops responding in Schema test if Schema has &#39;\&quot;&#39; character in the description
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-934'>SCB-934</a>] -         Get all dependency rules will panic
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-938'>SCB-938</a>] -         Should check self presevation max ttl
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-951'>SCB-951</a>] -         Wrong help information in scctl
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-958'>SCB-958</a>] -         The instance delete event delay more than 2s
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-977'>SCB-977</a>] -         Dependencies will not be updated in 5min when micro service is changed
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-980'>SCB-980</a>] -         The dependency will be broken when commit etcd failed
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-981'>SCB-981</a>] -         Can not remove the microservice and instance properties
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-991'>SCB-991</a>] -         Optimize args parsing
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-993'>SCB-993</a>] -         Bug fixes
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-994'>SCB-994</a>] -         SC can not read the context when client using grpc api
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-1027'>SCB-1027</a>] -         Fix the core dump in SC which compiled with go1.10+
</li>
</ul>
        
<h2>        New Feature
</h2>
<ul>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-815'>SCB-815</a>] -         Support deploy in Kubernetes
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-850'>SCB-850</a>] -         Support discover instances from kubernetes cluster
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-869'>SCB-869</a>] -         SC cli tool
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-902'>SCB-902</a>] -         Support service discovery by Service Mesh
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-914'>SCB-914</a>] -         Support using scctl to download schemas
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-941'>SCB-941</a>] -         Support multiple datacenter deployment
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-949'>SCB-949</a>] -         Support access distinct kubernetes clusters
</li>
</ul>
        
<h2>        Improvement
</h2>
<ul>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-418'>SCB-418</a>] -         How to deploy a SC cluster in container environment
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-435'>SCB-435</a>] -         Add plugin document in ServiceCenter
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-792'>SCB-792</a>] -         More abundant metrics information
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-796'>SCB-796</a>] -         Update the paas-lager package
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-797'>SCB-797</a>] -         More information in dump API
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-807'>SCB-807</a>] -         Limit the topology view to only 100 microservices. 
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-808'>SCB-808</a>] -         Aut-refresh the dashboard and service-list page every 10sec
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-809'>SCB-809</a>] -         Verify the chinese version of the UI as all chinese text was translated using Google Translate
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-816'>SCB-816</a>] -         Update the protobuf version to 1.0.0
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-840'>SCB-840</a>] -         Support configable limit in buildin quota plugin
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-844'>SCB-844</a>] -         Update golang version to 1.9.2
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-848'>SCB-848</a>] -         Uses zap to replace the paas-lager
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-862'>SCB-862</a>] -         Using different environment variables in image
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-892'>SCB-892</a>] -         output plugins configs in version api
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-899'>SCB-899</a>] -         Support go1.11 module maintaining
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-901'>SCB-901</a>] -         Making service registration api idempotent
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-937'>SCB-937</a>] -         Customizable tracing sample rate
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-953'>SCB-953</a>] -         Support sync distinct Kubernetes service types to service-center
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-978'>SCB-978</a>] -         Fix translation issues for Chinese Locale on First Load
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-983'>SCB-983</a>] -         Output the QPS per domain
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-984'>SCB-984</a>] -         Add Health Check command
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-1015'>SCB-1015</a>] -         Support the forth microservice version number registration
</li>
</ul>
            
<h2>        Task
</h2>
<ul>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-720'>SCB-720</a>] -         Show the instance statistics in Dashboard and Instance List in Side Menu
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-1016'>SCB-1016</a>] -         Change git repo name
</li>
<li>[<a href='https://issues.apache.org/jira/browse/SCB-1028'>SCB-1028</a>] -         Prepare 1.1.0 Service-Center Release
</li>
</ul>
                                                                                                                                        
