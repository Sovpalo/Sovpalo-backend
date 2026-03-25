package service

import (
	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
)

type Service struct {
	Authorization
	Company
	Event
	Availability
	Idea
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repos.Authorization),
		Company:       NewCompanyService(repos.Company),
		Event:         NewEventService(repos.Event),
		Availability:  NewAvailabilityService(repos.Availability),
		Idea:          NewIdeaService(repos.Idea),
	}
}

type Authorization interface {
	CreateUser(user model.User) (int, error)
	ParseToken(token string) (int, error)
	UserExists(email string) (bool, error)
	SendCodeToEmail(to string, code string) error
	GenerateCode() string
	GenerateToken(email, password string) (string, error)
}

type Company interface {
	CreateCompany(userID int64, name string, description *string) (int64, error)
	GetCompany(companyID int64, userID int64) (model.Company, error)
	ListCompanies(userID int64) ([]model.Company, error)
	UpdateCompany(companyID int64, userID int64, input model.CompanyUpdateInput) error
	DeleteCompany(companyID int64, userID int64) error

	InviteUser(companyID int64, invitedBy int64, username string) (model.CompanyInvitation, error)
	ListInvitations(userID int64) ([]model.CompanyInvitationView, error)
	AcceptInvitation(inviteID int64, userID int64) error
	DeclineInvitation(inviteID int64, userID int64) error
	ListCompanyMembers(companyID int64, userID int64) ([]model.CompanyMemberView, error)
	RemoveCompanyMember(companyID int64, ownerID int64, memberUserID int64) error
}

type Event interface {
	CreateEvent(userID int64, input model.EventCreateInput) (int64, error)
	GetEvent(eventID int64, userID int64) (model.Event, error)
	ListEvents(userID int64) ([]model.Event, error)
	ListCompanyEvents(companyID int64, userID int64) ([]model.Event, error)
	UpdateEvent(eventID int64, userID int64, input model.EventUpdateInput) error
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
	CreateCompanyIdea(companyID int64, userID int64, input model.IdeaCreateInput) (int64, error)
	ListCompanyIdeas(companyID int64, userID int64) ([]model.IdeaView, error)
	GetCompanyIdea(companyID int64, userID int64, ideaID int64) (model.IdeaView, error)
	LikeCompanyIdea(companyID int64, userID int64, ideaID int64) error
	UnlikeCompanyIdea(companyID int64, userID int64, ideaID int64) error
}
