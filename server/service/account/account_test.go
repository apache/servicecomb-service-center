package account_test

import (
	"context"

	"testing"
	"time"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/server/service/account"
	"github.com/stretchr/testify/assert"
)

func TestIsBanned(t *testing.T) {
	t.Run("ban a key and check status, it should be banned, check other key should not be banned",
		func(t *testing.T) {
			err := account.Ban(context.TODO(), "dev_guy::127.0.0.1")
			assert.NoError(t, err)

			ok, err := account.IsBanned(context.TODO(), "dev_guy::127.0.0.1")
			assert.NoError(t, err)
			assert.True(t, ok)

			ok, err = account.IsBanned(context.TODO(), "test_guy::127.0.0.1")
			assert.NoError(t, err)
			assert.False(t, ok)

			time.Sleep(4 * time.Second)
			ok, err = account.IsBanned(context.TODO(), "dev_guy::127.0.0.1")
			assert.NoError(t, err)
			assert.False(t, ok)
		})
}
