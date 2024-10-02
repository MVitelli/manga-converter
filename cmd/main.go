package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/MVitelli/manga-converter/internal/models"
	"github.com/MVitelli/manga-converter/internal/pdf"
	"github.com/MVitelli/manga-converter/internal/scraper"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Definir el endpoint para obtener imágenes
	router.GET("/mangas/:name/chapters/:chapter/images", getMangaChapterImages)

	// Definir el endpoint para obtener PDF
	router.GET("/mangas/:name/chapters/:chapter/pdf", getMangaChapterPDF)

	// Ejecutar el servidor en el puerto 8080
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

// getMangaChapterPDF maneja las solicitudes GET para obtener un PDF de un capítulo específico.
func getMangaChapterPDF(c *gin.Context) {
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

	// Crear un directorio temporal para almacenar las imágenes descargadas
	tempDir, err := os.MkdirTemp("", "manga-images-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando directorio temporal"})
		return
	}
	defer os.RemoveAll(tempDir) // Eliminar el directorio temporal después

	var downloadedImages []string

	// Descargar las imágenes de manera concurrente
	downloadedImages, err = downloadImagesConcurrently(images, tempDir)
	if err != nil {
		log.Printf("Error al descargar imágenes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al descargar imágenes"})
		return
	}

	if len(downloadedImages) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron descargar imágenes"})
		return
	}

	// Definir la ruta del PDF de salida
	pdfPath := filepath.Join(tempDir, "chapter.pdf")

	// Convertir las imágenes a PDF
	err = pdf.ImagesToPDF(downloadedImages, pdfPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generando el PDF"})
		return
	}

	// Definir el nombre con el que se descargará el PDF
	downloadName := fmt.Sprintf("%s_Chapter_%s.pdf", strings.ReplaceAll(mangaName, " ", "_"), chapter)

	// Enviar el archivo al cliente con el nombre especificado
	c.FileAttachment(pdfPath, downloadName)
}

// downloadImage descarga una imagen desde una URL y la guarda en la ruta especificada.
func downloadImage(url, dest string) error {
	// Crear el archivo de destino
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Obtener la imagen
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Verificar el status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error al descargar imagen: status %d", resp.StatusCode)
	}

	// Copiar el contenido al archivo
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func downloadImagesConcurrently(urls []string, destDir string) ([]string, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var downloaded []string
	var errs []error

	// Limitar el número de goroutines simultáneas
	semaphore := make(chan struct{}, 5) // Por ejemplo, máximo 5 descargas simultáneas

	for idx, url := range urls {
		wg.Add(1)
		go func(idx int, url string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Adquirir el semáforo
			defer func() { <-semaphore }() // Liberar el semáforo

			// Obtener la extensión de la imagen
			ext := filepath.Ext(url)
			if ext == "" {
				ext = ".jpg" // Por defecto
			}

			// Crear el nombre del archivo
			imgPath := filepath.Join(destDir, fmt.Sprintf("%03d%s", idx+1, ext))

			// Descargar la imagen
			err := downloadImage(url, imgPath)
			if err != nil {
				log.Printf("Error descargando imagen %s: %v", url, err)
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}

			mu.Lock()
			downloaded = append(downloaded, imgPath)
			mu.Unlock()
		}(idx, url)
	}

	wg.Wait()

	if len(errs) > 0 {
		return downloaded, fmt.Errorf("ocurrieron errores al descargar algunas imágenes")
	}

	return downloaded, nil
}
