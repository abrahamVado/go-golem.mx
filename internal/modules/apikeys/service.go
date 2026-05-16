package apikeys

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/abrahamVado/go-paladin.mx/internal/security"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrClientNameRequired        = errors.New("client name is required")
	ErrClientNotFound            = errors.New("api client not found")
	ErrScopesRequired            = errors.New("at least one scope is required")
	ErrPublicKeyRequired         = errors.New("openssh public key is required")
	ErrPublicKeyInvalid          = errors.New("invalid openssh ed25519 public key")
	ErrPublicKeyNotFound         = errors.New("public key not found")
	ErrChallengeRequired         = errors.New("challenge is required")
	ErrChallengeSignatureInvalid = errors.New("invalid challenge signature")
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) ListClients(companyID uuid.UUID) ([]APIClient, error) {
	return s.repo.ListClients(companyID)
}

func (s *Service) CreateClient(companyID, userID uuid.UUID, req CreateClientRequest) (APIClient, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return APIClient{}, ErrClientNameRequired
	}

	return s.repo.CreateClient(APIClient{
		ID:              uuid.New(),
		CompanyID:       companyID,
		Name:            name,
		Description:     strings.TrimSpace(req.Description),
		Status:          "active",
		CreatedByUserID: &userID,
	})
}

func (s *Service) CreateKey(companyID, clientID uuid.UUID, req CreateKeyRequest) (CreateKeyResponse, error) {
	if _, err := s.repo.FindClient(companyID, clientID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return CreateKeyResponse{}, ErrClientNotFound
		}
		return CreateKeyResponse{}, err
	}

	scopes := normalizeScopes(req.Scopes)
	if len(scopes) == 0 {
		return CreateKeyResponse{}, ErrScopesRequired
	}

	secret, err := security.NewOpaqueToken(32)
	if err != nil {
		return CreateKeyResponse{}, err
	}

	keyIDSecret, err := security.NewOpaqueToken(12)
	if err != nil {
		return CreateKeyResponse{}, err
	}
	keyID := fmt.Sprintf("gk_%s", keyIDSecret)

	var expiresAt *time.Time
	var expiresAtRaw *string
	if req.ExpiresAt != nil && strings.TrimSpace(*req.ExpiresAt) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.ExpiresAt))
		if err != nil {
			return CreateKeyResponse{}, err
		}
		utc := parsed.UTC()
		expiresAt = &utc
		formatted := utc.Format(time.RFC3339)
		expiresAtRaw = &formatted
	}

	if err := s.repo.CreateKey(APIKey{
		ID:         uuid.New(),
		CompanyID:  companyID,
		ClientID:   clientID,
		KeyID:      keyID,
		SecretHash: security.HashToken(secret),
		Scopes:     marshalScopes(scopes),
		ExpiresAt:  expiresAt,
		Status:     "active",
	}); err != nil {
		return CreateKeyResponse{}, err
	}

	return CreateKeyResponse{
		KeyID:            keyID,
		Scopes:           scopes,
		ExpiresAt:        expiresAtRaw,
		APIKeySecretOnce: secret,
		FullTokenOnce:    keyID + "." + secret,
	}, nil
}

func (s *Service) RevokeKey(companyID, clientID uuid.UUID, keyID string) error {
	if _, err := s.repo.FindClient(companyID, clientID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrClientNotFound
		}
		return err
	}
	return s.repo.RevokeKey(companyID, clientID, strings.TrimSpace(keyID))
}

func (s *Service) UploadPublicKey(companyID, clientID uuid.UUID, req UploadPublicKeyRequest) (APIClientPublicKey, error) {
	if _, err := s.repo.FindClient(companyID, clientID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return APIClientPublicKey{}, ErrClientNotFound
		}
		return APIClientPublicKey{}, err
	}

	rawKey, err := parseOpenSSHEd25519PublicKey(req.OpenSSHPublicKey)
	if err != nil {
		return APIClientPublicKey{}, err
	}

	fingerprint := sha256Fingerprint(rawKey)
	return s.repo.CreatePublicKey(APIClientPublicKey{
		ID:                uuid.New(),
		CompanyID:         companyID,
		ClientID:          clientID,
		Algorithm:         "ed25519",
		PublicKeyRaw:      rawKey,
		FingerprintSHA256: fingerprint,
		SourceFormat:      "openssh",
		Status:            "pending",
	})
}

func (s *Service) ActivatePublicKey(companyID, clientID, publicKeyID uuid.UUID, req ActivatePublicKeyRequest) error {
	publicKey, err := s.repo.FindPublicKey(companyID, clientID, publicKeyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPublicKeyNotFound
		}
		return err
	}

	challenge := strings.TrimSpace(req.Challenge)
	if challenge == "" {
		return ErrChallengeRequired
	}

	signature, err := base64.StdEncoding.DecodeString(strings.TrimSpace(req.ChallengeSignature))
	if err != nil {
		signature, err = base64.RawStdEncoding.DecodeString(strings.TrimSpace(req.ChallengeSignature))
		if err != nil {
			return ErrChallengeSignatureInvalid
		}
	}

	if !ed25519.Verify(ed25519.PublicKey(publicKey.PublicKeyRaw), []byte(challenge), signature) {
		return ErrChallengeSignatureInvalid
	}

	return s.repo.ActivatePublicKey(companyID, clientID, publicKeyID)
}

func (s *Service) RevokePublicKey(companyID, clientID, publicKeyID uuid.UUID) error {
	if _, err := s.repo.FindPublicKey(companyID, clientID, publicKeyID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPublicKeyNotFound
		}
		return err
	}
	return s.repo.RevokePublicKey(companyID, clientID, publicKeyID)
}

func normalizeScopes(scopes []string) []string {
	out := make([]string, 0, len(scopes))
	seen := map[string]struct{}{}
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if _, exists := seen[scope]; exists {
			continue
		}
		seen[scope] = struct{}{}
		out = append(out, scope)
	}
	return out
}

func parseOpenSSHEd25519PublicKey(raw string) ([]byte, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, ErrPublicKeyRequired
	}

	parts := strings.Fields(value)
	if len(parts) < 2 || parts[0] != "ssh-ed25519" {
		return nil, ErrPublicKeyInvalid
	}

	blob, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrPublicKeyInvalid
	}
	if len(blob) < 4 {
		return nil, ErrPublicKeyInvalid
	}

	readString := func(input []byte, offset *int) ([]byte, error) {
		if len(input[*offset:]) < 4 {
			return nil, ErrPublicKeyInvalid
		}
		size := int(binary.BigEndian.Uint32(input[*offset : *offset+4]))
		*offset += 4
		if size < 0 || len(input[*offset:]) < size {
			return nil, ErrPublicKeyInvalid
		}
		value := input[*offset : *offset+size]
		*offset += size
		return value, nil
	}

	offset := 0
	algo, err := readString(blob, &offset)
	if err != nil || string(algo) != "ssh-ed25519" {
		return nil, ErrPublicKeyInvalid
	}
	key, err := readString(blob, &offset)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return nil, ErrPublicKeyInvalid
	}

	return key, nil
}

func sha256Fingerprint(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "SHA256:" + base64.StdEncoding.EncodeToString(sum[:])
}
