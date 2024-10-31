package services

import (
	"context"
	"math/big"
	"slices"
	"strings"
	"time"

	"example.com/tfgrid-kyc-service/internal/clients/idenfy"
	"example.com/tfgrid-kyc-service/internal/clients/substrate"
	"example.com/tfgrid-kyc-service/internal/configs"
	"example.com/tfgrid-kyc-service/internal/errors"
	"example.com/tfgrid-kyc-service/internal/logger"
	"example.com/tfgrid-kyc-service/internal/models"
	"example.com/tfgrid-kyc-service/internal/repository"
	"go.uber.org/zap"
)

type kycService struct {
	verificationRepo repository.VerificationRepository
	tokenRepo        repository.TokenRepository
	idenfy           *idenfy.Idenfy
	substrate        *substrate.Substrate
	config           *configs.Verification
	logger           *logger.Logger
	IdenfySuffix     string
}

func NewKYCService(verificationRepo repository.VerificationRepository, tokenRepo repository.TokenRepository, idenfy *idenfy.Idenfy, substrateClient *substrate.Substrate, config *configs.Config, logger *logger.Logger) KYCService {
	chainName, err := substrateClient.GetChainName()
	if err != nil {
		panic(errors.NewInternalError("error getting chain name", err))
	}
	chainNameParts := strings.Split(chainName, " ")
	chainNetworkName := strings.ToLower(chainNameParts[len(chainNameParts)-1])
	if config.Idenfy.Namespace != "" {
		chainNetworkName = config.Idenfy.Namespace + ":" + chainNetworkName
	}
	return &kycService{verificationRepo: verificationRepo, tokenRepo: tokenRepo, idenfy: idenfy, substrate: substrateClient, config: &config.Verification, logger: logger, IdenfySuffix: chainNetworkName}
}

// ---------------------------------------------------------------------------------------------------------------------
// token related methods
// ---------------------------------------------------------------------------------------------------------------------

func (s *kycService) GetorCreateVerificationToken(ctx context.Context, clientID string) (*models.Token, bool, error) {
	isVerified, err := s.IsUserVerified(ctx, clientID)
	if err != nil {
		s.logger.Error("Error checking if user is verified", zap.String("clientID", clientID), zap.Error(err))
		return nil, false, errors.NewInternalError("error getting verification status from database", err) // db error
	}
	if isVerified {
		return nil, false, errors.NewConflictError("user already verified", nil) // TODO: implement a custom error that can be converted in the handler to a 4xx such 409 status code
	}
	token, err_ := s.tokenRepo.GetToken(ctx, clientID)
	if err_ != nil {
		s.logger.Error("Error getting token from database", zap.String("clientID", clientID), zap.Error(err_))
		return nil, false, errors.NewInternalError("error getting token from database", err_) // db error
	}
	// check if token is found and not expired
	if token != nil {
		duration := time.Since(token.CreatedAt)
		if duration < time.Duration(token.ExpiryTime)*time.Second {
			remainingTime := time.Duration(token.ExpiryTime)*time.Second - duration
			token.ExpiryTime = int(remainingTime.Seconds())
			return token, false, nil
		}
	}

	// check if user account balance satisfies the minimum required balance, return an error if not
	hasRequiredBalance, err_ := s.AccountHasRequiredBalance(ctx, clientID)
	if err_ != nil {
		s.logger.Error("Error checking if user account has required balance", zap.String("clientID", clientID), zap.Error(err_))
		return nil, false, errors.NewExternalError("error checking if user account has required balance", err_)
	}
	if !hasRequiredBalance {
		return nil, false, errors.NewNotSufficientBalanceError("account does not have the required balance", nil)
	}
	// prefix clientID with tfchain network prefix
	uniqueClientID := clientID + ":" + s.IdenfySuffix
	newToken, err_ := s.idenfy.CreateVerificationSession(ctx, uniqueClientID)
	if err_ != nil {
		s.logger.Error("Error creating iDenfy verification session", zap.String("clientID", clientID), zap.String("uniqueClientID", uniqueClientID), zap.Error(err_))
		return nil, false, errors.NewExternalError("error creating iDenfy verification session", err_)
	}
	// save the token with the original clientID
	newToken.ClientID = clientID
	err_ = s.tokenRepo.SaveToken(ctx, &newToken)
	if err_ != nil {
		s.logger.Error("Error saving verification token to database", zap.String("clientID", clientID), zap.Error(err))
	}

	return &newToken, true, nil
}

func (s *kycService) DeleteToken(ctx context.Context, clientID string, scanRef string) error {

	err := s.tokenRepo.DeleteToken(ctx, clientID, scanRef)
	if err != nil {
		s.logger.Error("Error deleting verification token from database", zap.String("clientID", clientID), zap.String("scanRef", scanRef), zap.Error(err))
		return errors.NewInternalError("error deleting verification token from database", err)
	}
	return nil
}

func (s *kycService) AccountHasRequiredBalance(ctx context.Context, address string) (bool, error) {
	if s.config.MinBalanceToVerifyAccount == 0 {
		s.logger.Warn("Minimum balance to verify account is 0 which is not recommended", zap.String("address", address))
		return true, nil
	}
	balance, err := s.substrate.GetAccountBalance(address)
	if err != nil {
		s.logger.Error("Error getting account balance", zap.String("address", address), zap.Error(err))
		return false, errors.NewExternalError("error getting account balance", err)
	}
	return balance.Cmp(big.NewInt(int64(s.config.MinBalanceToVerifyAccount))) >= 0, nil
}

// ---------------------------------------------------------------------------------------------------------------------
// verification related methods
// ---------------------------------------------------------------------------------------------------------------------

func (s *kycService) GetVerificationData(ctx context.Context, clientID string) (*models.Verification, error) {
	verification, err := s.verificationRepo.GetVerification(ctx, clientID)
	if err != nil {
		s.logger.Error("Error getting verification from database", zap.String("clientID", clientID), zap.Error(err))
		return nil, errors.NewInternalError("error getting verification from database", err)
	}
	return verification, nil
}

func (s *kycService) GetVerificationStatus(ctx context.Context, clientID string) (*models.VerificationOutcome, error) {
	// check first if the clientID is in alwaysVerifiedAddresses
	if s.config.AlwaysVerifiedIDs != nil && slices.Contains(s.config.AlwaysVerifiedIDs, clientID) {
		final := true
		return &models.VerificationOutcome{
			Final:     &final,
			ClientID:  clientID,
			IdenfyRef: "",
			Outcome:   models.OutcomeApproved,
		}, nil
	}
	verification, err := s.verificationRepo.GetVerification(ctx, clientID)
	if err != nil {
		s.logger.Error("Error getting verification from database", zap.String("clientID", clientID), zap.Error(err))
		return nil, errors.NewInternalError("error getting verification from database", err)
	}
	var outcome models.Outcome
	if verification != nil {
		if verification.Status.Overall != nil && *verification.Status.Overall == models.OverallApproved || (s.config.SuspiciousVerificationOutcome == "APPROVED" && *verification.Status.Overall == models.OverallSuspected) {
			outcome = models.OutcomeApproved
		} else {
			outcome = models.OutcomeRejected
		}
	} else {
		return nil, nil
	}
	return &models.VerificationOutcome{
		Final:     verification.Final,
		ClientID:  clientID,
		IdenfyRef: verification.IdenfyRef,
		Outcome:   outcome,
	}, nil
}

func (s *kycService) GetVerificationStatusByTwinID(ctx context.Context, twinID string) (*models.VerificationOutcome, error) {
	// get the address from the twinID
	address, err := s.substrate.GetAddressByTwinID(twinID)
	if err != nil {
		s.logger.Error("Error getting address from twinID", zap.String("twinID", twinID), zap.Error(err))
		return nil, errors.NewExternalError("error looking up twinID address from TFChain", err)
	}
	return s.GetVerificationStatus(ctx, address)
}

func (s *kycService) ProcessVerificationResult(ctx context.Context, body []byte, sigHeader string, result models.Verification) error {
	err := s.idenfy.VerifyCallbackSignature(ctx, body, sigHeader)
	if err != nil {
		s.logger.Error("Error verifying callback signature", zap.String("sigHeader", sigHeader), zap.Error(err))
		return errors.NewAuthorizationError("error verifying callback signature", err)
	}
	// delete the token with the same clientID and same scanRef
	result.ClientID = strings.Split(result.ClientID, ":")[0] // TODO: should we check if it have correct suffix? callback misconfiguration maybe?
	err = s.tokenRepo.DeleteToken(ctx, result.ClientID, result.IdenfyRef)
	if err != nil {
		s.logger.Warn("Error deleting verification token from database", zap.String("clientID", result.ClientID), zap.String("scanRef", result.IdenfyRef), zap.Error(err))
	}
	// if the verification status is EXPIRED, we don't need to save it
	if result.Status.Overall != nil && *result.Status.Overall != models.Overall("EXPIRED") {
		// remove idenfy suffix from clientID
		err = s.verificationRepo.SaveVerification(ctx, &result)
		if err != nil {
			s.logger.Error("Error saving verification to database", zap.String("clientID", result.ClientID), zap.String("scanRef", result.IdenfyRef), zap.Error(err))
			return errors.NewInternalError("error saving verification to database", err)
		}
	}
	s.logger.Debug("Verification result processed successfully", zap.Any("result", result))
	return nil
}

func (s *kycService) ProcessDocExpirationNotification(ctx context.Context, clientID string) error {
	return nil
}

func (s *kycService) IsUserVerified(ctx context.Context, clientID string) (bool, error) {
	verification, err := s.verificationRepo.GetVerification(ctx, clientID)
	if err != nil {
		s.logger.Error("Error getting verification from database", zap.String("clientID", clientID), zap.Error(err))
		return false, errors.NewInternalError("error getting verification from database", err)
	}
	if verification == nil {
		return false, nil
	}
	return verification.Status.Overall != nil && (*verification.Status.Overall == models.OverallApproved || (s.config.SuspiciousVerificationOutcome == "APPROVED" && *verification.Status.Overall == models.OverallSuspected)), nil
}
