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

package task

import (
	"context"

	"github.com/go-chassis/cari/sync"
	"github.com/go-chassis/openlog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/eventbase/datasource/mongo/model"
)

type Dao struct {
}

func (d *Dao) Create(ctx context.Context, task *sync.Task) (*sync.Task, error) {
	collection := client.GetMongoClient().GetDB().Collection(model.CollectionTask)
	_, err := collection.InsertOne(ctx, task)
	if err != nil {
		openlog.Error("fail to create task" + err.Error())
		return nil, err
	}
	return task, nil
}

func (d *Dao) Update(ctx context.Context, task *sync.Task) error {
	collection := client.GetMongoClient().GetDB().Collection(model.CollectionTask)
	result, err := collection.UpdateOne(ctx,
		bson.M{model.ColumnTaskID: task.TaskID, model.ColumnDomain: task.Domain, model.ColumnProject: task.Project, model.ColumnTimestamp: task.Timestamp},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: model.ColumnStatus, Value: task.Status}}},
		})
	if err != nil {
		openlog.Error("fail to update task" + err.Error())
		return err
	}
	if result.ModifiedCount == 0 {
		openlog.Error("fail to update task" + datasource.ErrTaskNotExists.Error())
		return datasource.ErrTaskNotExists
	}
	return nil
}

func (d *Dao) Delete(ctx context.Context, tasks ...*sync.Task) error {
	tasksIDs := make([]string, len(tasks))
	filter := bson.A{}
	for i, task := range tasks {
		tasksIDs[i] = task.TaskID
		dFilter := bson.D{
			{model.ColumnDomain, task.Domain},
			{model.ColumnProject, task.Project},
			{model.ColumnTaskID, task.TaskID},
			{model.ColumnTimestamp, task.Timestamp},
		}
		filter = append(filter, dFilter)
	}

	var deleteFunc = func(sessionContext mongo.SessionContext) error {
		collection := client.GetMongoClient().GetDB().Collection(model.CollectionTask)
		_, err := collection.DeleteMany(sessionContext, bson.M{"$or": filter})
		return err
	}
	err := client.GetMongoClient().ExecTxn(ctx, deleteFunc)
	if err != nil {
		openlog.Error(err.Error())
	}
	return err
}
func (d *Dao) List(ctx context.Context, options ...datasource.TaskFindOption) ([]*sync.Task, error) {
	opts := datasource.NewTaskFindOptions()
	for _, o := range options {
		o(&opts)
	}
	collection := client.GetMongoClient().GetDB().Collection(model.CollectionTask)
	filter := bson.M{}
	if opts.Domain != "" {
		filter[model.ColumnDomain] = opts.Domain
	}
	if opts.Project != "" {
		filter[model.ColumnProject] = opts.Project
	}
	if opts.Action != "" {
		filter[model.ColumnAction] = opts.Action
	}
	if opts.DataType != "" {
		filter[model.ColumnDataType] = opts.DataType
	}
	if opts.Status != "" {
		filter[model.ColumnStatus] = opts.Status
	}
	opt := mopts.Find().SetSort(map[string]interface{}{
		model.ColumnTimestamp: 1,
	})
	cur, err := collection.Find(ctx, filter, opt)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	tasks := make([]*sync.Task, 0)
	for cur.Next(ctx) {
		task := &sync.Task{}
		if err := cur.Decode(task); err != nil {
			openlog.Error("decode to task error: " + err.Error())
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}
