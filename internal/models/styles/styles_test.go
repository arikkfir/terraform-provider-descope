package styles_test

import (
	"encoding/json"
	"testing"

	"github.com/descope/terraform-provider-descope/tools/testacc"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestStyles(t *testing.T) {
	projectID := testacc.ProjectID(t)
	s := testacc.Styles(t)
	exported := ""
	testacc.Run(t,
		// import the styles from the theme representation
		resource.TestStep{
			Config: s.Block(`
				project_id = "` + projectID + `"
				data = jsonencode({
					styles = {
						light = {}
						dark = {}
					}
				})
			`),
			Check: s.Check(map[string]any{
				"id.==":      "project_id",
				"project_id": testacc.AttributeIsSet,
				"data":       testacc.AttributeIsSet,
			}),
		},
		// changing the styles data updates the theme in place
		resource.TestStep{
			Config: s.Block(`
				project_id = "` + projectID + `"
				data = jsonencode({
					styles = {
						light = {
							designTokens = {}
						}
						dark = {}
					}
				})
			`),
			Check: s.Check(map[string]any{
				"id.==": "project_id",
				"data":  testacc.AttributeIsSet,
			}),
		},
		// importing the styles plans no change when the configuration leaves out the components version the export names
		resource.TestStep{
			PreConfig: func() {
				exported = exportedWithoutComponentsVersion(t, projectID)
			},
			Config: "variable \"styles\" {}\n" + s.Block(`
				project_id = "`+projectID+`"
				data = var.styles
			`),
			ConfigVariables:   config.Variables{"styles": lazyVariable(func() string { return exported })},
			ResourceName:      s.Path(),
			ImportState:       true,
			ImportStateKind:   resource.ImportBlockWithID,
			ImportStateIdFunc: testacc.GenerateImportStateID(s.Path(), "project_id"),
		},
		// data is re-serialized server-side on read, so verify the identity attributes only
		resource.TestStep{
			ResourceName:            s.Path(),
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateIdFunc:       testacc.GenerateImportStateID(s.Path(), "project_id"),
			ImportStateVerifyIgnore: []string{"data"},
		},
		// removing the resource is a no-op that leaves the theme in place
	)
}

func exportedWithoutComponentsVersion(t *testing.T, projectID string) string {
	body := testacc.OutOfBandPostData(t, projectID, "/v2/mgmt/theme/export", map[string]any{})
	theme, ok := body["theme"].(map[string]any)
	require.True(t, ok, "The theme export has no theme object")
	require.Contains(t, theme, "componentsVersion")
	delete(theme, "componentsVersion")
	data, err := json.Marshal(theme)
	require.NoError(t, err)
	return string(data)
}

type lazyVariable func() string

func (v lazyVariable) MarshalJSON() ([]byte, error) {
	return json.Marshal(v())
}
