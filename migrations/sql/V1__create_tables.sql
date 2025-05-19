CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    is_activated BOOLEAN DEFAULT FALSE,
    token VARCHAR(255) UNIQUE NOT NULL,
    frequency VARCHAR(50),
    city VARCHAR(255)
);

CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_token ON users (token);


CREATE TABLE cities (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE weather_descriptions (
    id BIGSERIAL PRIMARY KEY,
    description VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE weather (
    id BIGSERIAL PRIMARY KEY,
    city_id BIGINT NOT NULL,
    description_id BIGINT NOT NULL,
    temperature INTEGER,
    humidity INTEGER,
    observation_time TIMESTAMP WITH TIME ZONE
);

ALTER TABLE weather
ADD CONSTRAINT fk_weather_city
FOREIGN KEY (city_id)
REFERENCES cities (id);

ALTER TABLE weather
ADD CONSTRAINT fk_weather_description
FOREIGN KEY (description_id)
REFERENCES weather_descriptions (id);

CREATE INDEX idx_weather_city_id ON weather (city_id);
CREATE INDEX idx_weather_description_id ON weather (description_id);


CREATE TABLE application_settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT
);

INSERT INTO application_settings (key, value) VALUES
('Port', '8080'),
('EmailSubscribeText', 'Thank you for subscribing to our weather report service! \n\nTo confirm subscription please visit this page {1}\nIf you want to unsubscribe visit this page {2}'),
('EmailUpdateText', 'Thank you... {1} ... {2}'),
('LocalURL', 'http://localhost:8080'),
('SMTP_Adres', 'smtp.gmail.com'),
('SMTP_Port', '587'),
('SMTP_Username', 'pikkljuihilhu@gmail.com'),
('SMTP_Password', 'ueyg cmai exsb rsxm');