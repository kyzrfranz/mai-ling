package v1

import "time"

var (
	PrintJobStatusQueued = "queued"
	PrintJobStatusSent   = "sent"
)

type PrintJobStatus string

type PrintJob struct {
	Id           string         `json:"id,omitempty" bson:"_id"`
	RecipientIds []string       `json:"recipientIds,omitempty" bson:"recipientIds"`
	MyMdbs       []string       `json:"myMdbs,omitempty" bson:"myMdbs"`
	Status       PrintJobStatus `json:"status,omitempty" bson:"status"`
	Address      struct {
		Name    string `json:"name" bson:"name"`
		Company string `json:"company,omitempty" bson:"company"`
		Street  string `json:"street" bson:"street"`
		Number  int    `json:"number" bson:"number"`
		Zip     string `json:"zip" bson:"zip"`
		City    string `json:"city" bson:"city"`
	} `json:"address" bson:"address"`
	CreationDate time.Time `json:"creationDate,omitempty" bson:"creationDate"`
}
