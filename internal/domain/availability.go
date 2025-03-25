package domain

import (
	"time"
)

const HighOccupancyThreshold = 0.8

type Availability struct {
	ID           string    `json:"id"`
	RestaurantID string    `json:"restaurant_id"`
	Date         time.Time `json:"date"`
	TimeSlot     string    `json:"time_slot"`
	Capacity     int       `json:"capacity"`
	Reserved     int       `json:"reserved"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (a *Availability) AvailabilityStatus() string {
	if a.Capacity <= a.Reserved {
		return "fully_booked"
	}
	if float64(a.Reserved)/float64(a.Capacity) >= HighOccupancyThreshold {
		return "limited"
	}

	return "available"
}

func (a *Availability) AvailableSeats() int {
	return a.Capacity - a.Reserved
}
