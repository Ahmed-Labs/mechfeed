CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    webhook_url VARCHAR(2083),
    created timestamp DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_alerts (
    alert_id SERIAL PRIMARY KEY,
    id VARCHAR(36) NOT NULL,
    keyword VARCHAR(255) NOT NULL,
    FOREIGN KEY (id) REFERENCES users(id) ON DELETE CASCADE
);
