package v1

import "time"

type Stats struct {
	UniqueRequests int `bson:"uniqueRequests" json:"uniqueRequests"`
	TotalIds       int `bson:"totalIds" json:"totalLetters"`
	SentIds        int `bson:"sentIds" json:"sentIds"`
	TopIds         []struct {
		ID    string `bson:"_id" json:"id"`
		Count int    `bson:"count" json:"count"`
	} `bson:"topIds" json:"topIds"`
	StatusCounts struct {
		Queued int `bson:"queued" json:"queued"`
		Sent   int `bson:"sent" json:"sent"`
	} `json:"statusCounts" bson:"statusCounts"`
}

type StatsByLocation struct {
	City    string  `bson:"city,omitempty" json:"city,omitempty"`
	ZipCode string  `bson:"zipCode,omitempty" json:"zipCode,omitempty"`
	Lat     float64 `bson:"lat,omitempty" json:"lat,omitempty"`
	Lng     float64 `bson:"lng,omitempty" json:"lng,omitempty"`
	Count   int     `bson:"count" json:"count"`
}

type StatsByCreationDate struct {
	Count int       `bson:"count" json:"count"`
	Date  time.Time `bson:"date" json:"date"`
}
