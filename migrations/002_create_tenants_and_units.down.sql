-- Drop triggers
DROP TRIGGER IF EXISTS update_tenant_presidents_updated_at ON tenant_presidents;
DROP TRIGGER IF EXISTS update_user_unit_roles_updated_at ON user_unit_roles;
DROP TRIGGER IF EXISTS update_units_updated_at ON units;
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;

-- Drop tables in reverse order
DROP TABLE IF EXISTS tenant_presidents;
DROP TABLE IF EXISTS user_unit_roles;
DROP TABLE IF EXISTS units;
DROP TABLE IF EXISTS tenants;