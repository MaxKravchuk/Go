package models

type Request struct {
	RequestId  int    `json:"request_id,omitempty"`
	UrlPackage []int  `json:"url_package"`
	Ip         string `json:"ip"`
}
