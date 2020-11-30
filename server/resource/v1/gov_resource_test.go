package v1_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/gov"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/resource/v1"
	svc "github.com/apache/servicecomb-service-center/server/service/gov"
	"github.com/go-chassis/go-archaius"
	"github.com/stretchr/testify/assert"

	_ "github.com/apache/servicecomb-service-center/server/service/gov/mock"
)

func init() {
	config.Configurations.Gov = &config.Gov{
		DistOptions: []config.DistributorOptions{
			{
				Name: "mock",
				Type: "mock",
			},
		},
	}
	err := svc.Init()
	if err != nil {
		log.Fatal("", err)
	}
}
func TestAuthResource_Login(t *testing.T) {
	err := archaius.Init(archaius.WithMemorySource(), archaius.WithENVSource())
	assert.NoError(t, err)

	svc.Init()
	rest.RegisterServant(&v1.Governance{})

	t.Run("create policy", func(t *testing.T) {
		b, _ := json.Marshal(&gov.LoadBalancer{
			GovernancePolicy: &gov.GovernancePolicy{Name: "test"},
			Spec: &gov.LBSpec{
				Bo: &gov.BackOffPolicy{InitialInterval: 1}}})

		r, _ := http.NewRequest(http.MethodPost, "/v1/default/gov/loadBalancer", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

}
