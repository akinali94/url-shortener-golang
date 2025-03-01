package main

type Item struct{
	ID int `json: id`
	Name string `json: id`
}

type URLMapping struct{
	ShortUrl string `json:"shortUrl"`
	LongUrl string `json:"longUrl"`
}