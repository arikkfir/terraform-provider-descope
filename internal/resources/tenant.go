package resources

import (
	"context"
	"maps"

	"github.com/descope/go-sdk/descope"
	"github.com/descope/terraform-provider-descope/internal/infra"
	"github.com/descope/terraform-provider-descope/internal/models/tenant"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const tenantAlreadyExists = "E073307"

func NewTenantResource() resource.Resource {
	read := func(ctx context.Context, c *infra.Client, projectID, id string) (map[string]any, error) {
		return c.Get(ctx, projectID, "/v1/mgmt/tenant", map[string]string{"id": id})
	}
	update := func(ctx context.Context, c *infra.Client, projectID, id string, data map[string]any) (map[string]any, error) {
		body, err := read(ctx, c, projectID, id)
		if err != nil {
			return nil, err
		}
		maps.Copy(body, data)
		body["id"] = id
		if err := c.Post(ctx, projectID, "/v1/mgmt/tenant/update", body); err != nil {
			return nil, err
		}
		return read(ctx, c, projectID, id)
	}
	return newResource[tenant.TenantModel]("tenant", tenant.Schema, operations{
		CreateOverwrites: true,
		Create: func(ctx context.Context, c *infra.Client, projectID string, data map[string]any) (string, map[string]any, error) {
			id, _ := data["id"].(string)
			err := c.Post(ctx, projectID, "/v1/mgmt/tenant/create", data)
			if descope.IsError(err, tenantAlreadyExists) {
				if entity, updateErr := update(ctx, c, projectID, id, data); !infra.AsNotFoundError(updateErr) {
					return id, entity, updateErr
				}
			}
			if err != nil {
				return "", nil, err
			}
			entity, err := read(ctx, c, projectID, id)
			return id, entity, err
		},
		Read:   read,
		Update: update,
		Delete: func(ctx context.Context, c *infra.Client, projectID, id string) error {
			return c.Post(ctx, projectID, "/v1/mgmt/tenant/delete", map[string]any{"id": id})
		},
	})
}
