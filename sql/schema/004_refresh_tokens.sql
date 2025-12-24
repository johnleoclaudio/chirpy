-- +goose Up
CREATE TABLE refresh_tokens (
  token TEXT PRIMARY KEY,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  user_id UUID REFERENCES users(id) ON DELETE CASCADE NOT NUll ,
  expires_at TIMESTAMP NOT NULL,
  revoked_at TIMESTAMP
);

-- +goose Down
DROP TABLE refresh_tokens;
