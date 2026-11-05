CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE branches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    address TEXT,
    timezone VARCHAR(100) NOT NULL DEFAULT 'Asia/Jakarta',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    phone VARCHAR(30) NOT NULL,
    email VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, phone)
);

CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    description TEXT,
    duration_minutes INTEGER NOT NULL CHECK (duration_minutes > 0),
    price_amount BIGINT NOT NULL DEFAULT 0 CHECK (price_amount >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, name)
);

CREATE TABLE staff_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'STAFF' CHECK (role IN ('OWNER', 'ADMIN', 'RECEPTIONIST', 'STAFF')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE staff_services (
    staff_member_id UUID NOT NULL REFERENCES staff_members(id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    PRIMARY KEY (staff_member_id, service_id)
);

CREATE TYPE appointment_status AS ENUM (
    'BOOKED', 'CHECKED_IN', 'IN_QUEUE', 'CALLED', 'IN_SERVICE', 'COMPLETED', 'CANCELLED', 'NO_SHOW'
);

CREATE TABLE appointments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_code VARCHAR(20) NOT NULL UNIQUE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE RESTRICT,
    staff_member_id UUID NOT NULL REFERENCES staff_members(id) ON DELETE RESTRICT,
    scheduled_start_at TIMESTAMPTZ NOT NULL,
    scheduled_end_at TIMESTAMPTZ NOT NULL,
    status appointment_status NOT NULL DEFAULT 'BOOKED',
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (scheduled_end_at > scheduled_start_at)
);

CREATE TYPE queue_status AS ENUM ('WAITING', 'CALLED', 'IN_SERVICE', 'COMPLETED', 'SKIPPED');

CREATE TABLE queue_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    appointment_id UUID NOT NULL UNIQUE REFERENCES appointments(id) ON DELETE CASCADE,
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    queue_date DATE NOT NULL,
    queue_number INTEGER NOT NULL CHECK (queue_number > 0),
    status queue_status NOT NULL DEFAULT 'WAITING',
    checked_in_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    called_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (branch_id, queue_date, queue_number)
);

CREATE INDEX idx_appointments_branch_schedule ON appointments (branch_id, scheduled_start_at);
CREATE INDEX idx_appointments_staff_schedule ON appointments (staff_member_id, scheduled_start_at);
CREATE INDEX idx_queue_entries_branch_date ON queue_entries (branch_id, queue_date, queue_number);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_appointments_updated_at
BEFORE UPDATE ON appointments
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

WITH organization AS (
    INSERT INTO organizations (name, slug)
    VALUES ('Klinik Sehat Sentosa', 'klinik-sehat-sentosa')
    RETURNING id
), branch AS (
    INSERT INTO branches (organization_id, name, slug, address)
    SELECT id, 'Jakarta Selatan', 'jakarta-selatan', 'Jl. Kemang Raya No. 10, Jakarta Selatan'
    FROM organization
    RETURNING id, organization_id
), service_data AS (
    INSERT INTO services (organization_id, name, description, duration_minutes, price_amount)
    SELECT organization_id, 'Konsultasi Umum', 'Konsultasi dengan dokter umum.', 30, 150000 FROM branch
    UNION ALL
    SELECT organization_id, 'Pemeriksaan Gigi', 'Pemeriksaan awal kesehatan gigi.', 45, 200000 FROM branch
    RETURNING id, name
), staff AS (
    INSERT INTO staff_members (branch_id, name, role)
    SELECT id, 'dr. Andi Pratama', 'STAFF' FROM branch
    UNION ALL
    SELECT id, 'drg. Sinta Larasati', 'STAFF' FROM branch
    RETURNING id, name
)
INSERT INTO staff_services (staff_member_id, service_id)
SELECT staff.id, service_data.id
FROM staff
CROSS JOIN service_data
WHERE (staff.name = 'dr. Andi Pratama' AND service_data.name = 'Konsultasi Umum')
   OR (staff.name = 'drg. Sinta Larasati' AND service_data.name = 'Pemeriksaan Gigi');
