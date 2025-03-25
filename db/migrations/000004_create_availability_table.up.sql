CREATE TABLE IF NOT EXISTS availability (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    restaurant_id UUID NOT NULL,
    date DATE NOT NULL,
    time_slot VARCHAR(5) NOT NULL, -- Формат: "HH:MM"
    capacity INT NOT NULL,
    reserved INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_restaurant FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
    CONSTRAINT unique_slot UNIQUE (restaurant_id, date, time_slot)
);

CREATE INDEX idx_availability_restaurant_date ON availability(restaurant_id, date);
CREATE INDEX idx_availability_date ON availability(date); 