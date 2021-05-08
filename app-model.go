package main

type Slot struct {
	Center         string
	AvailableSlots string
	Date           string
	Vaccine        string
	Age            string
	FeeType        string
}

type Configuration struct {
	Listeners           []Listeners         `json:"listeners"`
	Notificationconfigs Notificationconfigs `json:"notificationConfigs"`
	Pollinginterval     string              `json:"pollingInterval"`
}
type Filters struct {
	MinAge   int    `json:"minAge"`
	Vaccine  string `json:"vaccine"`
	Fees     string `json:"fees"`
	MinSlots int    `json:"minSlots"`
}
type Listeners struct {
	State     string   `json:"state"`
	District  string   `json:"district"`
	Pin       string   `json:"pin"`
	Receivers []string `json:"receivers"`
	Filters   Filters  `json:"filters,omitempty"`
}
type SMTP struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type Notificationconfigs struct {
	SMTP SMTP `json:"smtp"`
}
