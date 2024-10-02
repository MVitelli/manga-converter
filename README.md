# manga-converter

Init the app: 

```bash
go run cmd/main.go
```

## Endpoints

### Retrieve images from a manga specific chapter 

```go
router.GET("/mangas/:name/chapters/:chapter/images", getMangaChapterImages)
```

Example request:
```curl
curl http://localhost:8088/mangas/one-piece/chapters/1/images   
```

### Retrieve a PDF from a manga specific chapter

```go
router.GET("/mangas/:name/chapters/:chapter/pdf", getMangaChapterPDF)
```

Example request:
```curl
curl -OJ http://localhost:8088/mangas/one-piece/chapters/1000/pdf   
```
