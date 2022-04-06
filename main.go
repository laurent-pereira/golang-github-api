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
	"github.com/patrickmn/go-cache"
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

// Language is the single language data structure
type Language struct {
	Language string
	Bytes int
}

// Stats contains the Stats data structure
type Stats struct {
	Languages map[string]int
}

// JSONData contains the GitHub API response
type JSONData struct {
	Count int `json:"total_count"`
	Items []Item
}

// Filters is the data structure for the filters
type Filters struct {
	Owner string
	License string
	Language string
	Repository string
}

var ItemCache CacheItf

func main() {
	log := logger.Default()
	
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}

	if os.Getenv("GITHUB_TOKEN") == "" {
		panic("GITHUB_TOKEN is not set")
	}

	log.Info("Initializing app")
	InitCache() // Initialize the cache

	cfg, err := NewConfig()
	if err != nil {
		log.WithError(err).Error("Fail to initialize configuration")
		os.Exit(-1)
	}

	log.Info("Initializing routes")
	router := handlers.NewRouter(log)
	router.HandleFunc("/ping", PongHandler)
	router.HandleFunc("/repos", repoHandler)
	router.HandleFunc("/stats", statHandler)

	log.WithField("port", cfg.Port).Info("Listening...")
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), router)
}

type CacheItf interface {
	Set(key string, data interface{}, expiration time.Duration) error
	Get(key string) ([]byte, error)
}

type AppCache struct {
	client *cache.Cache
}

func InitCache() {
	ItemCache = &AppCache{
		client: cache.New(5*time.Minute, 10*time.Minute),
	}
}

func (r *AppCache) Set(key string, data interface{}, expiration time.Duration) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	r.client.Set(key, b, expiration)
	return nil
}

func (r *AppCache) Get(key string) ([]byte, error) {
	res, exist := r.client.Get(key)
	if !exist {
		return nil, nil
	}

	resByte, ok := res.([]byte)
	if !ok {
		return nil, nil
	}

	return resByte, nil
}

/*
	Set the filters for the query
*/
func setFilters(filters Filters) (string) {
	var query string = "stars:>=0" // Default filter

	if filters.Repository != "" {
		query = query + "%20repo:" + filters.Repository
	} else if filters.Owner != "" {
		query = query + "%20user:" + filters.Owner
	} 
	if filters.License != "" {
		query = query + "%20license:" + filters.License
	}
	if filters.Language != "" {
		query = query + "%20language:" + filters.Language
	}

	return query
}

/* 
	Fetch the last 100 repositories from GitHub API with filters
	https://docs.github.com/en/rest/reference/repos#list-public-repositories
*/
func fechLastRepositories(params Filters) (JSONData) {
	log := logger.Default()
	// Execution time of the function
	start := time.Now()
	defer func() {
		fmt.Println("fechLastRepositories Execution Time: ", time.Since(start))
	}()
	
	query := setFilters(params)
	var url = fmt.Sprintf("https://api.github.com/search/repositories?q=%s&sort=updated&order=desc&per_page=50", query)

	req, err := http.NewRequest("GET", url, nil)
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

	data := JSONData{}
	json.Unmarshal(body, &data)

	// Use WaitGroup to wait for all goroutines to finish
	wg := sync.WaitGroup{}
	for index, repository := range data.Items {
		wg.Add(1)
		go func(repository Item, index int) {
			if res, _ := ItemCache.Get(repository.FullName); res != nil {
				log.Info("Cache hit for ", repository.FullName)
				item := Item{}
				json.Unmarshal(res, &item)
				data.Items[index] = item
			} else {
				// Get the languages for the repository
				log.Info("API Call for ", repository.FullName)
				data.Items[index].Languages = fetchLanguages(map[string]string{"languages_url": repository.LanguagesUrl})
				ItemCache.Set(repository.FullName, data.Items[index], cache.DefaultExpiration)
			}
			wg.Done()
		}(repository, index)
	}
	wg.Wait()
	return data
}

/*
	Fetch the languages of a repository from GitHub API
*/
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

/*
	Aggregate Stats language
*/
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
/*
	Function for the /stat endpoint
*/
func statHandler(w http.ResponseWriter, r *http.Request, params map[string]string) error {
	// Execution time of the function
	start := time.Now()
	defer func() {
		fmt.Println("statHandler Execution Time: ", time.Since(start))
	}()

	decoder := json.NewDecoder(r.Body)
	var filters Filters
	decoder.Decode(&filters)

	data := fechLastRepositories(filters)
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

/*
	Fuction for the /repo endpoint
*/
func repoHandler(w http.ResponseWriter, r *http.Request, params map[string]string) error {
	// Execution time of the function
	start := time.Now()
	defer func() {
		fmt.Println("repoHandler Execution Time: ", time.Since(start))
	}()
	
	decoder := json.NewDecoder(r.Body)
	var filters Filters
	decoder.Decode(&filters)

	data := fechLastRepositories(filters)

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
