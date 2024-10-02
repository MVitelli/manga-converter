package models

type MangaChapter struct {
	MangaName string   `json:"manga_name"`
	Chapter   string   `json:"chapter"`
	ImageURLs []string `json:"image_urls"`
}
