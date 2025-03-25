CREATE TABLE IF NOT EXISTS working_hours (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    restaurant_id UUID NOT NULL,
    week_day INT NOT NULL CHECK (week_day BETWEEN 1 AND 7),
    open_time VARCHAR(5) NOT NULL, -- Формат: "HH:MM"
    close_time VARCHAR(5) NOT NULL, -- Формат: "HH:MM"
    is_closed BOOLEAN NOT NULL DEFAULT FALSE,
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_to TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_restaurant FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
    CONSTRAINT unique_restaurant_day_validity UNIQUE (restaurant_id, week_day, valid_from)
);

CREATE INDEX idx_working_hours_restaurant_id ON working_hours(restaurant_id);
CREATE INDEX idx_working_hours_validity ON working_hours(valid_from, valid_to); 