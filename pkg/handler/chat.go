package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	chatMultipartMaxMemory = (100 << 20) + 1024
	chatUploadReadLimit    = (100 << 20) + 1
)

var chatUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) listCompanyChatMessages(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid company id")
		return
	}

	beforeID, err := parseInt64Query(c, "before_id")
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid before_id")
		return
	}

	limit := 50
	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = parsed
	}

	messages, err := h.services.Chat.ListCompanyMessages(companyID, int64(userID), beforeID, limit)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h *Handler) createCompanyChatMessage(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid company id")
		return
	}

	contentType := c.GetHeader("Content-Type")
	var message model.ChatMessageView
	switch {
	case strings.HasPrefix(contentType, "multipart/form-data"):
		files, err := readChatMultipartFiles(c, "media")
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		message, err = h.services.Chat.CreateCompanyMediaMessage(companyID, int64(userID), files)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
	default:
		var input model.ChatMessageCreateInput
		if err := c.BindJSON(&input); err != nil {
			newErrorResponse(c, http.StatusBadRequest, "invalid input body")
			return
		}
		message, err = h.services.Chat.CreateCompanyTextMessage(companyID, int64(userID), input)
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
	}

	h.chatHub.BroadcastToCompany(companyID, model.ChatRealtimeEvent{
		Type:      "message_created",
		CompanyID: companyID,
		Message:   &message,
	})

	c.JSON(http.StatusOK, message)
}

func (h *Handler) deleteCompanyChatMessage(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid company id")
		return
	}

	messageID, err := strconv.ParseInt(c.Param("message_id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid message id")
		return
	}

	if err := h.services.Chat.DeleteCompanyMessage(companyID, messageID, int64(userID)); err != nil {
		switch {
		case errors.Is(err, service.ErrChatMessageNotFound):
			newErrorResponse(c, http.StatusNotFound, err.Error())
		default:
			newErrorResponse(c, http.StatusBadRequest, err.Error())
		}
		return
	}

	h.chatHub.BroadcastToCompany(companyID, model.ChatRealtimeEvent{
		Type:      "message_deleted",
		CompanyID: companyID,
		MessageID: &messageID,
	})

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) markCompanyChatMessagesRead(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid company id")
		return
	}

	var input model.ChatMarkReadInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	result, err := h.services.Chat.MarkCompanyMessagesRead(companyID, int64(userID), input)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	currentUserID := int64(userID)
	unreadCount := result.UnreadCount
	h.chatHub.BroadcastToCompany(companyID, model.ChatRealtimeEvent{
		Type:        "messages_read",
		CompanyID:   companyID,
		UserID:      &currentUserID,
		MessageIDs:  result.MessageIDs,
		UnreadCount: &unreadCount,
		ReadAt:      &result.ReadAt,
	})

	c.JSON(http.StatusOK, result)
}

func (h *Handler) getCompanyChatUnreadCount(c *gin.Context) {
	userID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid company id")
		return
	}

	count, err := h.services.Chat.GetCompanyUnreadCount(companyID, int64(userID))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

func (h *Handler) companyChatWebSocket(c *gin.Context) {
	companyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid company id")
		return
	}

	token := c.Query("token")
	if token == "" {
		header := c.GetHeader(authorizationHeader)
		if strings.HasPrefix(header, "Bearer ") {
			token = strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		}
	}
	if token == "" {
		newErrorResponse(c, http.StatusUnauthorized, "No authorization header")
		return
	}

	userID, err := h.services.Authorization.ParseToken(token)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	if _, err := h.services.Chat.GetCompanyUnreadCount(companyID, int64(userID)); err != nil {
		newErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}

	conn, err := chatUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &chatClient{
		hub:       h.chatHub,
		conn:      conn,
		send:      make(chan []byte, 32),
		companyID: companyID,
	}
	h.chatHub.Register(client)

	go client.writePump()
	client.readPump()
}

func readChatMultipartFiles(c *gin.Context, field string) ([]service.ChatUploadFile, error) {
	if err := c.Request.ParseMultipartForm(chatMultipartMaxMemory); err != nil {
		return nil, errors.New("invalid multipart form")
	}
	if text := strings.TrimSpace(c.PostForm("text")); text != "" {
		return nil, service.ErrChatMessageMixedContent
	}

	fileHeaders := c.Request.MultipartForm.File[field]
	if len(fileHeaders) == 0 {
		return nil, service.ErrChatMessageContentRequired
	}

	files := make([]service.ChatUploadFile, 0, len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			return nil, errors.New("failed to open media file")
		}

		fileData, err := io.ReadAll(io.LimitReader(file, chatUploadReadLimit))
		_ = file.Close()
		if err != nil {
			return nil, errors.New("failed to read media file")
		}

		files = append(files, service.ChatUploadFile{
			FileName: fileHeader.Filename,
			FileData: fileData,
		})
	}

	return files, nil
}

func parseInt64Query(c *gin.Context, key string) (int64, error) {
	value := c.Query(key)
	if value == "" {
		return 0, nil
	}
	return strconv.ParseInt(value, 10, 64)
}
