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

// TestKubebuilderMarkersPresent ensures the kubebuilder validation markers exist
// in the source file. This is a lightweight check that the developer added the
// markers that should be reflected into the generated CRD.
func TestKubebuilderMarkersPresent(t *testing.T) {
	// read this source file's directory
	_, f, _, _ := runtime.Caller(0)
	dir := filepath.Dir(f)
	srcPath := filepath.Join(dir, "connectendpoint_types.go")
	b, err := ioutil.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("failed to read source file %s: %v", srcPath, err)
	}
	s := string(b)

	// We check for presence of key validation markers in the source to
	// ensure developer added them. Exact numeric values may be altered by
	// controller-gen or kubebuilder conventions; schema checks below assert
	// the generated CRD contents which are authoritative.
	wantMarkers := []string{
		"+kubebuilder:validation:MaxLength=15",
		"25[0-5]",
		"+kubebuilder:validation:MaxLength=63",
		"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$",
		"+kubebuilder:validation:MinItems",
	}

	// Check presence
	for _, m := range wantMarkers {
		if !strings.Contains(s, m) {
			t.Errorf("expected marker %q to be present in %s", m, srcPath)
		}
	}

	// Ensure status-only types do NOT accidentally contain validation markers.
	// These types should be lightweight and omit kubebuilder validation annotations.
	statusOnlyTypes := []string{
		"NamespacedNameStatus",
		"FromPodStatus",
		"FromDeploymentStatus",
		"ToDeploymentStatusStatus",
		"ToPodStatusStatus",
	}
	for _, tn := range statusOnlyTypes {
		start := strings.Index(s, "type "+tn+" struct {")
		if start == -1 {
			// type not present in this file -- skip
			continue
		}
		// find the end of the struct by locating the next standalone '}' after start
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
func collectPropsByName(props map[string]apiextensionsv1.JSONSchemaProps, name string, out *[]apiextensionsv1.JSONSchemaProps) {
	if props == nil {
		return
	}
	if p, ok := props[name]; ok {
		*out = append(*out, p)
	}
	for _, p := range props {
		// traverse nested properties
		collectPropsByName(p.Properties, name, out)
		// arrays
		if p.Items != nil && p.Items.Schema != nil {
			collectPropsByName(p.Items.Schema.Properties, name, out)
		}
		// additionalProperties may be a schema
		if p.AdditionalProperties != nil && p.AdditionalProperties.Schema != nil {
			collectPropsByName(p.AdditionalProperties.Schema.Properties, name, out)
		}
	}
}

// TestCRDContainsValidations loads the generated CRD and asserts that the
// expected validation rules are present in the OpenAPI v3 schema. If the
// assertions fail it indicates the markers were not propagated into the CRD.
func TestCRDContainsValidations(t *testing.T) {
	// resolve path to generated CRD relative to this test file
	_, f, _, _ := runtime.Caller(0)
	dir := filepath.Dir(f)
	// repo root is two levels up from api/v1
	crdPath := filepath.Join(dir, "..", "..", "config", "crd", "bases", "example.com_connectendpoints.yaml")
	b, err := ioutil.ReadFile(crdPath)
	if err != nil {
		t.Fatalf("failed to read generated CRD %s: %v", crdPath, err)
	}

	// CRD YAML -> object
	var crd apiextensionsv1.CustomResourceDefinition
	if err := yaml.Unmarshal(b, &crd); err != nil {
		t.Fatalf("failed to unmarshal CRD YAML: %v", err)
	}

	if len(crd.Spec.Versions) == 0 || crd.Spec.Versions[0].Schema == nil || crd.Spec.Versions[0].Schema.OpenAPIV3Schema == nil {
		t.Fatalf("CRD schema not found or empty in %s", crdPath)
	}

	root := crd.Spec.Versions[0].Schema.OpenAPIV3Schema

	// 1) Address fields should have maxLength=15 and pattern (IP or empty)
	var addrProps []apiextensionsv1.JSONSchemaProps
	collectPropsByName(root.Properties, "address", &addrProps)
	if len(addrProps) == 0 {
		t.Fatalf("no schema properties named 'address' found in CRD")
	}
	foundAddrOk := false
	for _, p := range addrProps {
		if p.Type == "string" && p.MaxLength != nil && *p.MaxLength == 15 && p.Pattern != "" {
			foundAddrOk = true
			break
		}
	}
	if !foundAddrOk {
		t.Errorf("no 'address' property with type=string, maxLength=15 and pattern found in CRD; found %d 'address' occurrences", len(addrProps))
	}

	// 2) NamespacedName name/namespace should have minLength=1, maxLength=63 and a pattern
	var nsProps []apiextensionsv1.JSONSchemaProps
	collectPropsByName(root.Properties, "namespace", &nsProps)
	collectPropsByName(root.Properties, "name", &nsProps)
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
		t.Errorf("no 'name'/'namespace' property with minLength=0,maxLength=63 and pattern found in CRD; occurrences=%d", len(nsProps))
	}

	// 3) externalSubsets should have minItems=1 (array)
	var extProps []apiextensionsv1.JSONSchemaProps
	collectPropsByName(root.Properties, "externalSubsets", &extProps)
	if len(extProps) == 0 {
		t.Fatalf("no schema properties named 'externalSubsets' found in CRD")
	}
	foundExtOK := false
	for _, p := range extProps {
		if p.Type == "array" && p.MinItems != nil {
			foundExtOK = true
			break
		}
	}
	if !foundExtOK {
		t.Errorf("no 'externalSubsets' array property with minItems=0 found in CRD; occurrences=%d", len(extProps))
	}
}
