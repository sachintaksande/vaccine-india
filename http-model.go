package main

type CowinStates struct {
	States []States `json:"states"`
	TTL    int      `json:"ttl"`
}

type States struct {
	StateID   int    `json:"state_id"`
	StateName string `json:"state_name"`
}

type CowinDistricts struct {
	Districts []Districts `json:"districts"`
	TTL       int         `json:"ttl"`
}
type Districts struct {
	DistrictID   int    `json:"district_id"`
	DistrictName string `json:"district_name"`
}

type CowinSlots struct {
	Centers []Centers `json:"centers"`
}

type Sessions struct {
	SessionID         string   `json:"session_id"`
	Date              string   `json:"date"`
	AvailableCapacity float64  `json:"available_capacity"`
	MinAgeLimit       int      `json:"min_age_limit"`
	Vaccine           string   `json:"vaccine"`
	Slots             []string `json:"slots"`
}

func (s *Sessions) getRoundedAvailableCapacity() int {
	return int(s.AvailableCapacity)
}

type Centers struct {
	CenterID     int        `json:"center_id"`
	Name         string     `json:"name"`
	Address      string     `json:"address"`
	StateName    string     `json:"state_name"`
	DistrictName string     `json:"district_name"`
	BlockName    string     `json:"block_name"`
	Pincode      int        `json:"pincode"`
	Lat          int        `json:"lat"`
	Long         int        `json:"long"`
	From         string     `json:"from"`
	To           string     `json:"to"`
	FeeType      string     `json:"fee_type"`
	Sessions     []Sessions `json:"sessions"`
}
