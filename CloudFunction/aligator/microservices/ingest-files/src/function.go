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
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"context"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
)

// GCSEvent is the payload of a GCS event.
type GCSEvent struct {
	Kind                    string                 `json:"kind"`
	ID                      string                 `json:"id"`
	SelfLink                string                 `json:"selfLink"`
	Name                    string                 `json:"name"`
	Bucket                  string                 `json:"bucket"`
	Generation              string                 `json:"generation"`
	Metageneration          string                 `json:"metageneration"`
	ContentType             string                 `json:"contentType"`
	TimeCreated             time.Time              `json:"timeCreated"`
	Updated                 time.Time              `json:"updated"`
	TemporaryHold           bool                   `json:"temporaryHold"`
	EventBasedHold          bool                   `json:"eventBasedHold"`
	RetentionExpirationTime time.Time              `json:"retentionExpirationTime"`
	StorageClass            string                 `json:"storageClass"`
	TimeStorageClassUpdated time.Time              `json:"timeStorageClassUpdated"`
	Size                    string                 `json:"size"`
	MD5Hash                 string                 `json:"md5Hash"`
	MediaLink               string                 `json:"mediaLink"`
	ContentEncoding         string                 `json:"contentEncoding"`
	ContentDisposition      string                 `json:"contentDisposition"`
	CacheControl            string                 `json:"cacheControl"`
	Metadata                map[string]interface{} `json:"metadata"`
	CRC32C                  string                 `json:"crc32c"`
	ComponentCount          int                    `json:"componentCount"`
	Etag                    string                 `json:"etag"`
	CustomerEncryption      struct {
		EncryptionAlgorithm string `json:"encryptionAlgorithm"`
		KeySha256           string `json:"keySha256"`
	}
	KMSKeyName    string `json:"kmsKeyName"`
	ResourceState string `json:"resourceState"`
}

// the message send to the next PubSub topic
type NextsMessages struct {
	GcsRef                 bigquery.GCSReference           `json:"gcs_ref"`
	Config                 Config                          `json:"config"`
	DatasetID              string                          `json:"dataset_id"`
	TableID                string                          `json:"table_id"`
	TableCreateDisposition bigquery.TableCreateDisposition `json:"Table_create_disposition"`
	TableWriteDisposition  bigquery.TableWriteDisposition  `json:"Table_write_disposition"`
	PubSubTopic            string                          `json:"pub_sub_topic"`
}

// struc file map table table of configs json
type FileMapTable struct {
	Files []Config `json:"files"`
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

// the var envs of the cloud function
type EnvVars struct {
	GCP_PROJECT    string
	BUCKET_PARAM   string
	MAP_FILE_TABLE string
}

var err error
var env EnvVars
var gcsClient *storage.Client
var filesTable FileMapTable

// function to init paramèters
func init() {

	ctx := context.Background()
	env.GCP_PROJECT = os.Getenv("GCP_PROJECT")
	env.BUCKET_PARAM = os.Getenv("BUCKET_PARAM")
	env.MAP_FILE_TABLE = "map_file_table.json"

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Errorf("Error create storage Client : ", err)
	}

	rc, err := gcsClient.Bucket(env.BUCKET_PARAM).Object(env.MAP_FILE_TABLE).NewReader(ctx)
	if err != nil {
		log.Errorf("the file map does not exist, or the path is incorrect : ", err)
	}
	defer rc.Close()

	slurp, err := ioutil.ReadAll(rc)
	if err != nil {
		log.Errorf("the file map does not exist, or the path is incorrect :", err)
	}

	err = json.Unmarshal([]byte(slurp), &filesTable)
	if err != nil {
		log.Errorf("error unmarshal json map file table :", err)
	}

}

func Aligator(ctx context.Context, e GCSEvent) error {
	log.Printf("run job alogator")

	/* split Path */
	log.Printf("step 0 :")
	path, filePrefix := SplitName(e.Name)
	fileName := strings.Split(e.Name, "/")
	if strings.HasPrefix(path, "from") {
		for i, config := range filesTable {
			if strings.EqualFold(config.FileName, filePrefix) {
				config.Name = e.Name
				//CallNextsMessages(config, "load_bq")
				log.Printf("send message to load bq" + e.Name)
			}

		}

	}

	return nil
}

// function to push json message to pubsub topic
func CallNextsMessages(config Config, topic string) error {

	log.Println("start push to pubsub")
	l := &GetNextsMessages{filesTable}
	e, err := json.Marshal(l)
	if err != nil {
		log.Errorf("error json marshal logEvent pubsub", err)
		log.Println(err)
		return
	}
	ctx := context.Background()
	pubSub_client, err := pubsub.NewClient(ctx, os.Getenv("GCP_PROJECT"))
	if err != nil {
		log.Errorf("error create new pusbus client", err)
		log.Println(err)
		return
	}
	log.Println("get topic +" + PubSubTopic)
	topic := pubSub_client.Topic(PubSubTopic)
	topic.Publish(ctx, &pubsub.Message{
		Data: []byte(string(e)),
	})
	log.Println(string(e))
	log.Println("end push to pubsub")

}

func Split(Name string) (string, string) {
	nameSplit := strings.Split(Name, "/")
	sl := strings.Split(nameSplit[len(nameSplit)], "_")
	sl = sl[:len(sl)-1]
	filePrefix := strings.Join(sl[:], "_")
	return nameSplit[0], filePrefix
}
