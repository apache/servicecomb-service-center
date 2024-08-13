package rpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/go-chassis/v2/security/authr"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"google.golang.org/grpc/metadata"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/config"
)

var errWrongAccountNorRole = fmt.Errorf("account should be %s, and roles should contain %s", RbacAllowedAccountName, RbacAllowedRoleName)

func auth(ctx context.Context) error {
	if !config.GetConfig().Sync.RbacEnabled {
		return nil
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return rbac.NewError(rbac.ErrNoAuthHeader, "")
	}

	authHeader := md.Get(restful.HeaderAuth)
	if len(authHeader) == 0 {
		return rbac.NewError(rbac.ErrNoAuthHeader, fmt.Sprintf("header %s not found nor content empty", restful.HeaderAuth))
	}

	s := strings.Split(authHeader[0], " ")
	if len(s) != 2 {
		return rbac.ErrInvalidHeader
	}
	to := s[1]

	claims, err := authr.Authenticate(ctx, to)
	if err != nil {
		return err
	}
	m, ok := claims.(map[string]interface{})
	if !ok {
		log.Error("claims convert failed", rbac.ErrConvert)
		return rbac.ErrConvert
	}
	account, err := rbac.GetAccount(m)
	if err != nil {
		log.Error("get account from token failed", err)
		return err
	}

	if account.Name != RbacAllowedAccountName {
		return errWrongAccountNorRole
	}
	for _, role := range account.Roles {
		if role == RbacAllowedRoleName {
			return nil
		}
	}
	return errWrongAccountNorRole
}
