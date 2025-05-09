-- Add polling_interval field to issues table
ALTER TABLE issues ADD COLUMN polling_interval INTEGER NOT NULL DEFAULT 300; -- 300 seconds = 5 minutes