/*
 Licensed to the Apache Software Foundation (ASF) under one
 or more contributor license agreements.  See the NOTICE file
 distributed with this work for additional information
 regarding copyright ownership.  The ASF licenses this file
 to you under the Apache License, Version 2.0 (the
 "License"); you may not use this file except in compliance
 with the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing,
 software distributed under the License is distributed on an
 "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 KIND, either express or implied.  See the License for the
 specific language governing permissions and limitations
 under the License.
*/
package aligator

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/apsystole/log"
)

var bqClient *bigquery.Client
var err error

var config Config

var projectID string

var message Message

type PubSubMessage struct {
	Data []byte `json:"data"`
}

// the message send to the next PubSub topic
type Message struct {
	GcsRef                 bigquery.GCSReference           `json:"gcs_ref"`
	Config                 Config                          `json:"config"`
	DatasetID              string                          `json:"dataset_id"`
	TableID                string                          `json:"table_id"`
	TableCreateDisposition bigquery.TableCreateDisposition `json:"Table_create_disposition"`
	TableWriteDisposition  bigquery.TableWriteDisposition  `json:"Table_write_disposition"`
	PubSubTopic            string                          `json:"pub_sub_topic"`
}

// config of the flow
type Config struct {
	Name             string                         `json:"name"`
	FileName         string                         `json:"file_name"`
	TableName        string                         `json:"table_name"`
	FieldDelimiter   string                         `json:"field_delimiter"`
	FileSchema       string                         `json:"file_schema"`
	WriteDisposition bigquery.TableWriteDisposition `json:"write_disposition"`
	CheckRejet       bool                           `json:"CheckRejet"`
	Compression      string                         `json:"compression"`
	WebhookUrl       string                         `json:"webhook_url"`
	HistoQuery       string                         `json:"histo_query"`
	PubSubTopic      string                         `json:"pubsub_topic"`
}

func init() {

	projectID = os.Getenv("GCP_PROJECT")
	bqClient, err = bigquery.NewClient(context.Background(), projectID)
	if err != nil {
		log.Errorf("error create client bigquery.NewClient: ", err)
	}
}

func LoadBQ(ctx context.Context, m PubSubMessage) error {

	err = json.Unmarshal(m.Data, &message)
	if err != nil {
		log.Errorf("Error Json request format !", err)
	}

	loader := bqClient.Dataset(message.Config.TableName).Table(strings.Split(message.Config.TableName, ".")[1]).LoaderFrom(&message.GcsRef)
	loader.CreateDisposition = bigquery.CreateIfNeeded
	loader.WriteDisposition = config.WriteDisposition
	loader.JobID = "loadBQ_" + message.Config.TableName
	loader.AddJobIDSuffix = true
	job, err := loader.Run(context.Background())
	if err != nil {
		log.Errorf("destination file is empty ")
		return err
	}
	status, err := job.Wait(context.Background())
	if err != nil {
		log.Errorf("destination file is empty ")
		return err
	}

	if status.Err() != nil {
		log.Errorf("job completed with error: %v", status.Err())
		return err
	}

	return nil

}
