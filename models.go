package main

import "time"

type Item struct{
	ID int `json: id`
	Name string `json: id`
}

type URLMapping struct{
	ShortUrl string `json:"shortUrl"`
	LongUrl string `json:"longUrl"`
	CreatedAt time.Time `json:"createdAt`
}