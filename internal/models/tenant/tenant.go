package tenant

import (
	"context"

	"github.com/descope/terraform-provider-descope/internal/attrs/boolattr"
	"github.com/descope/terraform-provider-descope/internal/attrs/stringattr"
	"github.com/descope/terraform-provider-descope/internal/helpers"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var Schema = schema.Schema{
	MarkdownDescription: "Manages a single tenant in a Descope project, identified by a tenant ID chosen in the configuration.",
	Attributes:          TenantAttributes,
}

var TenantAttributes = map[string]schema.Attribute{
	"id":                  stringattr.Required(stringattr.MachineIDValidator, stringplanmodifier.RequiresReplace()),
	"project_id":          stringattr.Required(stringplanmodifier.RequiresReplace()),
	"deletion_protection": boolattr.Tristate(),
	"name":                stringattr.Required(stringattr.StandardLenValidator),
}

type TenantModel struct {
	ID                 stringattr.Type `tfsdk:"id"`
	ProjectID          stringattr.Type `tfsdk:"project_id"`
	DeletionProtection boolattr.Type   `tfsdk:"deletion_protection"`
	Name               stringattr.Type `tfsdk:"name"`
}

func (m *TenantModel) Values(h *helpers.Handler) map[string]any {
	data := map[string]any{}
	stringattr.Get(m.ID, data, "id")
	stringattr.Get(m.Name, data, "name")
	return data
}

func (m *TenantModel) SetValues(h *helpers.Handler, data map[string]any) {
	stringattr.Set(&m.ID, data, "id")
	stringattr.Set(&m.Name, data, "name")
}

func (m *TenantModel) DeletionProtectionDefault(_ context.Context) bool {
	return true
}

func (m *TenantModel) GetID() stringattr.Type        { return m.ID }
func (m *TenantModel) SetID(id stringattr.Type)      { m.ID = id }
func (m *TenantModel) GetProjectID() stringattr.Type { return m.ProjectID }
