package services

import (
	"context"
	"fmt"
	"time"

	"github.com/threefoldtech/tf-kyc-verifier/internal/errors"
	"github.com/threefoldtech/tf-kyc-verifier/internal/logger"
	"github.com/threefoldtech/tf-kyc-verifier/internal/models"
)

func (s *kycService) GetorCreateVerificationToken(ctx context.Context, clientID string) (*models.Token, bool, error) {
	isVerified, err := s.IsUserVerified(ctx, clientID)
	if err != nil {
		s.logger.Error("Error checking if user is verified", logger.Fields{"clientID": clientID, "error": err})
		return nil, false, errors.NewInternalError("error getting verification status from database", err) // db error
	}
	if isVerified {
		return nil, false, errors.NewConflictError("user already verified", nil) // TODO: implement a custom error that can be converted in the handler to a 4xx such 409 status code
	}
	token, err_ := s.tokenRepo.GetToken(ctx, clientID)
	if err_ != nil {
		s.logger.Error("Error getting token from database", logger.Fields{"clientID": clientID, "error": err_})
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
		s.logger.Error("Error checking if user account has required balance", logger.Fields{"clientID": clientID, "error": err_})
		return nil, false, errors.NewExternalError("error checking if user account has required balance", err_)
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
		return nil, false, errors.NewExternalError("error creating iDenfy verification session", err_)
	}
	// save the token with the original clientID
	newToken.ClientID = clientID
	err_ = s.tokenRepo.SaveToken(ctx, &newToken)
	if err_ != nil {
		s.logger.Error("Error saving verification token to database", logger.Fields{"clientID": clientID, "error": err_})
	}

	return &newToken, true, nil
}

func (s *kycService) DeleteToken(ctx context.Context, clientID string, scanRef string) error {

	err := s.tokenRepo.DeleteToken(ctx, clientID, scanRef)
	if err != nil {
		s.logger.Error("Error deleting verification token from database", logger.Fields{"clientID": clientID, "scanRef": scanRef, "error": err})
		return errors.NewInternalError("error deleting verification token from database", err)
	}
	return nil
}

func (s *kycService) AccountHasRequiredBalance(ctx context.Context, address string) (bool, error) {
	if s.config.MinBalanceToVerifyAccount == 0 {
		s.logger.Warn("Minimum balance to verify account is 0 which is not recommended", logger.Fields{"address": address})
		return true, nil
	}
	balance, err := s.substrate.GetAccountBalance(address)
	if err != nil {
		s.logger.Error("Error getting account balance", logger.Fields{"address": address, "error": err})
		return false, errors.NewExternalError("error getting account balance", err)
	}
	return balance >= s.config.MinBalanceToVerifyAccount, nil
}
