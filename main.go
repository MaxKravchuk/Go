package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Urls struct {
	gorm.Model
	id  int `gorm:"primaryKey`
	Url string
}

func setup(db *gorm.DB) {
	db.AutoMigrate(&Urls{})
	seed(db)
}

func seed(db *gorm.DB) {
	urls := []Urls{
		{Url: "http://inv-nets.admixer.net/test-dsp/dsp?responseType=1&profile=1"},
		{Url: "http://inv-nets.admixer.net/test-dsp/dsp?responseType=1&profile=2"},
		{Url: "http://inv-nets.admixer.net/test-dsp/dsp?responseType=1&profile=3"},
		{Url: "http://inv-nets.admixer.net/test-dsp/dsp?responseType=1&profile=4"},
	}

	for _, url := range urls {
		db.Create(&url)
	}
}

func main() {
	dsn := "host=localhost user=postgres password=admin dbname=Go"
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		panic("failed to connect database")
	}
	setup(db)
	
}
