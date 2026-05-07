package service

import (
	"context"
	"time"

	"github.com/Sovpalo/sovpalo-backend/internal/config"
	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
)

type Service struct {
	Authorization
	Company
	Event
	Availability
	Idea
	Chat
	Notification
}

func NewService(repos *repository.Repository, cfg config.Config) *Service {
	ideaGenerator := NewIdeaGenerator(cfg)
	pushSender := NewAPNSPushSender(cfg, repos.Notification)
	notificationService := NewNotificationService(repos.Notification, pushSender)

	return &Service{
		Authorization: NewAuthService(repos.Authorization),
		Company:       NewCompanyService(repos.Company),
		Event:         NewEventService(repos.Event, repos.Availability, notificationService),
		Availability:  NewAvailabilityService(repos.Availability),
		Idea:          NewIdeaService(repos.Idea, repos.Company, ideaGenerator),
		Chat:          NewChatService(repos.Chat, notificationService),
		Notification:  notificationService,
	}
}

type Authorization interface {
	CreateUser(user model.User) (int, error)
	ParseToken(token string) (int, error)
	UserExists(email string) (bool, error)
	UsernameExists(username string) (bool, error)
	GetProfile(userID int64) (model.UserProfile, error)
	UpdateAvatar(userID int64, fileName string, fileData []byte) (model.UserProfile, error)
	DeleteAvatar(userID int64) (model.UserProfile, error)
	DeleteUser(userID int64) error
	SendCodeToEmail(to string, code string) error
	GenerateCode() string
	GenerateToken(email, password string) (string, error)
	SignIn(input model.SignInInput) (string, error)
	SignInTelegram(input model.TelegramSignInInput) (string, error)
	StartRegistration(input model.SignUpInput) error
	VerifyRegistration(input model.SignUpVerifyInput) (string, error)
	ResendRegistrationCode(email string) error
	StartPasswordReset(email string) error
	VerifyPasswordReset(input model.ResetPasswordVerifyInput) error
	ResendPasswordResetCode(email string) error
	PendingRegistrationTTL() time.Duration
}

type Company interface {
	CreateCompany(userID int64, name string, description *string, avatarURL *string) (int64, error)
	GetCompany(companyID int64, userID int64) (model.Company, error)
	ListCompanies(userID int64) ([]model.Company, error)
	UpdateCompany(companyID int64, userID int64, input model.CompanyUpdateInput, avatarFileName string, avatarFileData []byte) error
	DeleteCompany(companyID int64, userID int64) error
	LeaveCompany(companyID int64, userID int64, newOwnerID *int64) error

	InviteUser(companyID int64, invitedBy int64, username string) (model.CompanyInvitation, error)
	ListInvitations(userID int64) ([]model.CompanyInvitationView, error)
	AcceptInvitation(inviteID int64, userID int64) error
	DeclineInvitation(inviteID int64, userID int64) error
	ListCompanyMembers(companyID int64, userID int64) ([]model.CompanyMemberView, error)
	RemoveCompanyMember(companyID int64, ownerID int64, memberUserID int64) error
}

type Event interface {
	CreateEvent(userID int64, input model.EventCreateInput, photoFileName string, photoFileData []byte) (int64, error)
	GetEvent(eventID int64, userID int64) (model.Event, error)
	ListEvents(userID int64) ([]model.Event, error)
	ListCompanyEvents(companyID int64, userID int64) ([]model.Event, error)
	GetCompanyEventFeatures(companyID int64, eventID int64, userID int64, now time.Time) (model.EventFeatures, error)
	UpdateEvent(eventID int64, userID int64, input model.EventUpdateInput, photoFileName string, photoFileData []byte) error
	DeleteEvent(eventID int64, userID int64) error
	SetCompanyEventAttendance(companyID int64, eventID int64, userID int64, status string) error
	ListCompanyEventAttendance(companyID int64, eventID int64, userID int64) ([]model.EventAttendanceView, error)
}

type Availability interface {
	CreateAvailability(companyID int64, userID int64, input model.AvailabilityCreateInput) (int64, error)
	ListAvailability(companyID int64, userID int64) ([]model.UserAvailability, error)
	ListCompanyAvailability(companyID int64, userID int64) ([]model.UserAvailability, error)
	UpdateAvailability(companyID int64, userID int64, availabilityID int64, input model.AvailabilityCreateInput) error
	DeleteAvailability(companyID int64, userID int64, availabilityID int64) error
	GetAvailabilityIntersections(companyID int64, userID int64, input model.AvailabilityRangeInput) ([]model.AvailabilityIntersection, error)
}

type Idea interface {
	CreateCompanyIdea(companyID int64, userID int64, input model.IdeaCreateInput, photoFileName string, photoFileData []byte) (int64, error)
	GenerateCompanyIdeas(ctx context.Context, companyID int64, userID int64, input model.IdeaGenerateInput) ([]model.GeneratedIdeaDraft, error)
	ListCompanyIdeas(companyID int64, userID int64) ([]model.IdeaView, error)
	GetCompanyIdea(companyID int64, userID int64, ideaID int64) (model.IdeaView, error)
	UpdateCompanyIdea(companyID int64, userID int64, ideaID int64, input model.IdeaUpdateInput, photoFileName string, photoFileData []byte) error
	LikeCompanyIdea(companyID int64, userID int64, ideaID int64) error
	UnlikeCompanyIdea(companyID int64, userID int64, ideaID int64) error
}

type Chat interface {
	ListCompanyMessages(companyID int64, userID int64, beforeMessageID int64, limit int) ([]model.ChatMessageView, error)
	CreateCompanyTextMessage(companyID int64, userID int64, input model.ChatMessageCreateInput) (model.ChatMessageView, error)
	CreateCompanyMediaMessage(companyID int64, userID int64, files []ChatUploadFile) (model.ChatMessageView, error)
	DeleteCompanyMessage(companyID int64, messageID int64, userID int64) error
	MarkCompanyMessagesRead(companyID int64, userID int64, input model.ChatMarkReadInput) (model.ChatReadResult, error)
	GetCompanyUnreadCount(companyID int64, userID int64) (int64, error)
}

type Notification interface {
	RegisterPushToken(userID int64, input model.PushTokenRegisterInput) error
	DeletePushToken(userID int64, token string) error
	Dispatch(notification model.PushNotification)
}
