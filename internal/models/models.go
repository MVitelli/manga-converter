package models

type Manga struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type MangaChapter struct {
	MangaName string   `json:"manga_name"`
	Chapter   string   `json:"chapter"`
	ImageURLs []string `json:"image_urls"`
}
