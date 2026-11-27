package tenant_test

import (
	"regexp"
	"testing"

	"github.com/descope/terraform-provider-descope/tools/testacc"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

func TestTenant(t *testing.T) {
	projectID := testacc.ProjectID(t)
	r := testacc.Tenant(t)
	id := testacc.GenerateAlias(t)
	otherID := id + "-b"
	testacc.RunWithDestroyCheck(t, "descope_tenant",
		resource.TestStep{
			Config: r.Block(`
				project_id = "` + projectID + `"
				deletion_protection = false
				id = "` + id + `"
				name = "` + id + `"
			`),
			Check: r.Check(map[string]any{
				"id":         id,
				"project_id": projectID,
				"name":       id,
			}),
		},
		// renaming keeps the settings the resource does not manage
		resource.TestStep{
			PreConfig: func() {
				testacc.OutOfBandPost(t, projectID, "/v1/mgmt/tenant/update", map[string]any{"id": id, "name": id, "selfProvisioningDomains": []string{"testacc.example.com"}})
			},
			Config: r.Block(`
				project_id = "` + projectID + `"
				deletion_protection = false
				id = "` + id + `"
				name = "` + id + ` renamed"
			`),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(r.Path(), plancheck.ResourceActionUpdate)},
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				r.Check(map[string]any{"name": id + " renamed"}),
				func(*terraform.State) error {
					tenant := searchTenant(t, projectID, id)
					require.Equal(t, []any{"testacc.example.com"}, tenant["selfProvisioningDomains"])
					return nil
				},
			),
		},
		// changing the id replaces the tenant
		resource.TestStep{
			Config: r.Block(`
				project_id = "` + projectID + `"
				deletion_protection = false
				id = "` + otherID + `"
				name = "` + otherID + `"
			`),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(r.Path(), plancheck.ResourceActionDestroyBeforeCreate)},
			},
			Check: r.Check(map[string]any{"id": otherID}),
		},
		resource.TestStep{
			ResourceName:            r.Path(),
			ImportState:             true,
			ImportStateIdFunc:       testacc.GenerateImportStateID(r.Path(), "project_id", "id"),
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: []string{"deletion_protection"},
		},
	)
}

func TestTenantAdoptsExisting(t *testing.T) {
	projectID := testacc.ProjectID(t)
	r := testacc.Tenant(t)
	id := testacc.GenerateAlias(t)
	testacc.RunWithDestroyCheck(t, "descope_tenant",
		// creating a tenant whose id already exists takes it over and keeps its other settings
		resource.TestStep{
			PreConfig: func() {
				testacc.OutOfBandPost(t, projectID, "/v1/mgmt/tenant/create", map[string]any{"id": id, "name": id + " old", "selfProvisioningDomains": []string{"testacc.example.com"}})
			},
			Config: r.Block(`
				project_id = "` + projectID + `"
				deletion_protection = false
				id = "` + id + `"
				name = "` + id + `"
			`),
			Check: resource.ComposeAggregateTestCheckFunc(
				r.Check(map[string]any{"id": id, "name": id}),
				func(*terraform.State) error {
					tenant := searchTenant(t, projectID, id)
					require.Equal(t, id, tenant["name"])
					require.Equal(t, []any{"testacc.example.com"}, tenant["selfProvisioningDomains"])
					return nil
				},
			),
		},
	)
}

func TestTenantNameTakenByAnotherTenant(t *testing.T) {
	projectID := testacc.ProjectID(t)
	r := testacc.Tenant(t)
	id := testacc.GenerateAlias(t)
	testacc.Run(t,
		// a name that belongs to a tenant with another id is an error rather than a takeover
		resource.TestStep{
			PreConfig: func() {
				testacc.OutOfBandPost(t, projectID, "/v1/mgmt/tenant/create", map[string]any{"id": id + "-other", "name": id})
				t.Cleanup(func() {
					testacc.OutOfBandPost(t, projectID, "/v1/mgmt/tenant/delete", map[string]any{"id": id + "-other"})
				})
			},
			Config: r.Block(`
				project_id = "` + projectID + `"
				deletion_protection = false
				id = "` + id + `"
				name = "` + id + `"
			`),
			ExpectError: regexp.MustCompile(`tenant name is duplicate`),
		},
	)
}

func searchTenant(t *testing.T, projectID, id string) map[string]any {
	data := testacc.OutOfBandPostData(t, projectID, "/v1/mgmt/tenant/search", map[string]any{"tenantIds": []string{id}})
	tenants, _ := data["tenants"].([]any)
	require.Len(t, tenants, 1)
	tenant, _ := tenants[0].(map[string]any)
	return tenant
}
