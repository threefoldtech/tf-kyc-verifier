/*
Package services contains the services for the application.
This layer is responsible for handling the business logic.
*/
package services

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/threefoldtech/tf-kyc-verifier/internal/clients/idenfy"
	"github.com/threefoldtech/tf-kyc-verifier/internal/clients/substrate"
	"github.com/threefoldtech/tf-kyc-verifier/internal/config"
	"github.com/threefoldtech/tf-kyc-verifier/internal/errors"
	"github.com/threefoldtech/tf-kyc-verifier/internal/logger"
	"github.com/threefoldtech/tf-kyc-verifier/internal/models"
	"github.com/threefoldtech/tf-kyc-verifier/internal/repository"
)

const TFT_CONVERSION_FACTOR = 10000000

type KYCService struct {
	verificationRepo repository.VerificationRepository
	tokenRepo        repository.TokenRepository
	idenfy           idenfy.IdenfyClient
	substrate        substrate.SubstrateClient
	config           *config.Verification
	logger           logger.Logger
	IdenfySuffix     string
}

func NewKYCService(verificationRepo repository.VerificationRepository, tokenRepo repository.TokenRepository, idenfy idenfy.IdenfyClient, substrateClient substrate.SubstrateClient, config *config.Config, logger logger.Logger) (*KYCService, error) {
	idenfySuffix, err := GetIdenfySuffix(substrateClient, config)
	if err != nil {
		return nil, fmt.Errorf("getting idenfy suffix: %w", err)
	}
	return &KYCService{verificationRepo: verificationRepo, tokenRepo: tokenRepo, idenfy: idenfy, substrate: substrateClient, config: &config.Verification, logger: logger, IdenfySuffix: idenfySuffix}, nil
}

func GetIdenfySuffix(substrateClient substrate.SubstrateClient, config *config.Config) (string, error) {
	idenfySuffix, err := GetChainNetworkName(substrateClient)
	if err != nil {
		return "", fmt.Errorf("getting chain network name: %w", err)
	}
	if config.Idenfy.Namespace != "" {
		idenfySuffix = config.Idenfy.Namespace + ":" + idenfySuffix
	}
	return idenfySuffix, nil
}

func GetChainNetworkName(substrateClient substrate.SubstrateClient) (string, error) {
	chainName, err := substrateClient.GetChainName()
	if err != nil {
		return "", err
	}
	chainNameParts := strings.Split(chainName, " ")
	chainNetworkName := strings.ToLower(chainNameParts[len(chainNameParts)-1])
	return chainNetworkName, nil
}

// -----------------------------
// Token related methods
// -----------------------------
func (s *KYCService) GetOrCreateVerificationToken(ctx context.Context, clientID string) (*models.Token, bool, error) {
	isVerified, err := s.IsUserVerified(ctx, clientID)
	if err != nil {
		s.logger.Error("Error checking if user is verified", logger.Fields{"clientID": clientID, "error": err})
		return nil, false, errors.NewInternalError("getting verification status from database", err) // db error
	}
	if isVerified {
		return nil, false, errors.NewConflictError("user already verified", nil) // TODO: implement a custom error that can be converted in the handler to a 4xx such 409 status code
	}
	token, err_ := s.tokenRepo.GetToken(ctx, clientID)
	if err_ != nil {
		s.logger.Error("Error getting token from database", logger.Fields{"clientID": clientID, "error": err_})
		return nil, false, errors.NewInternalError("getting token from database", err_) // db error
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
		s.logger.Error("Error checking if user account has required balance", logger.Fields{"clientID": clientID, "error": err_})
		return nil, false, errors.NewExternalError("checking if user account has required balance", err_)
	}
	if !hasRequiredBalance {
		requiredBalance := s.config.MinBalanceToVerifyAccount / TFT_CONVERSION_FACTOR
		return nil, false, errors.NewNotSufficientBalanceError(fmt.Sprintf("account does not have the minimum required balance to verify (%d) TFT", requiredBalance), nil)
	}
	// prefix clientID with tfchain network prefix
	uniqueClientID := clientID + ":" + s.IdenfySuffix
	newToken, err_ := s.idenfy.CreateVerificationSession(ctx, uniqueClientID)
	if err_ != nil {
		s.logger.Error("Error creating iDenfy verification session", logger.Fields{"clientID": clientID, "uniqueClientID": uniqueClientID, "error": err_})
		return nil, false, errors.NewExternalError("creating iDenfy verification session", err_)
	}
	// save the token with the original clientID
	newToken.ClientID = clientID
	err_ = s.tokenRepo.SaveToken(ctx, &newToken)
	if err_ != nil {
		s.logger.Error("Error saving verification token to database", logger.Fields{"clientID": clientID, "error": err_})
	}

	return &newToken, true, nil
}

func (s *KYCService) DeleteToken(ctx context.Context, clientID string, scanRef string) error {

	err := s.tokenRepo.DeleteToken(ctx, clientID, scanRef)
	if err != nil {
		s.logger.Error("Error deleting verification token from database", logger.Fields{"clientID": clientID, "scanRef": scanRef, "error": err})
		return errors.NewInternalError("deleting verification token from database", err)
	}
	return nil
}

func (s *KYCService) AccountHasRequiredBalance(ctx context.Context, address string) (bool, error) {
	if s.config.MinBalanceToVerifyAccount == 0 {
		s.logger.Warn("Minimum balance to verify account is 0 which is not recommended", logger.Fields{"address": address})
		return true, nil
	}
	balance, err := s.substrate.GetAccountBalance(address)
	if err != nil {
		s.logger.Error("Error getting account balance", logger.Fields{"address": address, "error": err})
		return false, errors.NewExternalError("getting account balance", err)
	}
	return balance >= s.config.MinBalanceToVerifyAccount, nil
}

// -----------------------------
// Verifications related methods
// -----------------------------
func (s *KYCService) GetVerificationData(ctx context.Context, clientID string) (*models.Verification, error) {
	verification, err := s.verificationRepo.GetVerification(ctx, clientID)
	if err != nil {
		s.logger.Error("Error getting verification from database", logger.Fields{"clientID": clientID, "error": err})
		return nil, errors.NewInternalError("getting verification from database", err)
	}
	return verification, nil
}

func (s *KYCService) GetVerificationStatus(ctx context.Context, clientID string) (*models.VerificationOutcome, error) {
	// check first if the clientID is in alwaysVerifiedAddresses
	if s.config.AlwaysVerifiedIDs != nil && slices.Contains(s.config.AlwaysVerifiedIDs, clientID) {
		final := true
		s.logger.Info("ClientID is in always verified addresses. skipping verification", logger.Fields{"clientID": clientID})
		return &models.VerificationOutcome{
			Final:     &final,
			ClientID:  clientID,
			IdenfyRef: "",
			Outcome:   models.OutcomeApproved,
		}, nil
	}
	verification, err := s.verificationRepo.GetVerification(ctx, clientID)
	if err != nil {
		s.logger.Error("Error getting verification from database", logger.Fields{"clientID": clientID, "error": err})
		return nil, errors.NewInternalError("getting verification from database", err)
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

func (s *KYCService) GetVerificationStatusByTwinID(ctx context.Context, twinID string) (*models.VerificationOutcome, error) {
	// get the address from the twinID
	twinIDUint64, err := strconv.ParseUint(twinID, 10, 32)
	if err != nil {
		s.logger.Error("Error parsing twinID", logger.Fields{"twinID": twinID, "error": err})
		return nil, errors.NewInternalError("parsing twinID", err)
	}
	address, err := s.substrate.GetAddressByTwinID(uint32(twinIDUint64))
	if err != nil {
		s.logger.Error("Error getting address from twinID", logger.Fields{"twinID": twinID, "error": err})
		return nil, errors.NewExternalError("looking up twinID address from TFChain", err)
	}
	return s.GetVerificationStatus(ctx, address)
}

func (s *KYCService) ProcessVerificationResult(ctx context.Context, body []byte, sigHeader string, result models.Verification) error {
	err := s.idenfy.VerifyCallbackSignature(ctx, body, sigHeader)
	if err != nil {
		s.logger.Error("Error verifying callback signature", logger.Fields{"sigHeader": sigHeader, "error": err})
		return errors.NewAuthorizationError("verifying callback signature", err)
	}
	clientIDParts := strings.Split(result.ClientID, ":")
	if len(clientIDParts) < 2 {
		s.logger.Error("clientID have no network suffix", logger.Fields{"clientID": result.ClientID})
		return errors.NewInternalError("invalid clientID", nil)
	}
	networkSuffix := clientIDParts[len(clientIDParts)-1]
	if networkSuffix != s.IdenfySuffix {
		s.logger.Error("clientID has different network suffix", logger.Fields{"clientID": result.ClientID, "expectedSuffix": s.IdenfySuffix, "actualSuffix": networkSuffix})
		return errors.NewInternalError("invalid clientID", nil)
	}
	// delete the token with the same clientID and same scanRef
	result.ClientID = clientIDParts[0]

	err = s.tokenRepo.DeleteToken(ctx, result.ClientID, result.IdenfyRef)
	if err != nil {
		s.logger.Warn("Error deleting verification token from database", logger.Fields{"clientID": result.ClientID, "scanRef": result.IdenfyRef, "error": err})
	}
	// if the verification status is EXPIRED, we don't need to save it
	if result.Status.Overall != nil && *result.Status.Overall != models.Overall("EXPIRED") {
		// remove idenfy suffix from clientID
		err = s.verificationRepo.SaveVerification(ctx, &result)
		if err != nil {
			s.logger.Error("Error saving verification to database", logger.Fields{"clientID": result.ClientID, "scanRef": result.IdenfyRef, "error": err})
			return errors.NewInternalError("saving verification to database", err)
		}
	}
	s.logger.Debug("Verification result processed successfully", logger.Fields{"result": result})
	return nil
}

func (s *KYCService) ProcessDocExpirationNotification(ctx context.Context, clientID string) error {
	return nil
}

func (s *KYCService) IsUserVerified(ctx context.Context, clientID string) (bool, error) {
	verification, err := s.verificationRepo.GetVerification(ctx, clientID)
	if err != nil {
		s.logger.Error("Error getting verification from database", logger.Fields{"clientID": clientID, "error": err})
		return false, errors.NewInternalError("getting verification from database", err)
	}
	if verification == nil {
		return false, nil
	}
	return verification.Status.Overall != nil && (*verification.Status.Overall == models.OverallApproved || (s.config.SuspiciousVerificationOutcome == "APPROVED" && *verification.Status.Overall == models.OverallSuspected)), nil
}
