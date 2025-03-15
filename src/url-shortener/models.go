package urlshortener

import "time"

type IdGeneratorResponseModel struct {
	ID uint64 `json:"ID"`
}

type URLMapping struct {
	ID        string    `bson:"_id" json:"id"`
	ShortUrl  string    `bson:"shortUrl" json:"shortUrl"`
	LongUrl   string    `bson:"longUrl" json:"longUrl"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
