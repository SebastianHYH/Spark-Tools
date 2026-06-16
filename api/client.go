package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const baseURL = "https://fortnitecontent-website-prod07.ol.epicgames.com/content/api/pages/fortnite-game/spark-tracks"
const localPath = "./data/spark-tracks.json"

/*
Fetch retrieves data from baseURL
*/
func Fetch() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		fmt.Println("Fetching from local storage...")
		fetchFromLocal()
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error fetching from remote:", err)
		fmt.Println("Fetching from local storage...")
		fetchFromLocal()
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(body))

	// save to local for local fetching
	err = saveToLocal(body)
	if err != nil {
		log.Fatal(err)
	}
}

/*
saveToLocal saves data to json file
*/
func saveToLocal(data []byte) error {
	file, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

/*
fetchFromLocal fetches data from the local json file
*/
func fetchFromLocal() {
	file, err := os.Open(localPath)
	if err != nil {
		log.Fatal("No remote connection and no local cache available:", err)
	}
	defer file.Close()

	body, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(body))
}