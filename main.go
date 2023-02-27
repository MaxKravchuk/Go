package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

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

func GetUrls(ids []int, db *gorm.DB) []string {
	var urls []models.Urls
	db.Where("id IN ?", ids).Find(&urls)
	var result []string
	for _, url := range urls {
		result = append(result, url.Url)
	}
	return result
}

func main() {
	dsn := "host=localhost user=postgres password=admin dbname=Go"
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		panic("failed to connect database")
	}
	setup(db)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handlePost(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Listening on :8080...")
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {

	if r.Method == "POST" {
		handlePost(w, r, db)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

/*
func handleGet(w http.ResponseWriter, r *http.Request) {
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
	handleRequest(Request{
		RequestId:  requestId,
		UrlPackage: urlPackage,
		Ip:         ip,
	}, w)
}*/

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
	if req.Ip != "" {
		if net.ParseIP(req.Ip) == nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
	}
	if len(req.UrlPackage) == 0 {
		http.Error(w, "No urls provided", http.StatusNoContent)
		return
	}
	prices := make([]float64, len(req.UrlPackage))
	urls := GetUrls(req.UrlPackage, db)
	println(urls)
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
