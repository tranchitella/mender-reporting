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
package reporting

import (
	"context"
	"encoding/json"
	"log"

	"github.com/mendersoftware/reporting/client/elasticsearch"

	"github.com/mendersoftware/reporting/model"
)

type App interface {
	// return 'inventory' compatible device list
	SearchDevices(ctx context.Context, searchParams *model.SearchParams) ([]interface{}, int, error)
	// return the raw output from ES instead of 'inventory' devices
	DebugSearchDevicesRawES(ctx context.Context, searchParams *model.SearchParams) (interface{}, error)
}

type app struct {
	esClient elasticsearch.Client
}

func NewApp(esClient elasticsearch.Client) App {
	return &app{
		esClient: esClient,
	}
}

func (app *app) doSearchDevices(ctx context.Context, searchParams *model.SearchParams) (map[string]interface{}, error) {
	// search
	//
	// {
	//   "query": {
	//     "bool": {
	//       "must_not": [...nested queries...],
	//		 "must": [...nested queries...],
	//
	//       "from":
	// 		 "size":
	//
	//		 "sort": [...nested sort queries...]
	// }}}
	must := []interface{}{}
	mustNot := []interface{}{}

	for _, f := range searchParams.Filters {
		m, mn := toQuery(f)
		if m != nil {
			must = append(must, m)
		} else if mn != nil {
			mustNot = append(mustNot, mn)
		}
	}

	qBool := map[string]interface{}{}

	if len(must) > 0 {
		qBool["must"] = must
	}
	if len(mustNot) > 0 {
		qBool["must_not"] = mustNot
	}

	// sort
	s := sorts(searchParams.Sort)

	// page
	size := searchParams.PerPage
	from := (searchParams.Page - 1) * size

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": qBool,
		},
		"sort": s,
		"from": from,
		"size": size,
	}

	res, err := app.esClient.Search(ctx, query)

	enc, err := json.MarshalIndent(query, "", "  ")
	log.Printf("query: \n%v", string(enc))

	return res, err
}

func (app *app) SearchDevices(ctx context.Context, searchParams *model.SearchParams) ([]interface{}, int, error) {

	esDevs, err := app.doSearchDevices(ctx, searchParams)
	if err != nil {
		return nil, 0, err
	}

	// 'select' arg is implemented here, in the app layer
	// ES '_source' and 'fields' work on field names,
	// doesn't seem possible to use values ('Attributes.name') to project responses

	return toInventory(esDevs, searchParams.Attributes),
		int(esDevs["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		err
}

func (app *app) DebugSearchDevicesRawES(ctx context.Context, searchParams *model.SearchParams) (interface{}, error) {
	res, err := app.doSearchDevices(ctx, searchParams)
	if err != nil {
		return nil, err
	}

	return res, err
}

// toQuery returns the (nested) attribute filter
// returns ('must', 'must_not') to signal which part of the toplevel query the filter contributes to
func toQuery(f model.FilterPredicate) (map[string]interface{}, map[string]interface{}) {
	//
	// {
	//   "nested": {
	//	   "path": "inventoryAttributes",
	//	   "query": {
	//	     "bool": {
	//		   "must": [
	//		     {"match": ... },
	//			 ...
	//		   ]
	//		 }
	//	   }
	//   }
	// }

	qscope := f.Scope + "Attributes"

	qbool := map[string]interface{}{}

	switch f.Type {
	case "$eq":
		// doc says 'any' value allowed, do we accept arrays for exact comparison?
		field, _ := assertValue(f.Value)
		qbool["must"] = []map[string]interface{}{
			map[string]interface{}{
				"match": map[string]interface{}{
					qscope + ".name": f.Attribute,
				},
			},
			map[string]interface{}{
				"match": map[string]interface{}{
					qscope + field: f.Value,
				},
			},
		}

		return nested(qscope, qbool), nil
	case "$ne":
		field, _ := assertValue(f.Value)
		qbool["must"] = []map[string]interface{}{
			map[string]interface{}{
				"match": map[string]interface{}{
					qscope + ".name": f.Attribute,
				},
			},
			map[string]interface{}{
				"match": map[string]interface{}{
					qscope + field: f.Value,
				},
			},
		}
		return nil, nested(qscope, qbool)
	case "$regex":
		if _, ok := f.Value.(string); !ok {
			panic("value must be string")
		}

		qbool["must"] = []map[string]interface{}{
			map[string]interface{}{
				"match": map[string]interface{}{
					qscope + ".name": f.Attribute,
				},
			},
			map[string]interface{}{
				"regexp": map[string]interface{}{
					qscope + ".string": f.Value,
				},
			},
		}
		return nested(qscope, qbool), nil
	case "$in":
		field, arr := assertValue(f.Value)
		if !arr {
			panic("value must be an array")
		}

		qbool["must"] = []map[string]interface{}{
			map[string]interface{}{
				"match": map[string]interface{}{
					qscope + ".name": f.Attribute,
				},
			},
			map[string]interface{}{
				"terms": map[string]interface{}{
					qscope + field: f.Value,
				},
			},
		}
		return nested(qscope, qbool), nil
	case "$nin":
		field, arr := assertValue(f.Value)
		if !arr {
			panic("value must be an array")
		}

		qbool["must"] = []map[string]interface{}{
			map[string]interface{}{
				"match": map[string]interface{}{
					qscope + ".name": f.Attribute,
				},
			},
			map[string]interface{}{
				"terms": map[string]interface{}{
					qscope + field: f.Value,
				},
			},
		}
		return nil, nested(qscope, qbool)
	case "$exists":
		val, ok := f.Value.(bool)
		if !ok {
			panic("value must be bool")
		}

		qbool["must"] = []map[string]interface{}{
			map[string]interface{}{
				"match": map[string]interface{}{
					qscope + ".name": f.Attribute,
				},
			}}

		qbool["should"] = []map[string]interface{}{
			map[string]interface{}{
				"exists": map[string]interface{}{
					"field": qscope + ".string",
				},
			},
			map[string]interface{}{
				"exists": map[string]interface{}{
					"field": qscope + ".numeric",
				},
			},
		}

		qbool["minimum_should_match"] = 1

		if val {
			return nested(qscope, qbool), nil
		} else {
			return nil, nested(qscope, qbool)
		}
	case "$gt", "$gte", "$lt", "$lte":
		// doc says 'any' value allowed, does it even make sense for arrays?
		field, arr := assertValue(f.Value)
		if arr {
			panic("value must not be an array")
		}
		op := f.Type[1:]
		qbool["must"] = []map[string]interface{}{
			map[string]interface{}{
				"match": map[string]interface{}{
					qscope + ".name": f.Attribute,
				},
			},
			map[string]interface{}{
				"range": map[string]interface{}{
					qscope + field: map[string]interface{}{
						op: f.Value,
					},
				},
			},
		}
		return nested(qscope, qbool), nil
	}

	return nil, nil
}

// nested preps the 'nested' query block based on attribute 'bool' query
func nested(attrPath string, qbool interface{}) map[string]interface{} {
	return map[string]interface{}{
		"nested": map[string]interface{}{
			"path": attrPath,
			"query": map[string]interface{}{
				"bool": qbool,
			},
		},
	}
}

func sorts(crit []model.SortCriteria) []map[string]interface{} {
	sorts := []map[string]interface{}{}

	for _, c := range crit {
		scope := c.Scope + "Attributes"

		sstr := sort(scope, c.Attribute, c.Order, "string")
		snum := sort(scope, c.Attribute, c.Order, "numeric")

		sorts = append(sorts, sstr, snum)
	}

	return sorts
}

func sort(scope, name, ord, typ string) map[string]interface{} {
	return map[string]interface{}{
		scope + "." + typ: map[string]interface{}{
			"mode":  "max",
			"order": ord,
			"nested": map[string]interface{}{
				"path": scope,
				"filter": map[string]interface{}{
					"term": map[string]interface{}{
						scope + ".name": name,
					},
				},
			},
		},
	}
}

func toInventory(esDevs map[string]interface{}, sel []model.SelectAttribute) []interface{} {
	devs := []interface{}{}

	for _, v := range esDevs["hits"].(map[string]interface{})["hits"].([]interface{}) {
		devs = append(devs, toInventoryDevice(v.(map[string]interface{}), sel))
	}

	return devs
}

func toInventoryDevice(esDev map[string]interface{}, sel []model.SelectAttribute) interface{} {
	d := esDev["_source"].(map[string]interface{})

	attrs := []model.InvDeviceAttribute{}

	scopes := []string{"custom", "inventory", "identity"}
	for _, s := range scopes {
		esAttrs := d[s+"Attributes"].([]interface{})

		for _, esa := range esAttrs {
			esaMap := esa.(map[string]interface{})

			name := esaMap["name"].(string)

			if !isSelected(name, s, sel) {
				continue
			}

			var val interface{}
			if esaMap["string"] != nil {
				val = esaMap["string"]
			}

			if esaMap["numeric"] != nil {
				val = esaMap["numeric"]
			}

			valarr := val.([]interface{})

			// reduce 1-len arrays to a single elem for api compatibility
			if len(valarr) == 1 {
				val = valarr[0]
			}

			attrs = append(attrs, model.InvDeviceAttribute{
				Name:  esaMap["name"].(string),
				Scope: s,
				Value: val,
			})
		}
	}

	ret := model.InvDevice{
		ID:         model.DeviceID(d["id"].(string)),
		Attributes: attrs,
		//TODO UpdatedTs
	}

	return ret
}

func isSelected(name, scope string, sel []model.SelectAttribute) bool {

	if len(sel) == 0 {
		return true
	}

	for _, s := range sel {
		if s.Scope == scope && s.Attribute == name {
			return true
			log.Println("TRUE")
		}
	}

	return false
}

// assertValue() maps attribute's filter value type to search field type
// returns field type, is_array, or panics on disallowed type
func assertValue(value interface{}) (string, bool) {
	if _, ok := value.(string); ok {
		return ".string", false
	}
	if _, ok := value.(float64); ok {
		return ".numeric", false
	}

	// values can also be arrays - peek type of first elem
	if v, ok := value.([]interface{}); ok {
		for _, val := range v {
			switch val.(type) {
			case string:
				return ".string", true
			case float64:
				return ".numeric", true
			}
		}
	}

	// this should be normal error handling
	panic("disallowed value type")
}
