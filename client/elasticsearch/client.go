// Copyright 2021 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"strings"

	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/pkg/errors"

	"github.com/mendersoftware/reporting/model"
)

const (
	indexDevices         = "devices"
	indexDevicesTemplate = `{
	"index_patterns": ["devices-*"],
	"priority": 1,
	"template": {
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 1
		},
		"mappings": {
			"_source": {
				"enabled": true
			},
			"properties": {
				"id": {
					"type": "keyword"
				},
				"tenantID": {
					"type": "keyword"
				},
				"name": {
					"type": "keyword"
				},
				"groupName": {
					"type": "keyword"
				},
				"status": {
					"type": "keyword"
				},
				"customAttributes": {
					"type": "nested",
					"properties": {
						"name": {
							"type": "keyword"
						},
						"string": {
							"type": "keyword"
						},
						"numeric": {
							"type": "double"
						}
					}
				},
				"identityAttributes": {
					"type": "nested",
					"properties": {
						"name": {
							"type": "keyword"
						},
						"string": {
							"type": "keyword"
						},
						"numeric": {
							"type": "double"
						}
					}
				},
				"inventoryAttributes": {
					"type": "nested",
					"properties": {
						"name": {
							"type": "keyword"
						},
						"string": {
							"type": "keyword"
						},
						"numeric": {
							"type": "double"
						}
					}
				},
				"createdAt": {
					"type": "date"
				},
				"updatedAt": {
					"type": "date"
				}
			}
		}
	}
}`
)

type Client interface {
	IndexDevice(ctx context.Context, device *model.Device) error
	BulkIndexDevices(ctx context.Context, devices []*model.Device) error
	Migrate(ctx context.Context) error
	Search(ctx context.Context, query map[string]interface{}) (map[string]interface{}, error)
}

type ElasticsearchClient struct {
	addresses []string
	client    *es.Client
}

type ElasticsearchClientOption func(*ElasticsearchClient)

func WithServerAddresses(addresses []string) ElasticsearchClientOption {
	return func(c *ElasticsearchClient) {
		c.addresses = addresses
	}
}

func NewClient(opts ...ElasticsearchClientOption) (Client, error) {
	client := &ElasticsearchClient{}
	for _, opt := range opts {
		opt(client)
	}

	cfg := es.Config{
		Addresses: client.addresses,
	}
	esClient, err := es.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "invalid Elasticsearch configuration")
	}

	_, err = esClient.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect to Elasticsearch")
	}

	client.client = esClient
	return client, nil
}

func (e *ElasticsearchClient) IndexDevice(ctx context.Context, device *model.Device) error {
	req := esapi.IndexRequest{
		Index:      indexDevices + "-" + device.GetTenantID(),
		DocumentID: device.GetID(),
		Body:       esutil.NewJSONReader(device),
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return errors.Wrap(err, "failed to index")
	}
	defer res.Body.Close()

	return nil
}

type bulkAction struct {
	Index *bulkActionIndex `json:"index"`
}

type bulkActionIndex struct {
	ID    string `json:"_id"`
	Index string `json:"_index"`
}

func (e *ElasticsearchClient) BulkIndexDevices(ctx context.Context, devices []*model.Device) error {
	data := ""
	for _, device := range devices {
		actionJSON, err := json.Marshal(bulkAction{
			Index: &bulkActionIndex{
				ID:    device.GetID(),
				Index: indexDevices + "-" + device.GetTenantID(),
			},
		})
		if err != nil {
			return err
		}
		deviceJSON, err := json.Marshal(device)
		if err != nil {
			return err
		}
		data += string(actionJSON) + "\n" + string(deviceJSON) + "\n"
	}
	req := esapi.BulkRequest{
		Body: strings.NewReader(data),
	}
	res, err := req.Do(ctx, e.client)
	if err != nil {
		return errors.Wrap(err, "failed to bulk index")
	}
	defer res.Body.Close()

	return nil
}

func (e *ElasticsearchClient) Migrate(ctx context.Context) error {
	req := esapi.IndicesPutIndexTemplateRequest{
		Name: indexDevices + "-tenant1",
		Body: strings.NewReader(indexDevicesTemplate),
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return errors.Wrap(err, "failed to put the index template")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errors.New("failed to set up the index template")
	}

	return nil
}

func (e *ElasticsearchClient) Search(ctx context.Context, query map[string]interface{}) (map[string]interface{}, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	tenant, ok := ctx.Value("tenant").(string)
	if tenant == "" {
		panic("no tenant")
	}
	if !ok {
		panic("ctx")
	}

	resp, err := e.client.Search(
		e.client.Search.WithContext(ctx),
		e.client.Search.WithIndex("devices-"+tenant),
		e.client.Search.WithBody(&buf),
		e.client.Search.WithTrackTotalHits(true),
		//es.Search.WithPretty(),
	)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, errors.New(resp.String())
	}

	var ret map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}

	log.Printf("elasticsearch raw:\n%v", ret)

	return ret, nil
}
