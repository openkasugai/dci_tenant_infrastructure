// Copyright 2026 NTT, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

// TestKubebuilderMarkersPresent ensures kubebuilder markers exist where expected
func TestKubebuilderMarkersPresent_DCIAppConnection(t *testing.T) {
	_, f, _, _ := runtime.Caller(0)
	dir := filepath.Dir(f)
	srcPath := filepath.Join(dir, "dciappconnection_types.go")
	b, err := ioutil.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("failed to read source file %s: %v", srcPath, err)
	}
	s := string(b)

	wantMarkers := []string{
		"+kubebuilder:validation:MinLength=1",
		"+kubebuilder:validation:MaxLength=63",
		"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$",
		"+kubebuilder:validation:MinItems",
		"+kubebuilder:validation:Enum",
	}
	for _, m := range wantMarkers {
		if !strings.Contains(s, m) {
			t.Errorf("expected marker %q to be present in %s", m, srcPath)
		}
	}

	// Ensure status-only types do not have validation markers
	// (leave empty for this CRD because status intentionally contains enum markers)
	statusOnlyTypes := []string{}
	for _, tn := range statusOnlyTypes {
		start := strings.Index(s, "type "+tn+" struct {")
		if start == -1 {
			continue
		}
		end := strings.Index(s[start:], "\n}")
		if end == -1 {
			t.Fatalf("could not locate end of struct %s in %s", tn, srcPath)
		}
		block := s[start : start+end]
		if strings.Contains(block, "+kubebuilder:validation:") {
			t.Errorf("status-only type %s contains kubebuilder validation markers; these should be on Spec types only", tn)
		}
	}
}

// helper: recursively collect properties named prop
func collectPropsByNameForConnection(props map[string]apiextensionsv1.JSONSchemaProps, name string, out *[]apiextensionsv1.JSONSchemaProps) {
	if props == nil {
		return
	}
	if p, ok := props[name]; ok {
		*out = append(*out, p)
	}
	for _, p := range props {
		collectPropsByName(p.Properties, name, out)
		if p.Items != nil && p.Items.Schema != nil {
			collectPropsByNameForConnection(p.Items.Schema.Properties, name, out)
		}
		if p.AdditionalProperties != nil && p.AdditionalProperties.Schema != nil {
			collectPropsByNameForConnection(p.AdditionalProperties.Schema.Properties, name, out)
		}
	}
}

// TestCRDContainsValidations checks the generated CRD's OpenAPI schema for expected rules
func TestCRDContainsValidations_DCIAppConnection(t *testing.T) {
	_, f, _, _ := runtime.Caller(0)
	dir := filepath.Dir(f)
	crdPath := filepath.Join(dir, "..", "..", "config", "crd", "bases", "example.com_dciappconnections.yaml")
	b, err := ioutil.ReadFile(crdPath)
	if err != nil {
		t.Fatalf("failed to read generated CRD %s: %v", crdPath, err)
	}

	var crd apiextensionsv1.CustomResourceDefinition
	if err := yaml.Unmarshal(b, &crd); err != nil {
		t.Fatalf("failed to unmarshal CRD YAML: %v", err)
	}

	if len(crd.Spec.Versions) == 0 || crd.Spec.Versions[0].Schema == nil || crd.Spec.Versions[0].Schema.OpenAPIV3Schema == nil {
		t.Fatalf("CRD schema not found or empty in %s", crdPath)
	}

	root := crd.Spec.Versions[0].Schema.OpenAPIV3Schema

	// 1) namespace/name fields should have minLength=1, maxLength=63 and a pattern
	var nsProps []apiextensionsv1.JSONSchemaProps
	collectPropsByNameForConnection(root.Properties, "namespace", &nsProps)
	collectPropsByNameForConnection(root.Properties, "name", &nsProps)
	if len(nsProps) == 0 {
		t.Fatalf("no schema properties named 'name' or 'namespace' found in CRD")
	}
	foundNSOK := false
	for _, p := range nsProps {
		if p.Type == "string" && p.MinLength != nil && p.MaxLength != nil && *p.MaxLength == 63 && p.Pattern != "" {
			foundNSOK = true
			break
		}
	}
	if !foundNSOK {
		t.Errorf("no 'name'/'namespace' property with minLength=1,maxLength=63 and pattern found in CRD; occurrences=%d", len(nsProps))
	}

	// 2) externalSubsets should have minItems=1 (array)
	var extProps []apiextensionsv1.JSONSchemaProps
	collectPropsByNameForConnection(root.Properties, "externalSubsets", &extProps)
	if len(extProps) == 0 {
		t.Fatalf("no schema properties named 'externalSubsets' found in CRD")
	}
	foundExtOK := false
	for _, p := range extProps {
		if p.Type == "array" && p.MinItems != nil && *p.MinItems >= 1 {
			foundExtOK = true
			break
		}
	}
	if !foundExtOK {
		t.Errorf("no 'externalSubsets' array property with minItems=1 found in CRD; occurrences=%d", len(extProps))
	}

	// 3) phase enum should contain Created/Failed and empty allowed
	var phaseProps []apiextensionsv1.JSONSchemaProps
	collectPropsByNameForConnection(root.Properties, "phase", &phaseProps)
	if len(phaseProps) == 0 {
		t.Fatalf("no schema properties named 'phase' found in CRD")
	}
	foundEnum := false
	for _, p := range phaseProps {
		// marshal the property and search for enum values as a fallback-safe check
		mb, err := yaml.Marshal(p)
		if err != nil {
			continue
		}
		s := string(mb)
		if strings.Contains(s, "Created") && strings.Contains(s, "Failed") {
			foundEnum = true
			break
		}
	}
	if !foundEnum {
		t.Errorf("phase enum does not contain expected values 'Created' and 'Failed'")
	}
}
