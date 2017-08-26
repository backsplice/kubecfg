package prototype

import (
	"sort"
	"testing"
)

const (
	unmarshalErrPattern = "Expected value of %s: '%s', got: '%s'"
)

var simpleService = `{
  "apiVersion": "0.1",
  "name": "io.ksonnet.pkg.simple-service",
  "template": {
    "description": "Generates a simple service with a port exposed",
    "body": [
      "local k = import 'ksonnet.beta.2/k.libsonnet';",
      "",
      "local service = k.core.v1.service;",
      "local servicePort = k.core.v1.service.mixin.spec.portsType;",
      "local port = servicePort.new(std.extVar('port'), std.extVar('portName'));",
      "",
      "local name = std.extVar('name');",
      "k.core.v1.service.new('%-service' % name, {app: name}, port)"
    ]
  }
}`

var simpleDeployment = `{
  "apiVersion": "0.1",
  "name": "io.ksonnet.pkg.simple-deployment",
  "template": {
    "description": "Generates a simple service with a port exposed",
    "body": [
      "local k = import 'ksonnet.beta.2/k.libsonnet';",
      "local deployment = k.apps.v1beta1.deployment;",
      "local container = deployment.mixin.spec.template.spec.containersType;",
      "",
      "local appName = std.extVar('name');",
      "local appContainer = container.new(appName, std.extVar('image'));",
      "deployment.new(appName, std.extVar('replicas'), appContainer, {app: appName})"
    ]
  }
}`

func unmarshal(t *testing.T, bytes []byte) *Specification {
	p, err := Unmarshal(bytes)
	if err != nil {
		t.Fatalf("Failed to deserialize prototype:\n%v", err)
	}

	return p
}

func assertProp(t *testing.T, name string, expected string, actual string) {
	if actual != expected {
		t.Errorf(unmarshalErrPattern, name, expected, actual)
	}
}

func TestSimpleUnmarshal(t *testing.T) {
	p := unmarshal(t, []byte(simpleService))

	assertProp(t, "apiVersion", p.APIVersion, "0.1")
	assertProp(t, "name", p.Name, "io.ksonnet.pkg.simple-service")
	assertProp(t, "description", p.Template.Description, "Generates a simple service with a port exposed")
}

var testPrototypes = map[string]string{
	"io.ksonnet.pkg.simple-service": simpleService,
}

func assertSearch(t *testing.T, idx Index, opts SearchOptions, query string, expectedNames []string) {
	ps, err := idx.SearchNames(query, opts)
	if err != nil {
		t.Fatalf("Failed to search index:\n%v", err)
	}

	sort.Slice(ps, func(i, j int) bool {
		return ps[i].Name < ps[j].Name
	})

	actualNames := []string{}
	for _, p := range ps {
		actualNames = append(actualNames, p.Name)
	}

	sort.Slice(expectedNames, func(i, j int) bool {
		return expectedNames[i] < expectedNames[j]
	})

	if len(expectedNames) != len(ps) {
		t.Fatalf("Query '%s' returned results:\n%s, but expected:\n%s", query, actualNames, expectedNames)
	}

	for i := 0; i < len(expectedNames); i++ {
		if actualNames[i] != expectedNames[i] {
			t.Fatalf("Query '%s' returned results:\n%s, but expected:\n%s", query, actualNames, expectedNames)
		}
	}
}

func TestSearch(t *testing.T) {
	svc := unmarshal(t, []byte(simpleService))
	depl := unmarshal(t, []byte(simpleDeployment))
	idx := NewIndex([]*Specification{svc, depl})

	// Prefix searches.
	assertSearch(t, idx, Prefix, "service", []string{})
	assertSearch(t, idx, Prefix, "simple", []string{})
	assertSearch(t, idx, Prefix, "io.ksonnet", []string{"io.ksonnet.pkg.simple-service", "io.ksonnet.pkg.simple-deployment"})
	assertSearch(t, idx, Prefix, "foo", []string{})

	// Suffix searches.
	assertSearch(t, idx, Suffix, "service", []string{"io.ksonnet.pkg.simple-service"})
	assertSearch(t, idx, Suffix, "simple", []string{})
	assertSearch(t, idx, Suffix, "io.ksonnet", []string{})
	assertSearch(t, idx, Suffix, "foo", []string{})

	// Substring searches.
	assertSearch(t, idx, Substring, "service", []string{"io.ksonnet.pkg.simple-service"})
	assertSearch(t, idx, Substring, "simple", []string{"io.ksonnet.pkg.simple-service", "io.ksonnet.pkg.simple-deployment"})
	assertSearch(t, idx, Substring, "io.ksonnet", []string{"io.ksonnet.pkg.simple-service", "io.ksonnet.pkg.simple-deployment"})
	assertSearch(t, idx, Substring, "foo", []string{})
}
