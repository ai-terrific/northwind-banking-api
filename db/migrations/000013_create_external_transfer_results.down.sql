DROP TABLE IF EXISTS external_transfer_results;

DELETE FROM regulator_events WHERE transfer_id IS NULL;

ALTER TABLE regulator_events ALTER COLUMN transfer_id SET NOT NULL;