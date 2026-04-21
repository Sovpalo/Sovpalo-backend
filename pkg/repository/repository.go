package repository

import (
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	Authorization
	Company
	Event
	Availability
	Idea
}

func NewRepository(pool *pgxpool.Pool, cache *redis.Client) *Repository {
	return &Repository{
		Authorization: NewAuthRepository(pool, cache),
		Company:       NewCompanyRepository(pool),
		Event:         NewEventRepository(pool),
		Availability:  NewAvailabilityRepository(pool),
		Idea:          NewIdeaRepository(pool),
	}
}

type Authorization interface {
	UserExists(email string) (bool, error)
	UsernameExists(username string) (bool, error)
	CreateUser(user model.User) (int, error)
	GetUser(email, password string) (model.User, error)
	GetUserByEmail(email string) (model.User, error)
	GetUserByID(userID int64) (model.User, error)
	UpdateUserAvatar(userID int64, avatarURL *string) error
	DeleteUser(userID int64) error
	UpdateUserPassword(email string, passwordHash string) error
	SavePendingAuthChallenge(challenge model.PendingAuthChallenge, ttl time.Duration) error
	GetPendingAuthChallenge(challengeType model.AuthChallengeType, email string) (model.PendingAuthChallenge, error)
	DeletePendingAuthChallenge(challengeType model.AuthChallengeType, email string) error
}

type Company interface {
	CreateCompany(company model.Company) (int64, error)
	GetCompany(companyID int64, userID int64) (model.Company, error)
	ListCompanies(userID int64) ([]model.Company, error)
	UpdateCompany(companyID int64, userID int64, input model.CompanyUpdateInput) error
	DeleteCompany(companyID int64, userID int64) error
	LeaveCompany(companyID int64, userID int64, newOwnerID *int64) error

	CreateInvitation(companyID int64, invitedBy int64, username string) (model.CompanyInvitation, error)
	ListInvitations(userID int64) ([]model.CompanyInvitationView, error)
	AcceptInvitation(inviteID int64, userID int64) error
	DeclineInvitation(inviteID int64, userID int64) error

	ListCompanyMembers(companyID int64, userID int64) ([]model.CompanyMemberView, error)
	RemoveCompanyMember(companyID int64, ownerID int64, memberUserID int64) error
}

type Event interface {
	CreateEvent(event model.Event) (int64, error)
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
	ListCompanyMemberIDs(companyID int64) ([]int64, error)
	ListAvailabilityInRange(companyID int64, start time.Time, end time.Time) ([]model.UserAvailability, error)
}

type Idea interface {
	CreateCompanyIdea(companyID int64, userID int64, input model.IdeaCreateInput) (int64, error)
	ListCompanyIdeas(companyID int64, userID int64) ([]model.IdeaView, error)
	GetCompanyIdea(companyID int64, userID int64, ideaID int64) (model.IdeaView, error)
	UpdateCompanyIdea(companyID int64, userID int64, ideaID int64, input model.IdeaUpdateInput) error
	LikeCompanyIdea(companyID int64, userID int64, ideaID int64) error
	UnlikeCompanyIdea(companyID int64, userID int64, ideaID int64) error
}
