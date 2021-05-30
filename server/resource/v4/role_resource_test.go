package v4_test

import (
	_ "github.com/apache/servicecomb-service-center/test"
	"strings"

	"bytes"
	"context"
	"encoding/json"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoleResource_CreateOrUpdateRole(t *testing.T) {
	var to = &rbacmodel.Token{}
	ctx := context.TODO()
	dao.DeleteAccount(ctx, "dev_test")
	dao.DeleteRole(ctx, "tester")
	t.Run("root login", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: "Complicated_password1"})

		r, err := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		assert.NoError(t, err)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		json.Unmarshal(w.Body.Bytes(), to)
	})

	t.Run("create account dev_test and add a role", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "dev_test", Password: "Complicated_password3", Roles: []string{"tester"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/accounts", bytes.NewBuffer(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("create a role name tester ", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "dev_test", Password: "Complicated_password3", Roles: []string{"tester"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		devToken := &rbacmodel.Token{}
		json.Unmarshal(w.Body.Bytes(), devToken)

		b2, _ := json.Marshal(&rbacmodel.Role{
			Name: "tester",
			Perms: []*rbacmodel.Permission{
				{
					Resources: []*rbacmodel.Resource{{Type: "service"}, {Type: "instance"}},
					Verbs:     []string{"get", "create", "update"},
				},
			},
		})

		r2, _ := http.NewRequest(http.MethodPost, "/v4/roles", bytes.NewReader(b2))
		r2.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)

		r3, _ := http.NewRequest(http.MethodGet, "/v4/roles", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)

		b4, _ := json.Marshal(&rbacmodel.Role{
			Name: "tester",
			Perms: []*rbacmodel.Permission{
				{
					Resources: []*rbacmodel.Resource{{Type: "service"}},
					Verbs:     []string{"get", "create", "update"},
				},
			},
		})
		r4, _ := http.NewRequest(http.MethodPut, "/v4/roles/tester", bytes.NewReader(b4))
		r4.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w4 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w4, r4)
		assert.Equal(t, http.StatusOK, w4.Code)
	})

	t.Run("Inquire role", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: "Complicated_password1"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacmodel.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		r2, _ := http.NewRequest(http.MethodGet, "/v4/roles/admin", nil)
		r2.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)

	})
	t.Run("delete role, given admin role, should fail", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: "Complicated_password1"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacmodel.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		r3, _ := http.NewRequest(http.MethodDelete, "/v4/roles/admin", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusForbidden, w3.Code)
	})
}
func TestRoleResource_MoreRoles(t *testing.T) {
	var to = &rbacmodel.Token{}
	ctx := context.TODO()
	dao.DeleteAccount(ctx, "dev_test")
	dao.DeleteAccount(ctx, "dev_test2")
	dao.DeleteRole(ctx, "tester")
	dao.DeleteRole(ctx, "tester2")
	t.Run("root login", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: "Complicated_password1"})

		r, err := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		if err != nil {
			t.Error(err)
		}
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		json.Unmarshal(w.Body.Bytes(), to)
	})

	t.Run("create role name tester", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Role{
			Name: "tester",
			Perms: []*rbacmodel.Permission{
				{
					Resources: []*rbacmodel.Resource{{Type: "service"}},
					Verbs:     []string{"get", "create", "update"},
				},
			},
		})

		r, _ := http.NewRequest(http.MethodPost, "/v4/roles", bytes.NewReader(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("create new role name tester2", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Role{
			Name: "tester2",
			Perms: []*rbacmodel.Permission{
				{
					Resources: []*rbacmodel.Resource{{Type: "rule"}},
					Verbs:     []string{"*"},
				},
			},
		})

		r, _ := http.NewRequest(http.MethodPost, "/v4/roles", bytes.NewReader(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		r2, _ := http.NewRequest(http.MethodGet, "/v4/roles", nil)
		r2.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)
	})

	t.Run("account dev_test2 support more than 1 role ", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "dev_test2", Password: "Complicated_password3", Roles: []string{"tester", "tester2"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/accounts", bytes.NewBuffer(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		r2, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)
		devToken := &rbacmodel.Token{}
		json.Unmarshal(w2.Body.Bytes(), devToken)

		r3, _ := http.NewRequest(http.MethodGet, "/v4/default/registry/microservices", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+devToken.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)

		reader := strings.NewReader("{\n  \"serviceIds\": [\n    \"MOCK\"\n  ]\n}")
		r4, _ := http.NewRequest(http.MethodDelete, "/v4/default/registry/microservices", reader)
		r4.Header.Set(restful.HeaderAuth, "Bearer "+devToken.TokenStr)
		w4 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w4, r4)
		assert.Equal(t, http.StatusUnauthorized, w4.Code)
	})
}
