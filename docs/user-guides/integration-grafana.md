# Integrate with Grafana

As Service-Center uses Prometheus lib to report metrics.
Then it is easy to integrate with Grafana.
Here is a [DEMO](https://github.com/apache/servicecomb-service-center/blob/master/examples/infrastructures/docker) to deploy Service-Center with Grafana, 
and this is the [template](https://github.com/apache/servicecomb-service-center/blob/master/integration/health-metrics-grafana.json) file 
can be imported in Grafana.

After the import, you can get the view like blow.

```eval_rst
.. image:: integration-grafana.PNG
```

Note: As the template has an ASF header, please remove the header first
if you import this template file.