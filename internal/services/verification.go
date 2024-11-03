package services

import (
	"context"
	"slices"
	"strings"

	"github.com/threefoldtech/tf-kyc-verifier/internal/errors"
	"github.com/threefoldtech/tf-kyc-verifier/internal/logger"
	"github.com/threefoldtech/tf-kyc-verifier/internal/models"
)

func (s *kycService) GetVerificationData(ctx context.Context, clientID string) (*models.Verification, error) {
	verification, err := s.verificationRepo.GetVerification(ctx, clientID)
	if err != nil {
		s.logger.Error("Error getting verification from database", logger.Fields{"clientID": clientID, "error": err})
		return nil, errors.NewInternalError("error getting verification from database", err)
	}
	return verification, nil
}

func (s *kycService) GetVerificationStatus(ctx context.Context, clientID string) (*models.VerificationOutcome, error) {
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
		s.logger.Error("Error getting address from twinID", logger.Fields{"twinID": twinID, "error": err})
		return nil, errors.NewExternalError("error looking up twinID address from TFChain", err)
	}
	return s.GetVerificationStatus(ctx, address)
}

func (s *kycService) ProcessVerificationResult(ctx context.Context, body []byte, sigHeader string, result models.Verification) error {
	err := s.idenfy.VerifyCallbackSignature(ctx, body, sigHeader)
	if err != nil {
		s.logger.Error("Error verifying callback signature", logger.Fields{"sigHeader": sigHeader, "error": err})
		return errors.NewAuthorizationError("error verifying callback signature", err)
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
			return errors.NewInternalError("error saving verification to database", err)
		}
	}
	s.logger.Debug("Verification result processed successfully", logger.Fields{"result": result})
	return nil
}

func (s *kycService) ProcessDocExpirationNotification(ctx context.Context, clientID string) error {
	return nil
}

func (s *kycService) IsUserVerified(ctx context.Context, clientID string) (bool, error) {
	verification, err := s.verificationRepo.GetVerification(ctx, clientID)
	if err != nil {
		s.logger.Error("Error getting verification from database", logger.Fields{"clientID": clientID, "error": err})
		return false, errors.NewInternalError("error getting verification from database", err)
	}
	if verification == nil {
		return false, nil
	}
	return verification.Status.Overall != nil && (*verification.Status.Overall == models.OverallApproved || (s.config.SuspiciousVerificationOutcome == "APPROVED" && *verification.Status.Overall == models.OverallSuspected)), nil
}
