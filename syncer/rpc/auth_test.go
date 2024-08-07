package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/go-chassis/v2/security/authr"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"

	"github.com/apache/servicecomb-service-center/syncer/config"
)

type testAuth struct{}

func (testAuth) Login(ctx context.Context, user string, password string, opts ...authr.LoginOption) (string, error) {
	return "", nil
}

func (testAuth) Authenticate(ctx context.Context, token string) (interface{}, error) {
	var claim map[string]interface{}
	return claim, json.Unmarshal([]byte(token), &claim)
}

func Test_auth(t *testing.T) {
	// use the custom auth plugin
	authr.Install("test", func(opts *authr.Options) (authr.Authenticator, error) {
		return testAuth{}, nil
	})
	assert.NoError(t, authr.Init(authr.WithPlugin("test")))

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		preDo   func()
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "sync rbac disables",
			preDo: func() {
				config.SetConfig(config.Config{
					Sync: &config.Sync{
						RbacEnabled: false,
					}})
			},
			args: args{
				ctx: context.Background(), // rbac disabled, empty ctx should pass the auth
			},
			wantErr: assert.NoError,
		},
		{
			name: "no header",
			preDo: func() {
				config.SetConfig(config.Config{
					Sync: &config.Sync{
						RbacEnabled: true,
					}})
			},
			args: args{
				ctx: context.Background(), // rbac enabled, empty ctx should not pass the auth
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				var errSvcErr *errsvc.Error
				ok := errors.As(err, &errSvcErr)
				assert.True(t, ok)

				return assert.Equal(t, rbac.ErrNoAuthHeader, errSvcErr.Code)
			},
		},
		{
			name: "with header but no auth header",
			args: args{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"fake": "fake"})),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				var errSvcErr *errsvc.Error
				ok := errors.As(err, &errSvcErr)
				assert.True(t, ok)

				return assert.Equal(t, rbac.ErrNoAuthHeader, errSvcErr.Code)
			},
		},
		{
			name: "auth header format error",
			args: args{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{restful.HeaderAuth: "fake"})),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, rbac.ErrInvalidHeader, err)
			},
		},
		{
			name: "wrong account nor role",
			args: args{
				ctx: metadata.NewIncomingContext(context.Background(),
					metadata.New(map[string]string{restful.HeaderAuth: `Bear {"account":"x","roles":["x"]}`})),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, errWrongAccountNorRole, err)
			},
		},
		{
			name: "valid token",
			args: args{
				ctx: metadata.NewIncomingContext(context.Background(),
					metadata.New(map[string]string{restful.HeaderAuth: `Bear {"account":"sync-user","roles":["sync-admin"]}`})),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preDo != nil {
				tt.preDo()
			}
			tt.wantErr(t, auth(tt.args.ctx), fmt.Sprintf("auth(%v)", tt.args.ctx))
		})
	}
}
