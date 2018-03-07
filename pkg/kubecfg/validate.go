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

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/discovery"

	"github.com/ksonnet/kubecfg/utils"
)

// ValidateCmd represents the validate subcommand
type ValidateCmd struct {
	Discovery     discovery.DiscoveryInterface
	IgnoreUnknown bool
}

// newGroupVersionChecker returns a predicate that returns true if the group version is known to the server.
func newGroupVersionChecker(d discovery.ServerGroupsInterface) (func(schema.GroupVersion) bool, error) {
	groupList, err := d.ServerGroups()
	if err != nil {
		return nil, err
	}

	groupVersions := sets.NewString()
	for _, gv := range v1.ExtractGroupVersions(groupList) {
		groupVersions.Insert(gv.String())
	}
	
	return func(gv schema.GroupVersion) bool {
		return groupVersions.Has(gv.String())
	}, nil
}

func (c ValidateCmd) Run(apiObjects []*unstructured.Unstructured, out io.Writer) error {
	isGroupVersionKnown, err := newGroupVersionChecker(c.Discovery)
	if err != nil {
		return err
	}

	hasError := false

	for _, obj := range apiObjects {
		desc := fmt.Sprintf("%s %s", utils.ResourceNameFor(c.Discovery, obj), utils.FqName(obj))
		log.Info("Validating ", desc)

		gv := obj.GroupVersionKind().GroupVersion()
		if c.IgnoreUnknown && !isGroupVersionKnown(gv) {
			log.Warnf("Skipping validation of %s because schema for %s is unknown", desc, gv)
			continue
		}

		var allErrs []error

		schema, err := utils.NewSwaggerSchemaFor(c.Discovery, gv)
		if err != nil {
			allErrs = append(allErrs, fmt.Errorf("Unable to fetch schema: %v", err))
		} else {
			// Validate obj
			allErrs = append(allErrs, schema.Validate(obj)...)
		}

		for _, err := range allErrs {
			log.Errorf("Error in %s: %v", desc, err)
			hasError = true
		}
	}

	if hasError {
		return fmt.Errorf("Validation failed")
	}

	return nil
}
