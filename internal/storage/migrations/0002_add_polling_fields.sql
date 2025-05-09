-- Add polling fields to issues table
ALTER TABLE issues ADD COLUMN polling_interval INTEGER NOT NULL DEFAULT 0;
ALTER TABLE issues ADD COLUMN last_polled_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;