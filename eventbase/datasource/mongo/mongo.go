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

package mongo

import (
	"strings"

	"github.com/go-chassis/cari/db"
	"github.com/go-chassis/openlog"
	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/mgo.v2"

	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/eventbase/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/eventbase/datasource/mongo/task"
	"github.com/apache/servicecomb-service-center/eventbase/datasource/mongo/tombstone"
)

type Datasource struct {
	taskDao   datasource.TaskDao
	tombstone datasource.TombstoneDao
}

func (d *Datasource) TaskDao() datasource.TaskDao {
	return d.taskDao
}

func (d *Datasource) TombstoneDao() datasource.TombstoneDao {
	return d.tombstone
}

func NewDatasource(config *db.Config) (datasource.DataSource, error) {
	inst := &Datasource{}
	inst.taskDao = &task.Dao{}
	inst.tombstone = &tombstone.Dao{}
	return inst, inst.initialize(config)
}

func (d *Datasource) initialize(config *db.Config) error {
	err := d.initClient(config)
	if err != nil {
		return err
	}
	ensureDB(config)
	return nil
}

func (d *Datasource) initClient(config *db.Config) error {
	client.NewMongoClient(config)
	select {
	case err := <-client.GetMongoClient().Err():
		return err
	case <-client.GetMongoClient().Ready():
		return nil
	}
}

func init() {
	datasource.RegisterPlugin("mongo", NewDatasource)
}

func ensureDB(config *db.Config) {
	session := openSession(config)
	defer session.Close()
	session.SetMode(mgo.Primary, true)

	ensureTask(session)
	ensureTombstone(session)
}

func openSession(c *db.Config) *mgo.Session {
	timeout := c.Timeout
	var err error
	session, err := mgo.DialWithTimeout(c.URI, timeout)
	if err != nil {
		openlog.Warn("can not dial db, retry once:" + err.Error())
		session, err = mgo.DialWithTimeout(c.URI, timeout)
		if err != nil {
			openlog.Fatal("can not dial db:" + err.Error())
		}
	}
	return session
}

func wrapError(err error, skipMsg ...string) {
	if err != nil {
		for _, str := range skipMsg {
			if strings.Contains(err.Error(), str) {
				openlog.Debug(err.Error())
				return
			}
		}
		openlog.Error(err.Error())
	}
}

func ensureTask(session *mgo.Session) {
	c := session.DB(model.DBName).C(model.CollectionTask)
	err := c.Create(&mgo.CollectionInfo{Validator: bson.M{
		model.ColumnID:        bson.M{"$exists": true},
		model.ColumnDomain:    bson.M{"$exists": true},
		model.ColumnProject:   bson.M{"$exists": true},
		model.ColumnTimestamp: bson.M{"$exists": true},
	}})
	wrapError(err)
	err = c.EnsureIndex(mgo.Index{
		Key:    []string{model.ColumnDomain, model.ColumnProject, model.ColumnID, model.ColumnTimestamp},
		Unique: true,
	})
	wrapError(err)
}

func ensureTombstone(session *mgo.Session) {
	c := session.DB(model.DBName).C(model.CollectionTombstone)
	err := c.Create(&mgo.CollectionInfo{Validator: bson.M{
		model.ColumnResourceID:   bson.M{"$exists": true},
		model.ColumnDomain:       bson.M{"$exists": true},
		model.ColumnProject:      bson.M{"$exists": true},
		model.ColumnResourceType: bson.M{"$exists": true},
	}})
	wrapError(err)
	err = c.EnsureIndex(mgo.Index{
		Key:    []string{model.ColumnDomain, model.ColumnProject, model.ColumnResourceID, model.ColumnResourceType},
		Unique: true,
	})
	wrapError(err)
}
