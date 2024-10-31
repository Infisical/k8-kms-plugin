package main

import (
	"errors"
	"fmt"
	"log"

	infisicalSdk "github.com/infisical/go-sdk"
)

type AuthStrategyType string

var AuthStrategy = struct {
	SERVICE_TOKEN                 AuthStrategyType
	SERVICE_ACCOUNT               AuthStrategyType
	UNIVERSAL_MACHINE_IDENTITY    AuthStrategyType
	KUBERNETES_MACHINE_IDENTITY   AuthStrategyType
	AWS_IAM_MACHINE_IDENTITY      AuthStrategyType
	AZURE_MACHINE_IDENTITY        AuthStrategyType
	GCP_ID_TOKEN_MACHINE_IDENTITY AuthStrategyType
	GCP_IAM_MACHINE_IDENTITY      AuthStrategyType
}{
	SERVICE_TOKEN:                 "SERVICE_TOKEN",
	SERVICE_ACCOUNT:               "SERVICE_ACCOUNT",
	UNIVERSAL_MACHINE_IDENTITY:    "UNIVERSAL_MACHINE_IDENTITY",
	KUBERNETES_MACHINE_IDENTITY:   "KUBERNETES_AUTH_MACHINE_IDENTITY",
	AWS_IAM_MACHINE_IDENTITY:      "AWS_IAM_MACHINE_IDENTITY",
	AZURE_MACHINE_IDENTITY:        "AZURE_MACHINE_IDENTITY",
	GCP_ID_TOKEN_MACHINE_IDENTITY: "GCP_ID_TOKEN_MACHINE_IDENTITY",
	GCP_IAM_MACHINE_IDENTITY:      "GCP_IAM_MACHINE_IDENTITY",
}

type AuthenticationDetails struct {
	authStrategy AuthStrategyType
}

type AuthHandler struct {
	infisicalClient           *infisicalSdk.InfisicalClientInterface
	clientId                  string
	clientSecret              string
	identityId                string
	resource                  string
	serviceAccountKeyfilePath string
}

var ErrAuthNotApplicable = errors.New("authentication not applicable")

func (r *AuthHandler) handleUniversalAuth() (AuthenticationDetails, error) {
	if r.clientId == "" && r.clientSecret == "" {
		return AuthenticationDetails{}, ErrAuthNotApplicable
	}

	_, err := (*r.infisicalClient).Auth().UniversalAuthLogin(r.clientId, r.clientSecret)
	if err != nil {
		return AuthenticationDetails{}, fmt.Errorf("unable to login with machine identity credentials [err=%s]", err)
	}

	return AuthenticationDetails{authStrategy: AuthStrategy.UNIVERSAL_MACHINE_IDENTITY}, nil
}

func (r *AuthHandler) handleAwsIamAuth() (AuthenticationDetails, error) {
	if r.identityId == "" {
		return AuthenticationDetails{}, ErrAuthNotApplicable
	}

	_, err := (*r.infisicalClient).Auth().AwsIamAuthLogin(r.identityId)
	if err != nil {
		return AuthenticationDetails{}, fmt.Errorf("unable to login with AWS IAM auth [err=%s]", err)
	}

	return AuthenticationDetails{authStrategy: AuthStrategy.AWS_IAM_MACHINE_IDENTITY}, nil
}

func (r *AuthHandler) handleAzureAuth() (AuthenticationDetails, error) {
	if r.identityId == "" {
		return AuthenticationDetails{}, ErrAuthNotApplicable
	}

	_, err := (*r.infisicalClient).Auth().AzureAuthLogin(r.identityId, r.resource) // If resource is empty(""), it will default to "https://management.azure.com/" in the SDK.
	if err != nil {
		return AuthenticationDetails{}, fmt.Errorf("unable to login with Azure auth [err=%s]", err)
	}

	return AuthenticationDetails{authStrategy: AuthStrategy.AZURE_MACHINE_IDENTITY}, nil
}

func (r *AuthHandler) handleGcpIdTokenAuth() (AuthenticationDetails, error) {
	if r.identityId == "" {
		return AuthenticationDetails{}, ErrAuthNotApplicable
	}

	_, err := (*r.infisicalClient).Auth().GcpIdTokenAuthLogin(r.identityId)
	if err != nil {
		return AuthenticationDetails{}, fmt.Errorf("unable to login with GCP Id Token auth [err=%s]", err)
	}

	return AuthenticationDetails{authStrategy: AuthStrategy.GCP_ID_TOKEN_MACHINE_IDENTITY}, nil
}

func (r *AuthHandler) handleGcpIamAuth() (AuthenticationDetails, error) {
	if r.identityId == "" && r.serviceAccountKeyfilePath == "" {
		return AuthenticationDetails{}, ErrAuthNotApplicable
	}

	_, err := (*r.infisicalClient).Auth().GcpIamAuthLogin(r.identityId, r.serviceAccountKeyfilePath)
	if err != nil {
		return AuthenticationDetails{}, fmt.Errorf("unable to login with GCP IAM auth [err=%s]", err)
	}

	return AuthenticationDetails{authStrategy: AuthStrategy.GCP_IAM_MACHINE_IDENTITY}, nil
}

func (r *AuthHandler) login() error {
	authStrategies := map[AuthStrategyType]func() (AuthenticationDetails, error){
		AuthStrategy.UNIVERSAL_MACHINE_IDENTITY:    r.handleUniversalAuth,
		AuthStrategy.AWS_IAM_MACHINE_IDENTITY:      r.handleAwsIamAuth,
		AuthStrategy.AZURE_MACHINE_IDENTITY:        r.handleAzureAuth,
		AuthStrategy.GCP_ID_TOKEN_MACHINE_IDENTITY: r.handleGcpIdTokenAuth,
		AuthStrategy.GCP_IAM_MACHINE_IDENTITY:      r.handleGcpIamAuth,
	}

	for authStrategy, authHandler := range authStrategies {
		authDetails, err := authHandler()
		if err == nil {
			log.Printf("Using auth method: %s\n", authDetails.authStrategy)
			return nil
		}

		if !errors.Is(err, ErrAuthNotApplicable) {
			return fmt.Errorf("authentication failed for strategy [%s] [err=%w]", authStrategy, err)
		}
	}

	return fmt.Errorf("no valid authentication provided")
}
