/*
Package services contains the services for the application.
This layer is responsible for handling the business logic.
*/
package services

import (
	"fmt"
	"strings"

	"github.com/threefoldtech/tf-kyc-verifier/internal/clients/idenfy"
	"github.com/threefoldtech/tf-kyc-verifier/internal/clients/substrate"
	"github.com/threefoldtech/tf-kyc-verifier/internal/config"
	"github.com/threefoldtech/tf-kyc-verifier/internal/logger"
	"github.com/threefoldtech/tf-kyc-verifier/internal/repository"
)

const TFT_CONVERSION_FACTOR = 10000000

type kycService struct {
	verificationRepo repository.VerificationRepository
	tokenRepo        repository.TokenRepository
	idenfy           idenfy.IdenfyClient
	substrate        substrate.SubstrateClient
	config           *config.Verification
	logger           logger.Logger
	IdenfySuffix     string
}

func NewKYCService(verificationRepo repository.VerificationRepository, tokenRepo repository.TokenRepository, idenfy idenfy.IdenfyClient, substrateClient substrate.SubstrateClient, config *config.Config, logger logger.Logger) (KYCService, error) {
	idenfySuffix, err := GetIdenfySuffix(substrateClient, config)
	if err != nil {
		return nil, fmt.Errorf("getting idenfy suffix: %w", err)
	}
	return &kycService{verificationRepo: verificationRepo, tokenRepo: tokenRepo, idenfy: idenfy, substrate: substrateClient, config: &config.Verification, logger: logger, IdenfySuffix: idenfySuffix}, nil
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
