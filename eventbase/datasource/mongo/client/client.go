/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/go-chassis/cari/db"
	"github.com/go-chassis/foundation/gopool"
	"github.com/go-chassis/openlog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/eventbase/datasource/mongo/model"
)

const (
	MongoCheckDelay     = 2 * time.Second
	HeathChekRetryTimes = 3
)

var (
	ErrOpenDbFailed  = errors.New("open db failed")
	ErrRootCAMissing = errors.New("rootCAFile is empty in config file")
)

var client *MongoClient

type MongoClient struct {
	client *mongo.Client
	db     *mongo.Database
	config *db.Config

	err       chan error
	ready     chan struct{}
	goroutine *gopool.Pool
}

func NewMongoClient(config *db.Config) {
	inst := &MongoClient{}
	if err := inst.Initialize(config); err != nil {
		openlog.Error("failed to init mongodb" + err.Error())
		inst.err <- err
	}
	client = inst
}

func (mc *MongoClient) Err() <-chan error {
	return mc.err
}

func (mc *MongoClient) Ready() <-chan struct{} {
	return mc.ready
}

func (mc *MongoClient) Close() {
	if mc.client != nil {
		if err := mc.client.Disconnect(context.TODO()); err != nil {
			openlog.Error("[close mongo client] failed disconnect the mongo client" + err.Error())
		}
	}
}

func (mc *MongoClient) Initialize(config *db.Config) (err error) {
	mc.err = make(chan error, 1)
	mc.ready = make(chan struct{})
	mc.goroutine = gopool.New()
	mc.config = config
	err = mc.newClient(context.Background())
	if err != nil {
		return
	}
	mc.startHealthCheck()
	close(mc.ready)
	return nil
}

func (mc *MongoClient) newClient(ctx context.Context) (err error) {
	clientOptions := []*options.ClientOptions{options.Client().ApplyURI(mc.config.URI)}
	clientOptions = append(clientOptions, options.Client().SetMaxPoolSize(uint64(mc.config.PoolSize)))
	if mc.config.SSLEnabled {
		if mc.config.RootCA == "" {
			err = ErrRootCAMissing
			return
		}
		pool := x509.NewCertPool()
		caCert, err := ioutil.ReadFile(mc.config.RootCA)
		if err != nil {
			err = fmt.Errorf("read ca cert file %s failed", mc.config.RootCA)
			openlog.Error("ca cert :" + err.Error())
			return err
		}
		pool.AppendCertsFromPEM(caCert)
		clientCerts := make([]tls.Certificate, 0)
		if mc.config.CertFile != "" && mc.config.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(mc.config.CertFile, mc.config.KeyFile)
			if err != nil {
				openlog.Error("load X509 keyPair failed: " + err.Error())
				return err
			}
			clientCerts = append(clientCerts, cert)
		}
		tc := &tls.Config{
			RootCAs:            pool,
			InsecureSkipVerify: !mc.config.VerifyPeer,
			Certificates:       clientCerts,
		}
		clientOptions = append(clientOptions, options.Client().SetTLSConfig(tc))
		openlog.Info("enabled ssl communication to mongodb")
	}
	mc.client, err = mongo.Connect(ctx, clientOptions...)
	if err != nil {
		openlog.Error("failed to connect to mongo" + err.Error())
		if derr := mc.client.Disconnect(ctx); derr != nil {
			openlog.Error("[init mongo client] failed to disconnect mongo clients" + err.Error())
		}
		return
	}
	mc.db = mc.client.Database(model.DBName)
	if mc.db == nil {
		return ErrOpenDbFailed
	}
	return nil
}

func (mc *MongoClient) startHealthCheck() {
	mc.goroutine.Do(mc.HealthCheck)
}

func (mc *MongoClient) HealthCheck(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			mc.Close()
			return
		case <-time.After(MongoCheckDelay):
			for i := 0; i < HeathChekRetryTimes; i++ {
				err := mc.client.Ping(context.Background(), nil)
				if err == nil {
					break
				}
				openlog.Error(fmt.Sprintf("retry to connect to mongodb %s after %s", mc.config.URI, MongoCheckDelay) + err.Error())
				select {
				case <-ctx.Done():
					mc.Close()
					return
				case <-time.After(MongoCheckDelay):
				}
			}
		}
	}
}

func GetMongoClient() *MongoClient {
	return client
}

// ExecTxn execute a transaction command
// want to abort transaction, return error in cmd fn impl, otherwise it will commit transaction
func (mc *MongoClient) ExecTxn(ctx context.Context, cmd func(sessionContext mongo.SessionContext) error) error {
	session, err := mc.client.StartSession()
	if err != nil {
		return err
	}
	if err = session.StartTransaction(); err != nil {
		return err
	}
	defer session.EndSession(ctx)
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err = cmd(sc); err != nil {
			if err = session.AbortTransaction(sc); err != nil {
				return err
			}
		} else {
			if err = session.CommitTransaction(sc); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (mc *MongoClient) GetDB() *mongo.Database {
	return mc.db
}

func (mc *MongoClient) CreateIndexes(ctx context.Context, Table string, indexes []mongo.IndexModel) error {
	_, err := mc.db.Collection(Table).Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}
	return nil
}
