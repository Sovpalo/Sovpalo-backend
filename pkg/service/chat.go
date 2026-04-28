package service

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
	"github.com/jackc/pgx/v5"
)

var (
	ErrChatMessageTextRequired    = errors.New("message text is required")
	ErrChatMessageContentRequired = errors.New("message must contain text or media")
	ErrChatMessageMixedContent    = errors.New("message cannot contain both text and media")
	ErrChatMessageTooManyFiles    = errors.New("too many media files")
	ErrChatMediaTooLarge          = errors.New("chat media file is too large")
	ErrChatMediaInvalidType       = errors.New("chat media must be an image or mp4/webm/mov video")
	ErrChatReadMessageIDsRequired = errors.New("message_ids are required")
	ErrChatMessageNotFound        = errors.New("message not found")
	ErrChatDeleteForbidden        = errors.New("only author can delete message")
)

const (
	maxChatAttachmentsPerMessage = 10
	maxChatImageSize             = 10 << 20
	maxChatVideoSize             = 100 << 20
)

type ChatService struct {
	repo repository.Chat
}

func NewChatService(repo repository.Chat) *ChatService {
	return &ChatService{repo: repo}
}

func (s *ChatService) ListCompanyMessages(companyID int64, userID int64, beforeMessageID int64, limit int) ([]model.ChatMessageView, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.ListCompanyChatMessages(companyID, userID, beforeMessageID, limit)
}

func (s *ChatService) CreateCompanyTextMessage(companyID int64, userID int64, input model.ChatMessageCreateInput) (model.ChatMessageView, error) {
	text := strings.TrimSpace(input.Text)
	if text == "" {
		return model.ChatMessageView{}, ErrChatMessageTextRequired
	}
	return s.repo.CreateCompanyChatMessage(companyID, userID, &text, nil)
}

func (s *ChatService) CreateCompanyMediaMessage(companyID int64, userID int64, files []ChatUploadFile) (model.ChatMessageView, error) {
	if len(files) == 0 {
		return model.ChatMessageView{}, ErrChatMessageContentRequired
	}
	if len(files) > maxChatAttachmentsPerMessage {
		return model.ChatMessageView{}, ErrChatMessageTooManyFiles
	}

	attachments := make([]model.ChatAttachmentCreate, 0, len(files))
	savedURLs := make([]string, 0, len(files))
	for idx, file := range files {
		attachment, err := saveChatMediaFile(companyID, userID, idx, file)
		if err != nil {
			for _, url := range savedURLs {
				_ = removeChatMediaByURL(url)
			}
			return model.ChatMessageView{}, err
		}
		savedURLs = append(savedURLs, attachment.FileURL)
		attachments = append(attachments, attachment)
	}

	message, err := s.repo.CreateCompanyChatMessage(companyID, userID, nil, attachments)
	if err != nil {
		for _, url := range savedURLs {
			_ = removeChatMediaByURL(url)
		}
		return model.ChatMessageView{}, err
	}

	return message, nil
}

func (s *ChatService) DeleteCompanyMessage(companyID int64, messageID int64, userID int64) error {
	attachments, err := s.repo.DeleteCompanyChatMessage(companyID, messageID, userID)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return ErrChatMessageNotFound
		case err.Error() == ErrChatDeleteForbidden.Error():
			return ErrChatDeleteForbidden
		default:
			return err
		}
	}

	for _, attachment := range attachments {
		_ = removeChatMediaByURL(attachment.FileURL)
	}
	return nil
}

func (s *ChatService) MarkCompanyMessagesRead(companyID int64, userID int64, input model.ChatMarkReadInput) (model.ChatReadResult, error) {
	if len(input.MessageIDs) == 0 {
		return model.ChatReadResult{}, ErrChatReadMessageIDsRequired
	}

	inserted, readAt, err := s.repo.MarkCompanyChatMessagesRead(companyID, userID, input.MessageIDs)
	if err != nil {
		return model.ChatReadResult{}, err
	}
	unreadCount, err := s.repo.GetCompanyChatUnreadCount(companyID, userID)
	if err != nil {
		return model.ChatReadResult{}, err
	}

	return model.ChatReadResult{
		MessageIDs:  inserted,
		ReadAt:      readAt,
		UnreadCount: unreadCount,
	}, nil
}

func (s *ChatService) GetCompanyUnreadCount(companyID int64, userID int64) (int64, error) {
	return s.repo.GetCompanyChatUnreadCount(companyID, userID)
}

type ChatUploadFile struct {
	FileName string
	FileData []byte
}

func saveChatMediaFile(companyID int64, userID int64, index int, file ChatUploadFile) (model.ChatAttachmentCreate, error) {
	if len(file.FileData) == 0 {
		return model.ChatAttachmentCreate{}, ErrChatMediaInvalidType
	}

	contentType := http.DetectContentType(file.FileData)
	ext, mediaType, maxSize, ok := chatMediaMeta(contentType)
	if !ok {
		return model.ChatAttachmentCreate{}, ErrChatMediaInvalidType
	}
	if len(file.FileData) > maxSize {
		return model.ChatAttachmentCreate{}, ErrChatMediaTooLarge
	}

	uploadDir := chatMediaStorageDir()
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return model.ChatAttachmentCreate{}, err
	}

	randomPart, err := randomHex(8)
	if err != nil {
		return model.ChatAttachmentCreate{}, err
	}

	safeName := strings.TrimSuffix(filepath.Base(file.FileName), filepath.Ext(file.FileName))
	if safeName == "." || safeName == string(filepath.Separator) || safeName == "" {
		safeName = "media"
	}

	fileBase := fmt.Sprintf("company-%d-user-%d-%d-%d-%s-%s%s", companyID, userID, time.Now().Unix(), index, safeName, randomPart, ext)
	fullPath := filepath.Join(uploadDir, fileBase)
	if err := os.WriteFile(fullPath, file.FileData, 0o644); err != nil {
		return model.ChatAttachmentCreate{}, err
	}

	return model.ChatAttachmentCreate{
		FileName:  file.FileName,
		FileURL:   "/uploads/chat/" + fileBase,
		FileType:  contentType,
		FileSize:  int64(len(file.FileData)),
		MediaType: mediaType,
	}, nil
}

func removeChatMediaByURL(fileURL string) error {
	filePath, ok := chatMediaURLToPath(fileURL)
	if !ok {
		return nil
	}

	err := os.Remove(filePath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func chatMediaStorageDir() string {
	if dir := os.Getenv("CHAT_MEDIA_UPLOAD_DIR"); dir != "" {
		return dir
	}
	return filepath.Join("uploads", "chat")
}

func chatMediaURLToPath(fileURL string) (string, bool) {
	const prefix = "/uploads/chat/"
	if !strings.HasPrefix(fileURL, prefix) {
		return "", false
	}

	fileName := filepath.Base(strings.TrimPrefix(fileURL, prefix))
	if fileName == "." || fileName == string(filepath.Separator) || fileName == "" {
		return "", false
	}

	return filepath.Join(chatMediaStorageDir(), fileName), true
}

func chatMediaMeta(contentType string) (ext string, mediaType string, maxSize int, ok bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", "photo", maxChatImageSize, true
	case "image/png":
		return ".png", "photo", maxChatImageSize, true
	case "image/webp":
		return ".webp", "photo", maxChatImageSize, true
	case "image/gif":
		return ".gif", "photo", maxChatImageSize, true
	case "video/mp4":
		return ".mp4", "video", maxChatVideoSize, true
	case "video/webm":
		return ".webm", "video", maxChatVideoSize, true
	case "video/quicktime":
		return ".mov", "video", maxChatVideoSize, true
	default:
		return "", "", 0, false
	}
}
