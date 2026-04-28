package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
)

type ideaRepoStub struct{}

func (ideaRepoStub) CreateCompanyIdea(companyID int64, userID int64, input model.IdeaCreateInput) (int64, error) {
	return 0, nil
}

func (ideaRepoStub) ListCompanyIdeas(companyID int64, userID int64) ([]model.IdeaView, error) {
	return nil, nil
}

func (ideaRepoStub) GetCompanyIdea(companyID int64, userID int64, ideaID int64) (model.IdeaView, error) {
	return model.IdeaView{}, nil
}

func (ideaRepoStub) UpdateCompanyIdea(companyID int64, userID int64, ideaID int64, input model.IdeaUpdateInput) error {
	return nil
}

func (ideaRepoStub) LikeCompanyIdea(companyID int64, userID int64, ideaID int64) error {
	return nil
}

func (ideaRepoStub) UnlikeCompanyIdea(companyID int64, userID int64, ideaID int64) error {
	return nil
}

type companyRepoStub struct {
	company model.Company
	err     error
}

func (s companyRepoStub) CreateCompany(company model.Company) (int64, error) {
	return 0, nil
}

func (s companyRepoStub) GetCompany(companyID int64, userID int64) (model.Company, error) {
	if s.err != nil {
		return model.Company{}, s.err
	}
	return s.company, nil
}

func (s companyRepoStub) ListCompanies(userID int64) ([]model.Company, error) {
	return nil, nil
}

func (s companyRepoStub) UpdateCompany(companyID int64, userID int64, input model.CompanyUpdateInput) error {
	return nil
}

func (s companyRepoStub) DeleteCompany(companyID int64, userID int64) error {
	return nil
}

func (s companyRepoStub) LeaveCompany(companyID int64, userID int64, newOwnerID *int64) error {
	return nil
}

func (s companyRepoStub) CreateInvitation(companyID int64, invitedBy int64, username string) (model.CompanyInvitation, error) {
	return model.CompanyInvitation{}, nil
}

func (s companyRepoStub) ListInvitations(userID int64) ([]model.CompanyInvitationView, error) {
	return nil, nil
}

func (s companyRepoStub) AcceptInvitation(inviteID int64, userID int64) error {
	return nil
}

func (s companyRepoStub) DeclineInvitation(inviteID int64, userID int64) error {
	return nil
}

func (s companyRepoStub) ListCompanyMembers(companyID int64, userID int64) ([]model.CompanyMemberView, error) {
	return nil, nil
}

func (s companyRepoStub) RemoveCompanyMember(companyID int64, ownerID int64, memberUserID int64) error {
	return nil
}

type ideaGeneratorStub struct {
	items []GeneratedIdeaContent
	err   error
}

func (s ideaGeneratorStub) GenerateIdeas(ctx context.Context, req IdeaGenerationRequest) ([]GeneratedIdeaContent, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.items, nil
}

func TestIdeaServiceGenerateCompanyIdeasReturnsDrafts(t *testing.T) {
	svc := NewIdeaService(
		ideaRepoStub{},
		companyRepoStub{
			company: model.Company{
				ID:        1,
				Name:      "Sovpalo",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		ideaGeneratorStub{
			items: []GeneratedIdeaContent{
				{Title: "Idea 1", Description: "Description 1"},
				{Title: "Idea 2", Description: "Description 2"},
			},
		},
	)

	items, err := svc.GenerateCompanyIdeas(context.Background(), 1, 10, model.IdeaGenerateInput{
		Topic: "Новые мероприятия для команды",
		Count: 2,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Source != yandexIdeaSourceName {
		t.Fatalf("unexpected source: %s", items[0].Source)
	}
	if items[0].LLMPrompt == "" {
		t.Fatal("expected llm prompt to be set")
	}
}

func TestIdeaServiceGenerateCompanyIdeasRejectsInvalidCount(t *testing.T) {
	svc := NewIdeaService(ideaRepoStub{}, companyRepoStub{}, ideaGeneratorStub{})

	_, err := svc.GenerateCompanyIdeas(context.Background(), 1, 10, model.IdeaGenerateInput{
		Topic: "Topic",
		Count: 6,
	})
	if !errors.Is(err, ErrIdeaCountOutOfRange) {
		t.Fatalf("expected %v, got %v", ErrIdeaCountOutOfRange, err)
	}
}

func TestIdeaServiceGenerateCompanyIdeasPropagatesMembershipError(t *testing.T) {
	expectedErr := errors.New("user is not a member of the company")
	svc := NewIdeaService(
		ideaRepoStub{},
		companyRepoStub{err: expectedErr},
		ideaGeneratorStub{},
	)

	_, err := svc.GenerateCompanyIdeas(context.Background(), 1, 10, model.IdeaGenerateInput{
		Topic: "Topic",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}
