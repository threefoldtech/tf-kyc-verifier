package services

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"example.com/tfgrid-kyc-service/internal/clients/idenfy"
	"example.com/tfgrid-kyc-service/internal/clients/substrate"
	"example.com/tfgrid-kyc-service/internal/configs"
	"example.com/tfgrid-kyc-service/internal/models"
	"example.com/tfgrid-kyc-service/internal/repository"
)

type kycService struct {
	verificationRepo repository.VerificationRepository
	tokenRepo        repository.TokenRepository
	idenfy           *idenfy.Idenfy
	substrate        *substrate.Substrate
	config           *configs.Verification
}

func NewKYCService(verificationRepo repository.VerificationRepository, tokenRepo repository.TokenRepository, idenfy *idenfy.Idenfy, substrateClient *substrate.Substrate, config *configs.Verification) KYCService {
	return &kycService{verificationRepo: verificationRepo, tokenRepo: tokenRepo, idenfy: idenfy, substrate: substrateClient, config: config}
}

// ---------------------------------------------------------------------------------------------------------------------
// token related methods
// ---------------------------------------------------------------------------------------------------------------------

func (s *kycService) GetorCreateVerificationToken(ctx context.Context, clientID string) (*models.Token, bool, error) {
	isVerified, err := s.IsUserVerified(ctx, clientID)
	if err != nil {
		return nil, false, err
	}
	if isVerified {
		return nil, false, errors.New("user already verified") // TODO: implement a custom error that can be converted in the handler to a 400 status code
	}
	token, err := s.tokenRepo.GetToken(ctx, clientID)
	if err != nil {
		return nil, false, err
	}
	// check if token is not nil and not expired or near expiry (2 min)
	if token != nil { //&& time.Since(token.CreatedAt)+2*time.Minute < time.Duration(token.ExpiryTime)*time.Second {
		return token, false, nil
	}
	fmt.Println("token is nil or expired")
	// check if user account balance satisfies the minimum required balance, return an error if not
	hasRequiredBalance, err := s.AccountHasRequiredBalance(ctx, clientID)
	if err != nil {
		return nil, false, err // todo: implement a custom error that can be converted in the handler to a 500 status code
	}
	if !hasRequiredBalance {
		return nil, false, errors.New("account does not have the required balance") // todo: implement a custom error that can be converted in the handler to a 402 status code
	}
	newToken, err := s.idenfy.CreateVerificationSession(ctx, clientID)
	if err != nil {
		return nil, false, err
	}
	fmt.Println("new token", newToken)
	err = s.tokenRepo.SaveToken(ctx, &newToken)
	if err != nil {
		fmt.Println("warning: was not able to save verification token to db", err)
	}

	return &newToken, true, nil
}

func (s *kycService) DeleteToken(ctx context.Context, clientID string, scanRef string) error {
	return s.tokenRepo.DeleteToken(ctx, clientID, scanRef)
}

func (s *kycService) AccountHasRequiredBalance(ctx context.Context, address string) (bool, error) {
	if s.config.MinBalanceToVerifyAccount == 0 {
		return true, nil
	}
	balance, err := s.substrate.GetAccountBalance(address)
	if err != nil {
		return false, err
	}
	return balance.Cmp(big.NewInt(int64(s.config.MinBalanceToVerifyAccount))) >= 0, nil
}

// ---------------------------------------------------------------------------------------------------------------------
// verification related methods
// ---------------------------------------------------------------------------------------------------------------------

func (s *kycService) GetVerification(ctx context.Context, clientID string) (*models.Verification, error) {
	verification, err := s.verificationRepo.GetVerification(ctx, clientID)
	if err != nil {
		return nil, err
	}
	return verification, nil
}

func (s *kycService) GetVerificationStatus(ctx context.Context, clientID string) (*models.VerificationOutcome, error) {
	verification, err := s.GetVerification(ctx, clientID)
	if err != nil {
		return nil, err
	}
	var outcome string
	if verification != nil {
		if verification.Status.Overall == "APPROVED" || (s.config.SuspiciousVerificationOutcome == "APPROVED" && verification.Status.Overall == "SUSPECTED") {
			outcome = "APPROVED"
		} else {
			outcome = "REJECTED"
		}
	} else {
		return nil, nil
	}
	return &models.VerificationOutcome{
		Final:     verification.Final,
		ClientID:  clientID,
		IdenfyRef: verification.ScanRef,
		Outcome:   outcome,
	}, nil
}

func (s *kycService) GetVerificationStatusByTwinID(ctx context.Context, twinID string) (*models.VerificationOutcome, error) {
	// get the address from the twinID
	address, err := s.substrate.GetAddressByTwinID(twinID)
	if err != nil {
		return nil, err
	}
	return s.GetVerificationStatus(ctx, address)
}

func (s *kycService) ProcessVerificationResult(ctx context.Context, body []byte, sigHeader string, result models.Verification) error {
	err := s.idenfy.VerifyCallbackSignature(ctx, body, sigHeader)
	if err != nil {
		return err
	}
	// delete the token with the same clientID and same scanRef
	err = s.tokenRepo.DeleteToken(ctx, result.ClientID, result.ScanRef)
	if err != nil {
		fmt.Printf("error deleting token: %v", err)
	}
	// if the verification status is EXPIRED, we don't need to save it
	if result.Status.Overall != "EXPIRED" {
		err = s.verificationRepo.SaveVerification(ctx, &result)
		if err != nil {
			fmt.Printf("error saving verification to the database: %v", err)
			return err
		}
	}
	// fmt the result
	fmt.Println(result)
	return nil
}

func (s *kycService) ProcessDocExpirationNotification(ctx context.Context, clientID string) error {
	return nil
}

func (s *kycService) IsUserVerified(ctx context.Context, clientID string) (bool, error) {
	verification, err := s.verificationRepo.GetVerification(ctx, clientID)
	if err != nil {
		return false, err
	}
	if verification == nil {
		return false, nil
	}
	return verification.Status.Overall == "APPROVED" || (s.config.SuspiciousVerificationOutcome == "APPROVED" && verification.Status.Overall == "SUSPECTED"), nil
}
