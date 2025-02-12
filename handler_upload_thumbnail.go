package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// upload
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}

	mediaType := header.Header.Get("Content-Type")
	// fileSlice, err := io.ReadAll(file)
	/*
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Unable to read file", err)
			return
		}
		defer file.Close()
	*/

	video, _ := cfg.db.GetVideo(videoID)
	if userID != video.UserID {
		respondWithError(w, http.StatusUnauthorized, "Not the video owner", err)
		return
	}

	mimeType, _, _ := mime.ParseMediaType(mediaType)

	fmt.Printf("%s", mimeType)

	if (mimeType != "image/png") && (mimeType != "image/jpeg") {
		respondWithError(w, http.StatusBadRequest, "Only jpeg and png are allowed to upload", err)
		return
	}

	extension := strings.TrimPrefix(mediaType, "image/")
	fileName := fmt.Sprintf("%s.%s", videoIDString, extension)
	fullFilePath := filepath.Join(cfg.assetsRoot, fileName)
	fmt.Printf("FULL PATH: %s", fullFilePath)

	dst, _ := os.Create(fullFilePath)
	defer dst.Close()

	io.Copy(dst, file)

	port := 8091
	newThumbUrl := fmt.Sprintf("http://localhost:%d/assets/%s.%s", port, videoIDString, extension)

	// base64Image := base64.StdEncoding.EncodeToString(fileSlice)
	// dataUrl := fmt.Sprintf("data:%s;base64,%s", mediaType, base64Image)
	video.ThumbnailURL = &newThumbUrl
	newVideo := cfg.db.UpdateVideo(video)

	respondWithJSON(w, http.StatusOK, newVideo)
}
