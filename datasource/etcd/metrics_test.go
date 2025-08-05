package etcd

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-chassis/cari/discovery"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource"
	pkgmetrics "github.com/apache/servicecomb-service-center/pkg/metrics"
	"github.com/apache/servicecomb-service-center/server/metrics"
)

func TestMetricsManager_Report(t *testing.T) {
	m := &MetricsManager{}
	reporter := &metrics.MetaReporter{}
	err := m.Report(context.TODO(), reporter)
	assert.NoError(t, err)

	expectedLabels := map[string]string{
		"domain":           datasource.RegistryDomain,
		"project":          datasource.RegistryProject,
		"framework":        discovery.Unknown,
		"frameworkVersion": discovery.Unknown,
		"instance":         pkgmetrics.InstanceName(),
	}

	domainTotal, err := pkgmetrics.GetMetrics(metrics.KeyDomainTotal)
	assert.NoError(t, err)
	assert.NotNil(t, domainTotal)
	assert.True(t, len(domainTotal.Metric) == 1)
	assert.True(t, labelEqual(domainTotal.Metric[0].Label, map[string]string{"instance": pkgmetrics.InstanceName()}))

	serviceTotal, err := pkgmetrics.GetMetrics(metrics.KeyServiceTotal)
	assert.NoError(t, err)
	assert.NotNil(t, serviceTotal)
	assert.True(t, len(serviceTotal.Metric) == 1)
	assert.True(t, labelEqual(serviceTotal.Metric[0].Label, expectedLabels))

	instanceTotal, err := pkgmetrics.GetMetrics(metrics.KeyInstanceTotal)
	assert.NoError(t, err)
	assert.NotNil(t, instanceTotal)
	assert.True(t, len(instanceTotal.Metric) == 1)
	assert.True(t, labelEqual(instanceTotal.Metric[0].Label, expectedLabels))

	frameworkTotal, err := pkgmetrics.GetMetrics(metrics.KeyFrameworkTotal)
	assert.NoError(t, err)
	assert.True(t, len(frameworkTotal.Metric) == 1)
	assert.True(t, labelEqual(frameworkTotal.Metric[0].Label, expectedLabels))

	schemaTotal, err := pkgmetrics.GetMetrics(metrics.KeySchemaTotal)
	assert.NoError(t, err)
	assert.NotNil(t, schemaTotal)
	assert.True(t, len(schemaTotal.Metric) == 1)
	assert.True(t, labelEqual(schemaTotal.Metric[0].Label, map[string]string{
		"instance": pkgmetrics.InstanceName(),
		"domain":   datasource.RegistryDomain,
		"project":  datasource.RegistryProject}))
}

func labelEqual(pairs []*io_prometheus_client.LabelPair, expect map[string]string) bool {
	if len(expect) != len(pairs) {
		return false
	}

	labels := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		if pair == nil || pair.Name == nil || pair.Value == nil {
			return false
		}
		labels[*pair.Name] = *pair.Value
	}
	return reflect.DeepEqual(labels, expect)
}
