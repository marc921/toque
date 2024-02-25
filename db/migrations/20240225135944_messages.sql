-- migrate:up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE messages (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v1(),
  message TEXT,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- migrate:down
DROP TABLE messages;
