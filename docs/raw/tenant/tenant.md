Tenant
======



id
--

- Type: `string` (required)

The ID that identifies the tenant, chosen in the configuration. If a tenant with this ID already exists
in the project, it is taken over and its name is set from the configuration. Changing this value will
require the resource to be deleted and recreated.



project_id
----------

- Type: `string` (required)

The ID of the Descope project this tenant belongs to. Changing this value will require the resource to
be deleted and recreated.



deletion_protection
-------------------

- Type: `bool`

Protects the tenant from being accidentally destroyed or replaced. Destroying or replacing a tenant
removes every user's membership and roles in it, so tenants are protected by default. To allow the
resource to be destroyed or replaced, for example in ephemeral test environments, set this attribute to
`false` and apply the change first.



name
----

- Type: `string` (required)

The display name of the tenant. It must be unique in the project.
