package domain

import (
	"time"
)

type Cuisine string

type Restaurant struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Address      string    `json:"address"`
	Cuisine      Cuisine   `json:"cuisine"`
	Description  string    `json:"description"`
	Facts        []Fact    `json:"facts"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ContactEmail string    `json:"contact_email"`
	ContactPhone string    `json:"contact_phone"`
}

type Fact struct {
	ID           string    `json:"id"`
	RestaurantID string    `json:"restaurant_id"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
}
