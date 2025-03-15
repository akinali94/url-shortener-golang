package urlshortener

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/akinali94/url-shortener-golang/pkg/repository"
)

type Service struct {
	repo *repository.GenericMongoRepo[URLMapping]
}

func NewService(r *repository.GenericMongoRepo[URLMapping]) *Service {
	return &Service{
		repo: r,
	}
}

func (s *Service) getLongUrl(shortUrl string) (string, error) {

	res, err := s.repo.GetByField(shortUrl, "shortUrl")
	if err != nil {
		fmt.Println("Failed to fetch shortURL from repository")
	}

	return res.LongUrl, nil
}

func (s *Service) generateShortUrl(longUrl string) (string, error) {

	resp, err := http.Get("http://localhost:8081/getid")
	if err != nil {
		fmt.Printf("Failed to fetch ID from id-generator-service, err: %s", err)
		return "", err
	}
	defer resp.Body.Close()

	var idModel IdGeneratorResponseModel
	err = json.NewDecoder(resp.Body).Decode(&idModel)
	if err != nil {
		fmt.Printf("error on 45. err: %s", err)
		return "", err
	}
	fmt.Println(idModel)

	/*
		id, err := strconv.ParseUint(idModel.ID, 10, 64)
		if err != nil {
			fmt.Printf("error on 52. err: %s", err)
			return "", err
		}
		fmt.Println(id)
	*/

	shortUrl := Base10toBase58(idModel.ID)

	urlMapping := URLMapping{
		ShortUrl:  shortUrl,
		LongUrl:   longUrl,
		CreatedAt: time.Now(),
	}

	_, err = s.repo.Add(urlMapping)
	if err != nil {
		fmt.Println("Failed to add repository")
	}

	return shortUrl, nil
}
