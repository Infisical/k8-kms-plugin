package main

import (
	"context"

	infisical "github.com/infisical/go-sdk"
)

type InfisicalKmsService struct {
	infisicalClient *infisical.InfisicalClientInterface
	kmsKeyId        string
}

func (s *InfisicalKmsService) Decrypt(ctx context.Context, uid string, req *DecryptRequest) ([]byte, error) {
	decryptedKey, err := (*s.infisicalClient).Kms().DecryptData(infisical.KmsDecryptDataOptions{
		KeyId:      req.KeyID,
		Ciphertext: string(req.Ciphertext),
	})

	if err != nil {
		return []byte{}, err
	}

	return []byte(decryptedKey), nil
}

func (s *InfisicalKmsService) Encrypt(ctx context.Context, uid string, data []byte) (*EncryptResponse, error) {
	encryptedKey, err := (*s.infisicalClient).Kms().EncryptData(infisical.KmsEncryptDataOptions{
		KeyId:     s.kmsKeyId,
		Plaintext: string(data),
	})

	if err != nil {
		return &EncryptResponse{}, err
	}

	return &EncryptResponse{
		Ciphertext: []byte(encryptedKey),
		KeyID:      s.kmsKeyId,
	}, nil
}

func (s *InfisicalKmsService) Status(ctx context.Context) (*StatusResponse, error) {
	_, err := s.Encrypt(ctx, "test-status-uid", []byte("test-encrypt-payload"))
	if err != nil {
		return nil, err
	}

	return &StatusResponse{
		Version: "v2",
		Healthz: "ok",
		KeyID:   s.kmsKeyId,
	}, nil
}

func NewInfisicalKmsService(infisicalClient *infisical.InfisicalClientInterface, kmsKeyId string) *InfisicalKmsService {
	return &InfisicalKmsService{
		infisicalClient: infisicalClient,
		kmsKeyId:        kmsKeyId,
	}
}
