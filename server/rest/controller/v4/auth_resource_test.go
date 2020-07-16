package v4_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	v4 "github.com/apache/servicecomb-service-center/server/rest/controller/v4"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
	"github.com/astaxie/beego"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/security/secret"
	"github.com/go-chassis/go-chassis/server/restful"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/apache/servicecomb-service-center/server/handler/auth"
	_ "github.com/apache/servicecomb-service-center/test"
)

func init() {
	beego.AppConfig.Set("rbac_enabled", "true")
	beego.AppConfig.Set(rbac.PubFilePath, "./rbac.pub")
	beego.AppConfig.Set("rbac_rsa_private_key_file", "./private.key")
}
func TestAuthResource_Login(t *testing.T) {
	err := archaius.Init(archaius.WithMemorySource(), archaius.WithENVSource())
	assert.NoError(t, err)

	pri, pub, err := secret.GenRSAKeyPair(4096)
	assert.NoError(t, err)

	b, err := secret.RSAPrivate2Bytes(pri)
	assert.NoError(t, err)
	ioutil.WriteFile("./private.key", b, 0600)
	b, err = secret.RSAPublicKey2Bytes(pub)
	err = ioutil.WriteFile("./rbac.pub", b, 0600)
	assert.NoError(t, err)

	archaius.Set(rbac.InitPassword, "Complicated_password1")

	ctx := context.TODO()
	dao.DeleteAccount(ctx, "root")
	archaius.Init(archaius.WithMemorySource())

	rbac.Init()
	rest.RegisterServant(&v4.AuthResource{})

	dao.DeleteAccount(ctx, "dev_account")

	t.Run("invalid user login", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "dev_account", Password: "Complicated_password1"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
	err = dao.CreateAccount(ctx, &rbacframe.Account{Name: "dev_account",
		Password: "Complicated_password1",
		Role:     "developer"})
	assert.NoError(t, err)

	t.Run("root login", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "root", Password: "Complicated_password1"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("invalid password", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "root", Password: "Complicated_password"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("dev_account login and change pwd,then login again", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "dev_account", Password: "Complicated_password1"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacframe.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		b2, _ := json.Marshal(&rbacframe.Account{CurrentPassword: "Complicated_password1", Password: "Complicated_password2"})
		r, _ = http.NewRequest(http.MethodPost, "/v4/account/dev_account/password", bytes.NewBuffer(b2))
		r.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w = httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		b3, _ := json.Marshal(&rbacframe.Account{Name: "dev_account", Password: "Complicated_password2"})
		r, _ = http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b3))
		w = httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

}
func TestAuthResource_DeleteAccount(t *testing.T) {
	t.Run("dev_account can not even delete him self", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "dev_account", Password: "Complicated_password2"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacframe.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		r2, _ := http.NewRequest(http.MethodDelete, "/v4/account/dev_account", nil)
		r2.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusUnauthorized, w2.Code)
	})
	t.Run("root can delete account", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "root", Password: "Complicated_password1"})
		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacframe.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		b, _ = json.Marshal(&rbacframe.Account{Name: "delete_account", Password: "Complicated_password1"})
		r2, _ := http.NewRequest(http.MethodPost, "/v4/account", bytes.NewBuffer(b))
		r2.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)

		r3, _ := http.NewRequest(http.MethodDelete, "/v4/account/delete_account", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusNoContent, w3.Code)
	})
}
func TestAuthResource_GetAccount(t *testing.T) {
	t.Run("get account", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "root", Password: "Complicated_password1"})
		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacframe.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		r3, _ := http.NewRequest(http.MethodGet, "/v4/account/dev_account", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)

		a := &rbacframe.Account{}
		json.Unmarshal(w3.Body.Bytes(), a)
		assert.Equal(t, "dev_account", a.Name)
		assert.Equal(t, "developer", a.Role)
		assert.Empty(t, a.Password)
	})
	t.Run("list account", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "root", Password: "Complicated_password1"})
		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacframe.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		r3, _ := http.NewRequest(http.MethodGet, "/v4/account", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)

		a := &rbacframe.AccountResponse{}
		json.Unmarshal(w3.Body.Bytes(), a)
		assert.Greater(t, a.Total, int64(2))
		assert.Empty(t, a.Accounts[0].Password)
	})

	t.Run("get a short expiration token", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "root", Password: "Complicated_password1", TokenExpirationTime: "10s"})
		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacframe.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		time.Sleep(11 * time.Second)
		r3, _ := http.NewRequest(http.MethodGet, "/v4/account", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusUnauthorized, w3.Code)
	})
}
func TestAuthResource_Login2(t *testing.T) {
	t.Run("bock user dev_account", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "dev_account", Password: "Complicated_password1"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		r, _ = http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		r, _ = http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		r, _ = http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		w = httptest.NewRecorder()
		r, _ = http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)

	})
}
