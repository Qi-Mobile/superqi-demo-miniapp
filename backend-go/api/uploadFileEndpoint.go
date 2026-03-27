package api

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	uploadDir     = "./uploads"
	maxUploadSize = 10 * 1024 * 1024 // 10MB
)

type UploadFileResponse struct {
	Success    bool   `json:"success"`
	FileName   string `json:"fileName,omitempty"`
	FileSize   int64  `json:"fileSize,omitempty"`
	FileType   string `json:"fileType,omitempty"`
	FilePath   string `json:"filePath,omitempty"`
	MD5        string `json:"md5,omitempty"`
	Message    string `json:"message,omitempty"`
	UploadTime string `json:"uploadTime,omitempty"`
}

func InitUploadFileEndpoint(group fiber.Router) {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("[ERROR] Failed to create upload directory: %v\n", err)
	}

	group.Post("/upload", handleFileUpload)
}

func handleFileUpload(ctx *fiber.Ctx) error {
	log.Println("=================================================================")
	log.Println("FILE UPLOAD REQUEST RECEIVED")
	log.Println("=================================================================")

	// Get the file from form data
	fileKey := ctx.FormValue("fileName", "file")
	fileType := ctx.FormValue("fileType", "")

	log.Printf("[INFO] File key: %s\n", fileKey)
	log.Printf("[INFO] File type: %s\n", fileType)

	// Get the uploaded file
	file, err := ctx.FormFile(fileKey)
	if err != nil {
		log.Printf("[ERROR] Failed to get file from form: %v\n", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(UploadFileResponse{
			Success: false,
			Message: "Failed to get file from form: " + err.Error(),
		})
	}

	// Validate file size
	if file.Size > maxUploadSize {
		log.Printf("[ERROR] File too large: %d bytes (max: %d bytes)\n", file.Size, maxUploadSize)
		return ctx.Status(fiber.StatusBadRequest).JSON(UploadFileResponse{
			Success: false,
			Message: fmt.Sprintf("File too large. Maximum size is %d MB", maxUploadSize/(1024*1024)),
		})
	}

	// Validate file type if provided
	if fileType != "" {
		validTypes := map[string]bool{
			"PDF": true, "DOC": true, "DOCX": true,
			"XLS": true, "XLSX": true, "PPT": true, "PPTX": true,
		}
		if !validTypes[fileType] {
			log.Printf("[WARNING] Unsupported file type: %s\n", fileType)
		}
	}

	log.Printf("[INFO] File name: %s\n", file.Filename)
	log.Printf("[INFO] File size: %d bytes (%.2f KB)\n", file.Size, float64(file.Size)/1024)
	log.Printf("[INFO] Content type: %s\n", file.Header.Get("Content-Type"))

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		log.Printf("[ERROR] Failed to open uploaded file: %v\n", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(UploadFileResponse{
			Success: false,
			Message: "Failed to open uploaded file: " + err.Error(),
		})
	}
	defer src.Close()

	// Generate unique filename to avoid conflicts
	timestamp := time.Now().Format("20060102-150405")
	ext := filepath.Ext(file.Filename)
	baseFilename := file.Filename[:len(file.Filename)-len(ext)]
	uniqueFilename := fmt.Sprintf("%s_%s%s", baseFilename, timestamp, ext)
	destPath := filepath.Join(uploadDir, uniqueFilename)

	// Create destination file
	dst, err := os.Create(destPath)
	if err != nil {
		log.Printf("[ERROR] Failed to create destination file: %v\n", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(UploadFileResponse{
			Success: false,
			Message: "Failed to create destination file: " + err.Error(),
		})
	}
	defer dst.Close()

	// Calculate MD5 hash while copying
	hash := md5.New()
	multiWriter := io.MultiWriter(dst, hash)

	// Copy file content
	bytesWritten, err := io.Copy(multiWriter, src)
	if err != nil {
		log.Printf("[ERROR] Failed to save file: %v\n", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(UploadFileResponse{
			Success: false,
			Message: "Failed to save file: " + err.Error(),
		})
	}

	md5Hash := hex.EncodeToString(hash.Sum(nil))

	log.Println("=================================================================")
	log.Println("FILE UPLOAD SUCCESS")
	log.Println("=================================================================")
	log.Printf("[SUCCESS] File saved to: %s\n", destPath)
	log.Printf("[SUCCESS] Bytes written: %d\n", bytesWritten)
	log.Printf("[SUCCESS] MD5 checksum: %s\n", md5Hash)
	log.Println("=================================================================")

	// Get absolute path for response
	absPath, _ := filepath.Abs(destPath)

	response := UploadFileResponse{
		Success:    true,
		FileName:   uniqueFilename,
		FileSize:   file.Size,
		FileType:   fileType,
		FilePath:   absPath,
		MD5:        md5Hash,
		Message:    "File uploaded successfully",
		UploadTime: time.Now().Format(time.RFC3339),
	}

	return ctx.JSON(response)
}
