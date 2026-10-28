package v1

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

// TestKubebuilderMarkersPresent ensures the kubebuilder validation markers exist
// in the source file for DCIAppState.
func TestKubebuilderMarkersPresent_DCIAppState(t *testing.T) {
	_, f, _, _ := runtime.Caller(0)
	dir := filepath.Dir(f)
	srcPath := filepath.Join(dir, "dciappstate_types.go")
	b, err := ioutil.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("failed to read source file %s: %v", srcPath, err)
	}
	s := string(b)

	wantMarkers := []string{
		"+kubebuilder:validation:Enum=\"computation\";\"connection\"",
	}

	for _, m := range wantMarkers {
		if !strings.Contains(s, m) {
			t.Errorf("expected marker %q to be present in %s", m, srcPath)
		}
	}
}

// TestCRDContainsValidations loads the generated CRD for DCIAppState and
// asserts that expected validation rules (enum values) are present.
func TestCRDContainsValidations_DCIAppState(t *testing.T) {
	_, f, _, _ := runtime.Caller(0)
	dir := filepath.Dir(f)
	crdPath := filepath.Join(dir, "..", "..", "config", "crd", "bases", "example.com_dciappstates.yaml")
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

	// Look up spec.type enum
	spec, ok := root.Properties["spec"]
	if !ok {
		t.Fatalf("spec property not found in CRD schema")
	}
	tProp, ok := spec.Properties["type"]
	if !ok {
		t.Fatalf("spec.type property not found in CRD schema")
	}
	if len(tProp.Enum) == 0 {
		t.Fatalf("spec.type enum not present in CRD schema")
	}

	// collect enum strings
	found := map[string]bool{}
	for _, je := range tProp.Enum {
		// apiextensionsv1.JSON contains Raw bytes; unmarshal to string
		var val string
		if err := json.Unmarshal(je.Raw, &val); err == nil {
			found[val] = true
		}
	}

	if !found["computation"] || !found["connection"] {
		t.Errorf("spec.type enum does not contain expected values; got=%v", found)
	}
}
