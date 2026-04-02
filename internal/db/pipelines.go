package db

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func DefaultPipeline(topIds int) mongo.Pipeline {
	return mongo.Pipeline{
		{{"$facet", bson.D{
			// Facet for unique summary based on address & ids
			{"uniqueSummary", bson.A{
				bson.D{{"$addFields", bson.D{
					// Fix: Use $ifNull to ensure $size receives an array
					{"idsCount", bson.D{{"$size", bson.D{{"$ifNull", bson.A{"$ids", bson.A{}}}}}}},
				}}},
				bson.D{{"$group", bson.D{
					{"_id", bson.D{
						{"address", "$address"},
						{"ids", "$ids"},
					}},
					{"requestIdsCount", bson.D{{"$first", "$idsCount"}}},
				}}},
				bson.D{{"$group", bson.D{
					{"_id", nil},
					{"uniqueRequests", bson.D{{"$sum", 1}}},
					{"totalIds", bson.D{{"$sum", "$requestIdsCount"}}},
				}}},
				bson.D{{"$project", bson.D{
					{"_id", 0},
					{"uniqueRequests", 1},
					{"totalIds", 1},
				}}},
			}},
			{"topIds", bson.A{
				// Unwind handles null/missing ids by simply dropping the doc from this facet
				bson.D{{"$unwind", "$ids"}},
				bson.D{{"$group", bson.D{
					{"_id", "$ids"},
					{"count", bson.D{{"$sum", 1}}},
				}}},
				bson.D{{"$sort", bson.D{{"count", -1}}}},
				bson.D{{"$limit", topIds}},
			}},
			// Facet for status counts with deduplication.
			{"statusCounts", bson.A{
				bson.D{{"$addFields", bson.D{
					// Fix: Use $ifNull to ensure $size receives an array
					{"idsCount", bson.D{{"$size", bson.D{{"$ifNull", bson.A{"$ids", bson.A{}}}}}}},
				}}},
				bson.D{{"$group", bson.D{
					{"_id", bson.D{
						{"address", "$address"},
						{"ids", "$ids"},
					}},
					{"status", bson.D{{"$first", "$status"}}},
					{"idsCount", bson.D{{"$first", "$idsCount"}}},
				}}},
				bson.D{{"$group", bson.D{
					{"_id", nil},
					{"queued", bson.D{{"$sum", bson.D{{"$cond", bson.A{
						bson.D{{"$eq", bson.A{"$status", "queued"}}},
						1,
						0,
					}}}}}},
					{"sent", bson.D{{"$sum", bson.D{{"$cond", bson.A{
						bson.D{{"$eq", bson.A{"$status", "sent"}}},
						1,
						0,
					}}}}}},
					{"sentLetters", bson.D{{"$sum", bson.D{{"$cond", bson.A{
						bson.D{{"$eq", bson.A{"$status", "sent"}}},
						"$idsCount",
						0,
					}}}}}},
				}}},
				bson.D{{"$project", bson.D{
					{"_id", 0},
					{"queued", 1},
					{"sent", 1},
					{"sentLetters", 1},
				}}},
			}},
		}}},
		{{"$project", bson.D{
			{"uniqueRequests", bson.D{{"$arrayElemAt", bson.A{"$uniqueSummary.uniqueRequests", 0}}}},
			{"totalIds", bson.D{{"$arrayElemAt", bson.A{"$uniqueSummary.totalIds", 0}}}},
			{"topIds", 1},
			{"sentIds", bson.D{{"$arrayElemAt", bson.A{"$statusCounts.sentLetters", 0}}}},
			{"statusCounts", bson.D{{"$arrayElemAt", bson.A{"$statusCounts", 0}}}},
		}}},
	}
}

func ByLocationPipeline(groupBy string) (mongo.Pipeline, error) {
	var groupKey any
	var projectFields bson.D

	switch groupBy {
	case "zip":
		groupKey = "$address.zip"
		projectFields = bson.D{
			{"zipCode", "$_id"},
			{"count", 1},
		}
	case "city":
		groupKey = bson.D{
			{"$toLower", bson.D{
				{"$trim", bson.D{
					{"input", "$address.city"},
				}},
			}},
		}
		projectFields = bson.D{
			{"city", "$_id"},
			{"count", 1},
		}
	default:
		return nil, fmt.Errorf("invalid groupBy: must be 'zip' or 'city'")
	}

	return mongo.Pipeline{
		{{"$unwind", "$ids"}},
		{{"$group", bson.D{
			{"_id", groupKey},
			{"count", bson.D{{"$sum", 1}}},
		}}},
		{{"$project", append(projectFields, bson.E{"_id", 0})}},
		{{"$sort", bson.D{{"count", -1}}}},
	}, nil
}

func ByCreationPipeline() mongo.Pipeline {
	return mongo.Pipeline{
		{{"$unwind", "$ids"}},
		{{"$addFields", bson.D{
			{"createdDay", bson.D{
				{"$dateTrunc", bson.D{
					{"date", "$creationdate"},
					{"unit", "day"},
				}},
			}},
		}}},
		{{"$group", bson.D{
			{"_id", "$createdDay"},
			{"count", bson.D{{"$sum", 1}}},
		}}},
		{{"$project", bson.D{
			{"date", "$_id"},
			{"count", 1},
			{"_id", 0},
		}}},
		{{"$sort", bson.D{{"count", -1}}}}, // <-- sort descending by count
	}
}
