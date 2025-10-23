package types

import "time"

type OSMWay struct {
	ID        string
	Name      string
	Priority  float64
	UpdatedAt time.Time
}
