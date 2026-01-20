package api

import (
	"context"
	"fmt"
	"go-chat/internals/store"
	"go-chat/internals/utils"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const (
	MaxUploadSize = 100 << 20
	UploadPath    = "./uploads"
)

type MediaHandler struct {
	MessageHandler      *MessageHandler
	ConversationHandler *ConversationHandler
	Logger              *log.Logger
}

func NewMediaHandler(messageHandler *MessageHandler, conversationHandler *ConversationHandler, logger *log.Logger) *MediaHandler {
	os.MkdirAll(filepath.Join(UploadPath, "images"), 0755)
	os.MkdirAll(filepath.Join(UploadPath, "videos"), 0755)
	os.MkdirAll(filepath.Join(UploadPath, "files"), 0755)
	return &MediaHandler{
		MessageHandler:      messageHandler,
		ConversationHandler: conversationHandler,
		Logger:              logger,
	}
}

func (h *MediaHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		http.Error(w, "file to large", http.StatusBadRequest)
		return
	}
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	conversationID := r.FormValue("conversation_id")
	recipientID := r.FormValue("recipient_id")
	caption := r.FormValue("caption")
	if conversationID == "" && recipientID == "" {
		http.Error(w, "conversation_id or recipient_id required", http.StatusBadRequest)
		return
	}
	contentType := fileHeader.Header.Get("Content-Type")
	var messageType store.MessageType
	var subdir string

	switch {
	case strings.HasPrefix(contentType, "image/"):
		messageType = store.MessageTypeImage
		subdir = "images"
	case strings.HasPrefix(contentType, "video/"):
		messageType = store.MessageTypeVideo
		subdir = "videos"
	default:
		messageType = store.MessageTypeFile
		subdir = "files"
	}

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filepath := filepath.Join(UploadPath, subdir, filename)
	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	fileSize, err := io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("userID").(string)
	senderUUID, _ := uuid.Parse(userID)

	// Construct media URL (relative path or full URL if using CDN)
	mediaURL := fmt.Sprintf("/uploads/%s/%s", subdir, filename)
	mediaSize := fileSize

	// Save message to database
	ctx := context.Background()
	var message *store.Message
	var recipients []uuid.UUID
	if recipientID != "" {
		// Direct message
		recipientUUID, _ := uuid.Parse(recipientID)
		message, recipients, err = h.MessageHandler.SendDirectMessage(
			ctx,
			senderUUID,
			recipientUUID,
			caption,
			messageType,
			&mediaURL,
			&mediaSize,
			&contentType,
		)
	} else {
		// Group message
		conversationUUID, _ := uuid.Parse(conversationID)
		message, recipients, err = h.MessageHandler.SendGroupMessage(
			ctx,
			senderUUID,
			conversationUUID,
			caption,
			messageType,
			&mediaURL,
			&mediaSize,
			&contentType,
		)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := map[string]interface{}{
		"message_id": message.ID.String(),
		"media_url":  mediaURL,
		"file_size":  fileSize,
		"mime_type":  contentType,
		"recipients": recipients,
	}
	utils.WriteJSON(w, http.StatusAccepted, response)
}

func (h *MediaHandler) ServeMedia(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.Contains(path, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	filepath := strings.TrimPrefix(path, "/")
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, filepath)
}
