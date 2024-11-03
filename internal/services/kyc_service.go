package services

import (
	"strings"

	"github.com/threefoldtech/tf-kyc-verifier/internal/clients/idenfy"
	"github.com/threefoldtech/tf-kyc-verifier/internal/clients/substrate"
	"github.com/threefoldtech/tf-kyc-verifier/internal/configs"
	"github.com/threefoldtech/tf-kyc-verifier/internal/errors"
	"github.com/threefoldtech/tf-kyc-verifier/internal/logger"
	"github.com/threefoldtech/tf-kyc-verifier/internal/repository"
)

const TFT_CONVERSION_FACTOR = 10000000

type kycService struct {
	verificationRepo repository.VerificationRepository
	tokenRepo        repository.TokenRepository
	idenfy           idenfy.IdenfyClient
	substrate        substrate.SubstrateClient
	config           *configs.Verification
	logger           logger.Logger
	IdenfySuffix     string
}

func NewKYCService(verificationRepo repository.VerificationRepository, tokenRepo repository.TokenRepository, idenfy idenfy.IdenfyClient, substrateClient substrate.SubstrateClient, config *configs.Config, logger logger.Logger) KYCService {
	idenfySuffix := GetIdenfySuffix(substrateClient, config)
	return &kycService{verificationRepo: verificationRepo, tokenRepo: tokenRepo, idenfy: idenfy, substrate: substrateClient, config: &config.Verification, logger: logger, IdenfySuffix: idenfySuffix}
}

func GetIdenfySuffix(substrateClient substrate.SubstrateClient, config *configs.Config) string {
	idenfySuffix := GetChainNetworkName(substrateClient)
	if config.Idenfy.Namespace != "" {
		idenfySuffix = config.Idenfy.Namespace + ":" + idenfySuffix
	}
	return idenfySuffix
}

func GetChainNetworkName(substrateClient substrate.SubstrateClient) string {
	chainName, err := substrateClient.GetChainName()
	if err != nil {
		panic(errors.NewInternalError("error getting chain name", err))
	}
	chainNameParts := strings.Split(chainName, " ")
	chainNetworkName := strings.ToLower(chainNameParts[len(chainNameParts)-1])
	return chainNetworkName
}
