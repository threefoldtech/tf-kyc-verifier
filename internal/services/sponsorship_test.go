package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/threefoldtech/tf-kyc-verifier/internal/config"
	"github.com/threefoldtech/tf-kyc-verifier/internal/errors"
	"github.com/threefoldtech/tf-kyc-verifier/internal/models"
	"github.com/threefoldtech/tf-kyc-verifier/internal/repository"
)

// MockVerificationRepository is a mock implementation of VerificationRepository
type MockVerificationRepository struct {
	mock.Mock
}

func (m *MockVerificationRepository) GetVerification(ctx context.Context, clientID string) (*models.Verification, error) {
	args := m.Called(ctx, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Verification), args.Error(1)
}

func (m *MockVerificationRepository) SaveVerification(ctx context.Context, verification *models.Verification) error {
	args := m.Called(ctx, verification)
	return args.Error(0)
}

func (m *MockVerificationRepository) UpdateExpirationStatus(ctx context.Context, clientID, scanRef string, expirationThreshold models.ExpirationThreshold) error {
	args := m.Called(ctx, clientID, scanRef, expirationThreshold)
	return args.Error(0)
}

// MockTokenRepository is a mock implementation of TokenRepository
type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) GetToken(ctx context.Context, clientID string) (*models.Token, error) {
	args := m.Called(ctx, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Token), args.Error(1)
}

func (m *MockTokenRepository) SaveToken(ctx context.Context, token *models.Token) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenRepository) DeleteToken(ctx context.Context, clientID, scanRef string) error {
	args := m.Called(ctx, clientID, scanRef)
	return args.Error(0)
}

// MockSponsorshipRepository is a mock implementation of SponsorshipRepository
type MockSponsorshipRepository struct {
	mock.Mock
}

func (m *MockSponsorshipRepository) Create(ctx context.Context, sponsorship *models.Sponsorship) error {
	args := m.Called(ctx, sponsorship)
	return args.Error(0)
}

func (m *MockSponsorshipRepository) GetBySponsor(ctx context.Context, sponsorClientID string, pagination repository.PaginationParams) ([]*models.Sponsorship, int64, error) {
	args := m.Called(ctx, sponsorClientID, pagination)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Sponsorship), args.Get(1).(int64), args.Error(2)
}

func (m *MockSponsorshipRepository) GetBySponsee(ctx context.Context, sponseeClientID string) (*models.Sponsorship, error) {
	args := m.Called(ctx, sponseeClientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Sponsorship), args.Error(1)
}

func (m *MockSponsorshipRepository) ListAll(ctx context.Context, pagination repository.PaginationParams) ([]*models.Sponsorship, int64, error) {
	args := m.Called(ctx, pagination)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Sponsorship), args.Get(1).(int64), args.Error(2)
}

// MockIdenfyClient is a mock implementation of IdenfyClient
type MockIdenfyClient struct {
	mock.Mock
}

func (m *MockIdenfyClient) CreateVerificationSession(ctx context.Context, clientID string) (models.Token, error) {
	args := m.Called(ctx, clientID)
	return args.Get(0).(models.Token), args.Error(1)
}

func (m *MockIdenfyClient) VerifyCallbackSignature(ctx context.Context, body []byte, signature string) error {
	args := m.Called(ctx, body, signature)
	return args.Error(0)
}

// MockSubstrateClient is a mock implementation of SubstrateClient
type MockSubstrateClient struct {
	mock.Mock
}

func (m *MockSubstrateClient) GetChainName() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockSubstrateClient) GetAccountBalance(address string) (uint64, error) {
	args := m.Called(address)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockSubstrateClient) GetAddressByTwinID(twinID uint32) (string, error) {
	args := m.Called(twinID)
	return args.String(0), args.Error(1)
}

func setupKYCService(t *testing.T) (*KYCService, *MockVerificationRepository, *MockTokenRepository, *MockSponsorshipRepository, *MockIdenfyClient, *MockSubstrateClient) {
	verificationRepo := new(MockVerificationRepository)
	tokenRepo := new(MockTokenRepository)
	sponsorshipRepo := new(MockSponsorshipRepository)
	idenfyClient := new(MockIdenfyClient)
	substrateClient := new(MockSubstrateClient)

	cfg := &config.Config{
		Verification: config.Verification{
			MinBalanceToVerifyAccount: 100,
			AlwaysVerifiedIDs:         []string{},
			AlwaysVerifiedIDsOnly:     false,
		},
		Idenfy: config.Idenfy{
			Namespace: "test",
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	substrateClient.On("GetChainName").Return("Threefold Testnet", nil).Maybe()

	service, err := NewKYCService(verificationRepo, tokenRepo, sponsorshipRepo, idenfyClient, substrateClient, cfg, logger)
	assert.NoError(t, err)

	return service, verificationRepo, tokenRepo, sponsorshipRepo, idenfyClient, substrateClient
}

func TestCreateSponsorship(t *testing.T) {
	ctx := context.Background()
	service, verificationRepo, _, sponsorshipRepo, _, _ := setupKYCService(t)

	sponsorClientID := "5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY"
	sponseeClientID := "5FHneW46xGXgs5mUapPM8ikVdJQvBLgVzW9ocmZZUadQbqh2"

	t.Run("successful sponsorship creation", func(t *testing.T) {
		// Reset mocks
		verificationRepo.Calls = []mock.Call{}
		sponsorshipRepo.Calls = []mock.Call{}

		// Mock sponsor as verified
		approved := models.OverallApproved
		verificationRepo.On("GetVerification", ctx, sponsorClientID).Return(&models.Verification{
			Status: models.Status{
				Overall: &approved,
			},
			Final: new(bool),
		}, nil).Once()

		// Mock sponsee not yet sponsored
		sponsorshipRepo.On("GetBySponsee", ctx, sponseeClientID).Return(nil, nil).Once()

		// Mock successful sponsorship creation
		sponsorshipRepo.On("Create", ctx, mock.AnythingOfType("*models.Sponsorship")).Return(nil).Once()

		sponsorship, err := service.CreateSponsorship(ctx, sponsorClientID, sponseeClientID)

		assert.NoError(t, err)
		assert.NotNil(t, sponsorship)
		assert.Equal(t, sponsorClientID, sponsorship.SponsorClientID)
		assert.Equal(t, sponseeClientID, sponsorship.SponseeClientID)
		verificationRepo.AssertExpectations(t)
		sponsorshipRepo.AssertExpectations(t)
	})

	t.Run("sponsor not KYC-verified", func(t *testing.T) {
		// Reset mocks
		verificationRepo.Calls = []mock.Call{}
		sponsorshipRepo.Calls = []mock.Call{}

		// Mock sponsor as not verified
		rejected := models.OverallDenied
		verificationRepo.On("GetVerification", ctx, sponsorClientID).Return(&models.Verification{
			Status: models.Status{
				Overall: &rejected,
			},
			Final: new(bool),
		}, nil).Once()

		sponsorship, err := service.CreateSponsorship(ctx, sponsorClientID, sponseeClientID)

		assert.Error(t, err)
		assert.Nil(t, sponsorship)
		assert.IsType(t, &errors.ServiceError{}, err)
		assert.Equal(t, errors.ErrorTypeAuthorization, err.(*errors.ServiceError).Type)
		verificationRepo.AssertExpectations(t)
		sponsorshipRepo.AssertNotCalled(t, "GetBySponsee", ctx, sponseeClientID)
		sponsorshipRepo.AssertNotCalled(t, "Create", ctx, mock.AnythingOfType("*models.Sponsorship"))
	})

	t.Run("sponsee already sponsored", func(t *testing.T) {
		// Reset mocks
		verificationRepo.Calls = []mock.Call{}
		sponsorshipRepo.Calls = []mock.Call{}

		// Mock sponsor as verified
		approved := models.OverallApproved
		verificationRepo.On("GetVerification", ctx, sponsorClientID).Return(&models.Verification{
			Status: models.Status{
				Overall: &approved,
			},
			Final: new(bool),
		}, nil).Once()

		// Mock sponsee already sponsored
		sponsorshipRepo.On("GetBySponsee", ctx, sponseeClientID).Return(&models.Sponsorship{}, nil).Once()

		sponsorship, err := service.CreateSponsorship(ctx, sponsorClientID, sponseeClientID)

		assert.Error(t, err)
		assert.Nil(t, sponsorship)
		assert.IsType(t, &errors.ServiceError{}, err)
		assert.Equal(t, errors.ErrorTypeConflict, err.(*errors.ServiceError).Type)
		verificationRepo.AssertExpectations(t)
		sponsorshipRepo.AssertExpectations(t)
		sponsorshipRepo.AssertNotCalled(t, "Create", ctx, mock.AnythingOfType("*models.Sponsorship"))
	})

	t.Run("error getting sponsor verification", func(t *testing.T) {
		// Reset mocks
		verificationRepo.Calls = []mock.Call{}
		sponsorshipRepo.Calls = []mock.Call{}

		verificationRepo.On("GetVerification", ctx, sponsorClientID).Return(nil, fmt.Errorf("db error")).Once()

		sponsorship, err := service.CreateSponsorship(ctx, sponsorClientID, sponseeClientID)

		assert.Error(t, err)
		assert.Nil(t, sponsorship)
		assert.Contains(t, err.Error(), "checking sponsor verification status")
		verificationRepo.AssertExpectations(t)
		sponsorshipRepo.AssertNotCalled(t, "GetBySponsee")
		sponsorshipRepo.AssertNotCalled(t, "Create")
	})

	t.Run("error checking existing sponsorship", func(t *testing.T) {
		// Reset mocks
		verificationRepo.Calls = []mock.Call{}
		sponsorshipRepo.Calls = []mock.Call{}

		approved := models.OverallApproved
		verificationRepo.On("GetVerification", ctx, sponsorClientID).Return(&models.Verification{
			Status: models.Status{
				Overall: &approved,
			},
			Final: new(bool),
		}, nil).Once()
		sponsorshipRepo.On("GetBySponsee", ctx, sponseeClientID).Return(nil, fmt.Errorf("db error")).Once()

		sponsorship, err := service.CreateSponsorship(ctx, sponsorClientID, sponseeClientID)

		assert.Error(t, err)
		assert.Nil(t, sponsorship)
		assert.Contains(t, err.Error(), "checking existing sponsorship")
		verificationRepo.AssertExpectations(t)
		sponsorshipRepo.AssertExpectations(t)
		sponsorshipRepo.AssertNotCalled(t, "Create")
	})

	t.Run("error creating sponsorship in repository", func(t *testing.T) {
		// Reset mocks
		verificationRepo.Calls = []mock.Call{}
		sponsorshipRepo.Calls = []mock.Call{}

		approved := models.OverallApproved
		verificationRepo.On("GetVerification", ctx, sponsorClientID).Return(&models.Verification{
			Status: models.Status{
				Overall: &approved,
			},
			Final: new(bool),
		}, nil).Once()
		sponsorshipRepo.On("GetBySponsee", ctx, sponseeClientID).Return(nil, nil).Once()
		sponsorshipRepo.On("Create", ctx, mock.AnythingOfType("*models.Sponsorship")).Return(fmt.Errorf("db error")).Once()

		sponsorship, err := service.CreateSponsorship(ctx, sponsorClientID, sponseeClientID)

		assert.Error(t, err)
		assert.Nil(t, sponsorship)
		assert.Contains(t, err.Error(), "creating sponsorship")
		verificationRepo.AssertExpectations(t)
		sponsorshipRepo.AssertExpectations(t)
	})
}

func TestGetSponsorshipsBySponsor(t *testing.T) {
	ctx := context.Background()
	service, _, _, sponsorshipRepo, _, substrateClient := setupKYCService(t)

	sponsorTwinID := uint32(123)
	sponsorAddress := "5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY"
	limit := int64(10)
	offset := int64(0)

	t.Run("successful retrieval", func(t *testing.T) {
		// Reset mocks
		sponsorshipRepo.Calls = []mock.Call{}
		substrateClient.Calls = []mock.Call{}

		expectedSponsorships := []*models.Sponsorship{
			{ID: primitive.NewObjectID(), SponsorClientID: sponsorAddress, SponseeClientID: "sponsee1", CreatedAt: time.Now()},
		}
		expectedTotal := int64(1)

		substrateClient.On("GetAddressByTwinID", sponsorTwinID).Return(sponsorAddress, nil).Once()
		sponsorshipRepo.On("GetBySponsor", ctx, sponsorAddress, repository.PaginationParams{Limit: limit, Offset: offset}).Return(expectedSponsorships, expectedTotal, nil).Once()

		sponsorships, total, err := service.GetSponsorshipsBySponsor(ctx, sponsorTwinID, limit, offset)

		assert.NoError(t, err)
		assert.Equal(t, expectedSponsorships, sponsorships)
		assert.Equal(t, expectedTotal, total)
		substrateClient.AssertExpectations(t)
		sponsorshipRepo.AssertExpectations(t)
	})

	t.Run("error getting address from twin ID", func(t *testing.T) {
		// Reset mocks
		sponsorshipRepo.Calls = []mock.Call{}
		substrateClient.Calls = []mock.Call{}

		substrateClient.On("GetAddressByTwinID", sponsorTwinID).Return("", fmt.Errorf("substrate error")).Once()

		sponsorships, total, err := service.GetSponsorshipsBySponsor(ctx, sponsorTwinID, limit, offset)

		assert.Error(t, err)
		assert.Nil(t, sponsorships)
		assert.Zero(t, total)
		assert.Contains(t, err.Error(), "getting address from twinID")
		substrateClient.AssertExpectations(t)
		sponsorshipRepo.AssertNotCalled(t, "GetBySponsor")
	})

	t.Run("error getting sponsorships from repository", func(t *testing.T) {
		// Reset mocks
		sponsorshipRepo.Calls = []mock.Call{}
		substrateClient.Calls = []mock.Call{}

		substrateClient.On("GetAddressByTwinID", sponsorTwinID).Return(sponsorAddress, nil).Once()
		sponsorshipRepo.On("GetBySponsor", ctx, sponsorAddress, repository.PaginationParams{Limit: limit, Offset: offset}).Return(nil, int64(0), fmt.Errorf("db error")).Once()

		sponsorships, total, err := service.GetSponsorshipsBySponsor(ctx, sponsorTwinID, limit, offset)

		assert.Error(t, err)
		assert.Nil(t, sponsorships)
		assert.Zero(t, total)
		assert.Contains(t, err.Error(), "db error")
		substrateClient.AssertExpectations(t)
		sponsorshipRepo.AssertExpectations(t)
	})
}

func TestGetSponsorshipBySponsee(t *testing.T) {
	ctx := context.Background()
	service, _, _, sponsorshipRepo, _, substrateClient := setupKYCService(t)

	sponseeTwinID := uint32(456)
	sponseeAddress := "5FHneW46xGXgs5mUapPM8ikVdJQvBLgVzW9ocmZZUadQbqh2"

	t.Run("successful retrieval", func(t *testing.T) {
		// Reset mocks
		sponsorshipRepo.Calls = []mock.Call{}
		substrateClient.Calls = []mock.Call{}

		expectedSponsorship := &models.Sponsorship{ID: primitive.NewObjectID(), SponsorClientID: "sponsor1", SponseeClientID: sponseeAddress, CreatedAt: time.Now()}

		substrateClient.On("GetAddressByTwinID", sponseeTwinID).Return(sponseeAddress, nil).Once()
		sponsorshipRepo.On("GetBySponsee", ctx, sponseeAddress).Return(expectedSponsorship, nil).Once()

		sponsorship, err := service.GetSponsorshipBySponsee(ctx, sponseeTwinID)

		assert.NoError(t, err)
		assert.Equal(t, expectedSponsorship, sponsorship)
		substrateClient.AssertExpectations(t)
		sponsorshipRepo.AssertExpectations(t)
	})

	t.Run("error getting address from twin ID", func(t *testing.T) {
		// Reset mocks
		sponsorshipRepo.Calls = []mock.Call{}
		substrateClient.Calls = []mock.Call{}

		substrateClient.On("GetAddressByTwinID", sponseeTwinID).Return("", fmt.Errorf("substrate error")).Once()

		sponsorship, err := service.GetSponsorshipBySponsee(ctx, sponseeTwinID)

		assert.Error(t, err)
		assert.Nil(t, sponsorship)
		assert.Contains(t, err.Error(), "getting address from twinID")
		substrateClient.AssertExpectations(t)
		sponsorshipRepo.AssertNotCalled(t, "GetBySponsee")
	})

	t.Run("error getting sponsorship from repository", func(t *testing.T) {
		// Reset mocks
		sponsorshipRepo.Calls = []mock.Call{}
		substrateClient.Calls = []mock.Call{}

		substrateClient.On("GetAddressByTwinID", sponseeTwinID).Return(sponseeAddress, nil).Once()
		sponsorshipRepo.On("GetBySponsee", ctx, sponseeAddress).Return(nil, fmt.Errorf("db error")).Once()

		sponsorship, err := service.GetSponsorshipBySponsee(ctx, sponseeTwinID)

		assert.Error(t, err)
		assert.Nil(t, sponsorship)
		assert.Contains(t, err.Error(), "db error")
		substrateClient.AssertExpectations(t)
		sponsorshipRepo.AssertExpectations(t)
	})
}

func TestListAllSponsorships(t *testing.T) {
	ctx := context.Background()
	service, _, _, sponsorshipRepo, _, _ := setupKYCService(t)

	limit := int64(10)
	offset := int64(0)

	t.Run("successful retrieval", func(t *testing.T) {
		// Reset mocks
		sponsorshipRepo.Calls = []mock.Call{}

		expectedSponsorships := []*models.Sponsorship{
			{ID: primitive.NewObjectID(), SponsorClientID: "sponsor1", SponseeClientID: "sponsee1", CreatedAt: time.Now()},
			{ID: primitive.NewObjectID(), SponsorClientID: "sponsor2", SponseeClientID: "sponsee2", CreatedAt: time.Now()},
		}
		expectedTotal := int64(2)

		sponsorshipRepo.On("ListAll", ctx, repository.PaginationParams{Limit: limit, Offset: offset}).Return(expectedSponsorships, expectedTotal, nil).Once()

		sponsorships, total, err := service.ListAllSponsorships(ctx, limit, offset)

		assert.NoError(t, err)
		assert.Equal(t, expectedSponsorships, sponsorships)
		assert.Equal(t, expectedTotal, total)
		sponsorshipRepo.AssertExpectations(t)
	})

	t.Run("error listing all sponsorships from repository", func(t *testing.T) {
		// Reset mocks
		sponsorshipRepo.Calls = []mock.Call{}

		sponsorshipRepo.On("ListAll", ctx, repository.PaginationParams{Limit: limit, Offset: offset}).Return(nil, int64(0), fmt.Errorf("db error")).Once()

		sponsorships, total, err := service.ListAllSponsorships(ctx, limit, offset)

		assert.Error(t, err)
		assert.Nil(t, sponsorships)
		assert.Zero(t, total)
		assert.Contains(t, err.Error(), "db error")
		sponsorshipRepo.AssertExpectations(t)
	})
}
