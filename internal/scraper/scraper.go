package scraper

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MVitelli/manga-converter/internal/models"
	"github.com/gocolly/colly"
)

// ScrapeMangaID busca y devuelve el identificador único del manga basado en su nombre.
func ScrapeMangaID(mangaName string) (string, error) {
	// Formatear el nombre del manga para la búsqueda (reemplazar espacios por guiones, etc.)
	formattedName := strings.ToLower(strings.ReplaceAll(mangaName, "-", " "))

	// URL de búsqueda en Mangakakalot
	searchURL := fmt.Sprintf("https://ww8.mangakakalot.tv/search/%s", formattedName)

	c := colly.NewCollector(
		colly.AllowedDomains("ww8.mangakakalot.tv", "mangakakalot.tv"),
		colly.UserAgent("MangaAPI/1.0"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*mangakakalot.tv",
		Parallelism: 2,
		Delay:       1 * time.Second,
	})

	var mangaID string

	c.OnHTML("h3.story_name a", func(e *colly.HTMLElement) {
		title := e.Text
		// Comparar el título encontrado con el buscado (ignorando mayúsculas/minúsculas)
		if strings.EqualFold(title, formattedName) {
			href := e.Attr("href")
			// href tiene la estructura: /manga/manga-aa951409/
			parts := strings.Split(href, "/")
			fmt.Printf("parts %s \n", strings.Join(parts, " "))
			if len(parts) >= 3 {
				mangaID = parts[2] // "manga-aa951409"
				log.Println("Manga ID found:", mangaID)
			}
		}
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error al buscar ID del manga: %v en %v\n", err, r.Request.URL)
	})

	err := c.Visit(searchURL)
	if err != nil {
		return "", err
	}

	if mangaID == "" {
		return "", fmt.Errorf("no se encontró el manga '%s'", mangaName)
	}

	return mangaID, nil
}

// ScrapeChapterImages extrae las URLs de las imágenes de un capítulo de manga usando el ID del manga.
func ScrapeChapterImages(mangaID string, chapter string) ([]string, error) {
	// Construir la URL del capítulo
	chapterURL := fmt.Sprintf("https://ww8.mangakakalot.tv/chapter/%s/chapter-%s", mangaID, chapter)

	c := colly.NewCollector(
		colly.AllowedDomains("ww8.mangakakalot.tv", "mangakakalot.tv"),
		colly.UserAgent("MangaAPI/1.0"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*mangakakalot.tv",
		Parallelism: 1, // Procesamiento secuencial
		Delay:       1 * time.Second,
	})

	var imageDataSlice []models.ImageData

	// Expresión regular para extraer el número de página
	pageRegex := regexp.MustCompile(`page\s+(\d+)`)

	c.OnHTML("div.vung-doc img.img-loading", func(e *colly.HTMLElement) {
		altText := e.Attr("alt")
		titleText := e.Attr("title")
		src := e.Attr("src")

		if src == "" {
			// Algunos sitios usan 'data-src' en lugar de 'src'
			src = e.Attr("data-src")
		}

		if src == "" {
			log.Println("No se encontró el atributo src o data-src en una imagen.")
			return
		}

		// Combinar alt y title para mayor precisión
		combinedText := fmt.Sprintf("%s %s", altText, titleText)

		matches := pageRegex.FindStringSubmatch(strings.ToLower(combinedText))
		if len(matches) < 2 {
			log.Printf("No se pudo extraer el número de página de: %s\n", combinedText)
			return
		}

		pageNumber, err := strconv.Atoi(matches[1])
		if err != nil {
			log.Printf("Error al convertir el número de página: %v\n", err)
			return
		}

		log.Printf("Imagen encontrada: Página %d - %s\n", pageNumber, src)

		// Agregar directamente al slice sin ordenar
		imageDataSlice = append(imageDataSlice, models.ImageData{
			PageNumber: pageNumber,
			URL:        src,
		})
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visitando", r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error al extraer imágenes: %v en %v\n", err, r.Request.URL)
	})

	// Iniciar la visita
	err := c.Visit(chapterURL)
	if err != nil {
		return nil, err
	}

	// Esperar a que todas las solicitudes se completen
	c.Wait()

	if len(imageDataSlice) == 0 {
		return nil, fmt.Errorf("no se encontraron imágenes para el capítulo %s del manga ID %s", chapter, mangaID)
	}

	// Extraer las URLs en el orden en que fueron añadidas
	sortedImageURLs := make([]string, len(imageDataSlice))
	for i, img := range imageDataSlice {
		sortedImageURLs[i] = img.URL
	}

	return sortedImageURLs, nil
}
