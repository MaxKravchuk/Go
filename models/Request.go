package models

import "net"

type Request struct {
	RequestId  int    `json:"request_id,omitempty"`
	UrlPackage []int  `json:"url_package"`
	Ip         string `json:"ip"`
}

func (r Request) ValidateRequest() bool {
	if r.Ip != "" {
		if net.ParseIP(r.Ip) == nil {
			return false
		}
	}
	if len(r.UrlPackage) == 0 {
		return false
	}
	return true
}
