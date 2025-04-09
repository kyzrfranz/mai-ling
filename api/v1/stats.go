package v1

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
