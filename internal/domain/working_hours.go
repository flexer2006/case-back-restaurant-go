// Package domain contains the main business entities of the application and their behavior
package domain

import (
	"time"
)

type WeekDay int

const (
	Monday WeekDay = iota + 1

	Tuesday

	Wednesday

	Thursday

	Friday

	Saturday

	Sunday
)

type TimeSlot struct {
	Start time.Time
	End   time.Time
}

type WorkingHours struct {
	ID           string    `json:"id"`
	RestaurantID string    `json:"restaurant_id"`
	WeekDay      WeekDay   `json:"week_day"`
	OpenTime     string    `json:"open_time"`
	CloseTime    string    `json:"close_time"`
	IsClosed     bool      `json:"is_closed"`
	ValidFrom    time.Time `json:"valid_from"`
	ValidTo      time.Time `json:"valid_to"`
}
