package main

import (
	"encoding/json"
	"fmt"
	"time"
	"sync"
	"io/ioutil"
	"net/http"
	"os"
	"github.com/joho/godotenv"
	"github.com/Scalingo/go-handlers"
	"github.com/Scalingo/go-utils/logger"
)

// Item is the single repository data structure
type Item struct {
	ID              int
	Name            string
	FullName        string `json:"full_name"`
	GitUrl				 string `json:"git_url"`
	HtmlUrl				string `json:"html_url"`
	LanguagesUrl	 string `json:"languages_url"`
	CreatedAt       string `json:"created_at"`
	PushedAt string `json:"pushed_at"`
	UpdatedAt string `json:"updated_at"`
	MainLanguage string `json:"language"`
	Languages			 map[string]int
}

type Language struct {
	Language string
	Bytes int
}

// Stats contains the Stats data structure
type Stats struct {
	//Languages []Language
	Languages map[string]int
}

// JSONData contains the GitHub API response
type JSONData struct {
	Count int `json:"total_count"`
	Items []Item
}

func main() {
	log := logger.Default()
	
	err := godotenv.Load(".env")
	if err != nil {
			log.Fatal("Error loading .env file")
	}
	log.Info("Initializing app")
	
	cfg, err := NewConfig()
	if err != nil {
		log.WithError(err).Error("Fail to initialize configuration")
		os.Exit(-1)
	}

	log.Info("Initializing routes")
	router := handlers.NewRouter(log)
	router.HandleFunc("/ping", PongHandler)
	router.HandleFunc("/repos", repoHandler)
	router.HandleFunc("/repos/{owner}", repoHandler)
	//router.HandleFunc("/repos/{owner}/{repository}", repoHandler)
	//router.HandleFunc("/repos/{owner}/{repository}/stats", statHandler)
	//router.HandleFunc("/repos/{owner}/{repository}/stats/{language}", statHandler)
	//router.HandleFunc("/repos/{owner}/{repository}/stats/{language}/{period}", statHandler)
	router.HandleFunc("/stats", statHandler)

	log.WithField("port", cfg.Port).Info("Listening...")
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), router)
}

// https://docs.github.com/en/rest/reference/repos#list-public-repositories
func fechLastRepositories(params map[string]string) (JSONData) {
	// Execution time of the function
	start := time.Now()
	defer func() {
		fmt.Println("Execution Time: ", time.Since(start))
	}()
	
	req, err := http.NewRequest("GET", "https://api.github.com/search/repositories?q=stars:>=0&sort=updated&order=desc&per_page=100", nil)
	req.Header.Add("Authorization", "Bearer " + os.Getenv("GITHUB_TOKEN"))
	//fmt.Println(os.Getenv("GITHUB_TOKEN"))
	if err != nil {
		fmt.Printf("%s", err)
 }

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s", err)
 }
	
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		fmt.Printf("%s", err)
 	}

	data := JSONData{}
	json.Unmarshal(body, &data)

	// Use WaitGroup to wait for all goroutines to finish
	wg := sync.WaitGroup{}
	for index, repository := range data.Items {
		wg.Add(1)
		go func(repository Item, index int) {
			data.Items[index].Languages = fetchLanguages(map[string]string{"languages_url": repository.LanguagesUrl})
			wg.Done()
		}(repository, index)
	}
	wg.Wait()
	return data
}

func fetchLanguages(params map[string]string) (map[string]int) {
	req, err := http.NewRequest("GET", params["languages_url"], nil)
	req.Header.Add("Authorization", "Bearer " + os.Getenv("GITHUB_TOKEN"))
	if err != nil {
		fmt.Printf("%s", err)
 	}
	
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		fmt.Printf("%s", err)
 	}

	data := map[string]int{}
	json.Unmarshal(body, &data)
	return data
}

func aggregateStat(item Item, stats Stats) (Stats) {
	for language, bytes := range item.Languages {
		if _, ok := stats.Languages[language]; ok {
			stats.Languages[language] += bytes
		} else {
			stats.Languages[language] = bytes
		}
	}
	return stats
}

func statHandler(w http.ResponseWriter, r *http.Request, params map[string]string) error {
	data := fechLastRepositories(params)
	stats := Stats{}
	stats.Languages = make(map[string]int)
	// Use WaitGroup to wait for all goroutines to finish
	wg := sync.WaitGroup{}
	for _, item := range data.Items {
		wg.Add(1)
		go func(item Item) {
			aggregateStat(item, stats)
			wg.Done()
		}(item)
		wg.Wait()
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	j, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		fmt.Printf("%s", err)
 	}
	w.Write(j)
	return nil
}

func repoHandler(w http.ResponseWriter, r *http.Request, params map[string]string) error {
	//log := logger.Get(r.Context())
	data := fechLastRepositories(params)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	j, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("%s", err)
 	}

	w.Write(j)
	return nil
}

func PongHandler(w http.ResponseWriter, r *http.Request, params map[string]string) error {
	log := logger.Get(r.Context())
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(map[string]string{"status": "pong"})
	if err != nil {
		log.WithError(err).Error("Fail to encode JSON")
	}
	return nil
}
