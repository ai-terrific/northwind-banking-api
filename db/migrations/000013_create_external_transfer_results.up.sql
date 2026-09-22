ALTER TABLE regulator_events ALTER COLUMN transfer_id DROP NOT NULL;

CREATE TABLE IF NOT EXISTS external_transfer_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    external_transfer_id VARCHAR(255),
    reference_number VARCHAR(255),
    status VARCHAR(50) NOT NULL,
    error TEXT,
    payload JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_external_transfer_results_external_id ON external_transfer_results (external_transfer_id);

CREATE TRIGGER update_external_transfer_results_updated_at
    BEFORE UPDATE ON external_transfer_results
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();