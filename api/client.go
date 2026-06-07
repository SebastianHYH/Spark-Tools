package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

const baseURL = "https://fortnitecontent-website-prod07.ol.epicgames.com/content/api/pages/fortnite-game/spark-tracks"

func Fetch() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(body))
}