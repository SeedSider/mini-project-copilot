package models

// ExchangeRate represents a currency exchange rate record.
type ExchangeRate struct {
	ID          string  `json:"id"`
	Country     string  `json:"country"`
	Currency    string  `json:"currency"`
	CountryCode string  `json:"countryCode"`
	Buy         float64 `json:"buy"`
	Sell        float64 `json:"sell"`
}

// InterestRate represents a deposit interest rate record.
type InterestRate struct {
	ID      string  `json:"id"`
	Kind    string  `json:"kind"`
	Deposit string  `json:"deposit"`
	Rate    float64 `json:"rate"`
}

// Branch represents a bank branch record.
type Branch struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Distance  string  `json:"distance"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
