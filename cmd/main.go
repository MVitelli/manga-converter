package main

import (
	"net/http"

	"github.com/MVitelli/manga-converter/internal/models"
	"github.com/MVitelli/manga-converter/internal/scraper"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Definir el endpoint para obtener imágenes
	router.GET("/mangas/:name/chapters/:chapter/images", getMangaChapterImages)

	// Ejecutar el servidor en el puerto 8088
	router.Run(":8088")
}

// getMangaChapterImages maneja las solicitudes GET para obtener imágenes de un capítulo específico.
func getMangaChapterImages(c *gin.Context) {
	mangaName := c.Param("name")
	chapter := c.Param("chapter")

	if mangaName == "" || chapter == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetros 'name' y 'chapter' son requeridos"})
		return
	}

	// Buscar el ID del manga
	mangaID, err := scraper.ScrapeMangaID(mangaName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Llamar a la función de scraping para obtener las URLs de las imágenes
	images, err := scraper.ScrapeChapterImages(mangaID, chapter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Crear la estructura de respuesta
	response := models.MangaChapter{
		MangaName: mangaName,
		Chapter:   chapter,
		ImageURLs: images,
	}

	// Devolver la respuesta en formato JSON
	c.JSON(http.StatusOK, response)
}
