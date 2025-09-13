-- Create tenants table (condominiums/properties)
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    legal_name VARCHAR(255),
    tax_id VARCHAR(50),
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100) DEFAULT 'PE',
    phone VARCHAR(20),
    email VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deleted')),
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create units table (apartments/properties within a tenant)
CREATE TABLE IF NOT EXISTS units (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    unit_number VARCHAR(50) NOT NULL,
    unit_type VARCHAR(50) DEFAULT 'apartment' CHECK (unit_type IN ('apartment', 'house', 'commercial', 'parking', 'storage')),
    floor_number INTEGER,
    building VARCHAR(50),
    area_sqm DECIMAL(10,2),
    bedrooms INTEGER,
    bathrooms INTEGER,
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'maintenance', 'sold')),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, unit_number)
);

-- Create user_unit_roles table (relationship between users and units with roles)
CREATE TABLE IF NOT EXISTS user_unit_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    unit_id UUID NOT NULL REFERENCES units(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL CHECK (role IN ('owner', 'tenant', 'family_member', 'administrator', 'manager')),
    is_president BOOLEAN DEFAULT FALSE,
    is_primary BOOLEAN DEFAULT FALSE,
    percentage_ownership DECIMAL(5,2) DEFAULT 100.00,
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'pending')),
    valid_from TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    valid_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, tenant_id, unit_id, role)
);

-- Create tenant_presidents table (track current president per tenant)
CREATE TABLE IF NOT EXISTS tenant_presidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    unit_id UUID NOT NULL REFERENCES units(id) ON DELETE CASCADE,
    appointed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    appointed_by UUID REFERENCES users(id),
    term_start TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    term_end TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'expired')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, status) DEFERRABLE INITIALLY DEFERRED
);

-- Create indexes
CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_country ON tenants(country);

CREATE INDEX idx_units_tenant_id ON units(tenant_id);
CREATE INDEX idx_units_status ON units(status);
CREATE INDEX idx_units_unit_number ON units(unit_number);

CREATE INDEX idx_user_unit_roles_user_id ON user_unit_roles(user_id);
CREATE INDEX idx_user_unit_roles_tenant_id ON user_unit_roles(tenant_id);
CREATE INDEX idx_user_unit_roles_unit_id ON user_unit_roles(unit_id);
CREATE INDEX idx_user_unit_roles_role ON user_unit_roles(role);
CREATE INDEX idx_user_unit_roles_is_president ON user_unit_roles(is_president);
CREATE INDEX idx_user_unit_roles_status ON user_unit_roles(status);

CREATE INDEX idx_tenant_presidents_tenant_id ON tenant_presidents(tenant_id);
CREATE INDEX idx_tenant_presidents_user_id ON tenant_presidents(user_id);
CREATE INDEX idx_tenant_presidents_status ON tenant_presidents(status);

-- Create updated_at triggers
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_units_updated_at BEFORE UPDATE ON units
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_unit_roles_updated_at BEFORE UPDATE ON user_unit_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tenant_presidents_updated_at BEFORE UPDATE ON tenant_presidents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();