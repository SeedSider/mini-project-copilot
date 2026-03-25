package models

// Menu represents a homepage menu item record in the database.
type Menu struct {
	ID       string `json:"id"`
	Index    int    `json:"index"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	IconURL  string `json:"icon_url"`
	IsActive bool   `json:"is_active"`
}

// MenuResponse wraps a list of menus for the API response.
type MenuResponse struct {
	Menus []Menu `json:"menus"`
}
