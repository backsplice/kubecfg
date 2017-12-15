// Copyright 2017 The kubecfg authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package kubecfg

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"

	isatty "github.com/mattn/go-isatty"
	log "github.com/sirupsen/logrus"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"

	"github.com/ksonnet/kubecfg/utils"
)

var ErrDiffFound = fmt.Errorf("Differences found.")

// DiffCmd represents the diff subcommand
type DiffCmd struct {
	ClientPool       dynamic.ClientPool
	Discovery        discovery.DiscoveryInterface
	DefaultNamespace string

	DiffStrategy string
}

func (c DiffCmd) Run(apiObjects []*unstructured.Unstructured, out io.Writer) error {
	sort.Sort(utils.AlphabeticalOrder(apiObjects))

	diffFound := false
	for _, obj := range apiObjects {
		desc := fmt.Sprintf("%s %s", utils.ResourceNameFor(c.Discovery, obj), utils.FqName(obj))
		log.Debugf("Fetching ", desc)

		client, err := utils.ClientForResource(c.ClientPool, c.Discovery, obj, c.DefaultNamespace)
		if err != nil {
			return err
		}

		liveObj, err := client.Get(obj.GetName(), metav1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			log.Debugf("%s doesn't exist on the server", desc)
			liveObj = nil
		} else if err != nil {
			return fmt.Errorf("Error fetching %s: %v", desc, err)
		}

		fmt.Fprintln(out, "---")
		fmt.Fprintf(out, "- live %s\n+ config %s\n", desc, desc)
		if liveObj == nil {
			fmt.Fprintf(out, "%s doesn't exist on server\n", desc)
			diffFound = true
			continue
		}

		liveObjObject := liveObj.Object
		if c.DiffStrategy == "subset" {
			liveObjObject = removeMapFields(obj.Object, liveObjObject)
		}
		diff := gojsondiff.New().CompareObjects(liveObjObject, obj.Object)

		if diff.Modified() {
			diffFound = true
			fcfg := formatter.AsciiFormatterConfig{
				Coloring: istty(out),
			}
			formatter := formatter.NewAsciiFormatter(liveObjObject, fcfg)
			text, err := formatter.Format(diff)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "%s", text)
		} else {
			fmt.Fprintf(out, "%s unchanged\n", desc)
		}
	}

	if diffFound {
		return ErrDiffFound
	}
	return nil
}

// isEmptyValue is taken from the encoding/json package in the
// standard library.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func removeFields(config, live interface{}) interface{} {
	switch c := config.(type) {
	case map[string]interface{}:
		return removeMapFields(c, live.(map[string]interface{}))
	case []interface{}:
		return removeListFields(c, live.([]interface{}))
	default:
		return live
	}
}

func removeMapFields(config, live map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for k, v1 := range config {
		v2, ok := live[k]
		if !ok {
			if isEmptyValue(reflect.ValueOf(v1)) {
				result[k] = v1
			}
			continue
		}
		result[k] = removeFields(v1, v2)
	}
	return result
}

func removeListFields(config, live []interface{}) []interface{} {
	// If live is longer than config, then the extra elements at the end of the
	// list will be returned as is so they appear in the diff.
	result := make([]interface{}, 0, len(live))
	for i, v2 := range live {
		if len(config) > i {
			result = append(result, removeFields(config[i], v2))
		} else {
			result = append(result, v2)
		}
	}
	return result
}

func istty(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return isatty.IsTerminal(f.Fd())
	}
	return false
}
