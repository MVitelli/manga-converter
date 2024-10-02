package pdf

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	_ "image/jpeg"
	_ "image/png"

	"github.com/jung-kurt/gofpdf"
)

// ImagesToPDF convierte una lista de rutas de imágenes en un archivo PDF.
func ImagesToPDF(imagePaths []string, outputPath string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(false, 0)

	for _, imgPath := range imagePaths {
		// Abrir la imagen para obtener su tamaño
		imgFile, err := os.Open(imgPath)
		if err != nil {
			return fmt.Errorf("error abriendo la imagen %s: %v", imgPath, err)
		}
		img, _, err := image.DecodeConfig(imgFile)
		imgFile.Close()
		if err != nil {
			return fmt.Errorf("error decodificando la imagen %s: %v", imgPath, err)
		}

		// Calcular proporciones para ajustar la imagen en la página A4
		pageWidth, pageHeight := 210.0, 297.0 // A4 en mm
		imgWidth, imgHeight := float64(img.Width), float64(img.Height)

		// Calcular la escala para ajustar la imagen
		scale := 1.0
		if imgWidth > pageWidth || imgHeight > pageHeight {
			scaleW := pageWidth / imgWidth
			scaleH := pageHeight / imgHeight
			if scaleW < scaleH {
				scale = scaleW
			} else {
				scale = scaleH
			}
		}

		// Añadir una nueva página
		pdf.AddPage()

		// Añadir la imagen
		pdf.ImageOptions(
			imgPath,
			0, 0,
			imgWidth*scale, imgHeight*scale,
			false,
			gofpdf.ImageOptions{ImageType: getImageType(imgPath), ReadDpi: true},
			0,
			"",
		)
	}

	// Guardar el PDF
	err := pdf.OutputFileAndClose(outputPath)
	if err != nil {
		return fmt.Errorf("error guardando el PDF: %v", err)
	}

	return nil
}

// getImageType devuelve el tipo de imagen basado en la extensión del archivo.
func getImageType(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".jpg", ".jpeg":
		return "JPG"
	case ".png":
		return "PNG"
	default:
		return "JPG"
	}
}
