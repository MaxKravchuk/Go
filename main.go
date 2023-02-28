package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setup(db *gorm.DB) {
	if (!db.Migrator().HasTable(&models.Urls{})) {
		db.AutoMigrate(&models.Urls{})
		seed(db)
	}
}

func seed(db *gorm.DB) {
	urls := []models.Urls{
		{Url: "http://inv-nets.admixer.net/test-dsp/dsp?responseType=1&profile=1"},
		{Url: "http://inv-nets.admixer.net/test-dsp/dsp?responseType=1&profile=2"},
		{Url: "http://inv-nets.admixer.net/test-dsp/dsp?responseType=1&profile=3"},
		{Url: "http://inv-nets.admixer.net/test-dsp/dsp?responseType=1&profile=4"},
	}

	for _, url := range urls {
		db.Create(&url)
	}
}

func GetUrls(ids []int, db *gorm.DB) ([]string, error) {
	var urls []models.Urls
	err := db.Where("id IN ?", ids).Find(&urls).Error
	if err != nil {
		return nil, err
	}
	if len(urls) != len(ids) {
		return nil, fmt.Errorf("not all ids found in the database")
	}
	var result []string
	for _, url := range urls {
		result = append(result, url.Url)
	}
	return result, nil
}

func main() {
	dsn := "host=localhost user=postgres password=admin dbname=Go"
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		panic("failed to connect database")
	}
	setup(db)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, db)
	})

	fmt.Println("Listening on :8080...")
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {

	if r.Method == "POST" {
		handlePost(w, r, db)
	} else if r.Method == "GET" {
		handleGet(w, r, db)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGet(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	query := r.URL.Query()
	requestId, err := strconv.Atoi(query.Get("request_id"))
	if err != nil {
		requestId = 0
	}
	urlPackageStr := query.Get("url_package")
	urlPackage := make([]int, 0)
	for _, idStr := range strings.Split(urlPackageStr, ",") {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		urlPackage = append(urlPackage, id)
	}
	ip := query.Get("ip")

	var req models.Request = models.Request{
		RequestId:  requestId,
		UrlPackage: urlPackage,
		Ip:         ip,
	}

	handleRequest(req, w, db)
}

func handlePost(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	decoder := json.NewDecoder(r.Body)
	var req models.Request
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	handleRequest(req, w, db)
}

func handleRequest(req models.Request, w http.ResponseWriter, db *gorm.DB) {
	if !req.ValidateRequest() {
		http.Error(w, "Invalid IP or url_package is empty", http.StatusNoContent)
		return
	}
	prices := make([]float64, len(req.UrlPackage))
	urls, error := GetUrls(req.UrlPackage, db)
	if error != nil {
		http.Error(w, "Invalid id in package", http.StatusNoContent)
		return
	}
	for i, _ := range urls {
		resp, err := http.Get(urls[i])
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		var priceResp models.Response
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&priceResp)
		if err != nil {
			prices[i] = 0
		} else {
			prices[i] = priceResp.Price
		}
	}
	maxPrice := 0.0
	for _, price := range prices {
		if price > maxPrice {
			maxPrice = price
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.Response{
		Price: maxPrice,
	})
}
