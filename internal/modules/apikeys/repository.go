package apikeys

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/google/uuid"
)

type Repository struct{ DB *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{DB: db} }

func (r *Repository) ListClients(companyID uuid.UUID) ([]APIClient, error) {
	var clients []APIClient
	err := r.DB.
		Where("company_id = ? AND deleted_at IS NULL", companyID).
		Order("created_at DESC").
		Find(&clients).Error
	return clients, err
}

func (r *Repository) CreateClient(client APIClient) (APIClient, error) {
	if err := r.DB.Create(&client).Error; err != nil {
		return APIClient{}, err
	}
	return client, nil
}

func (r *Repository) FindClient(companyID, clientID uuid.UUID) (APIClient, error) {
	var client APIClient
	err := r.DB.
		Where("company_id = ? AND id = ? AND deleted_at IS NULL", companyID, clientID).
		First(&client).Error
	return client, err
}

func (r *Repository) CreateKey(key APIKey) error {
	return r.DB.Create(&key).Error
}

func (r *Repository) RevokeKey(companyID, clientID uuid.UUID, keyID string) error {
	now := time.Now().UTC()
	return r.DB.Model(&APIKey{}).
		Where("company_id = ? AND client_id = ? AND key_id = ? AND status <> ?", companyID, clientID, keyID, "revoked").
		Updates(map[string]any{
			"status":     "revoked",
			"revoked_at": &now,
		}).Error
}

func (r *Repository) CreatePublicKey(publicKey APIClientPublicKey) (APIClientPublicKey, error) {
	if err := r.DB.Create(&publicKey).Error; err != nil {
		return APIClientPublicKey{}, err
	}
	return publicKey, nil
}

func (r *Repository) FindPublicKey(companyID, clientID, publicKeyID uuid.UUID) (APIClientPublicKey, error) {
	var publicKey APIClientPublicKey
	err := r.DB.
		Where("company_id = ? AND client_id = ? AND id = ?", companyID, clientID, publicKeyID).
		First(&publicKey).Error
	return publicKey, err
}

func (r *Repository) ActivatePublicKey(companyID, clientID, publicKeyID uuid.UUID) error {
	now := time.Now().UTC()
	return r.DB.Model(&APIClientPublicKey{}).
		Where("company_id = ? AND client_id = ? AND id = ?", companyID, clientID, publicKeyID).
		Updates(map[string]any{
			"status":       "active",
			"activated_at": &now,
		}).Error
}

func (r *Repository) RevokePublicKey(companyID, clientID, publicKeyID uuid.UUID) error {
	now := time.Now().UTC()
	return r.DB.Model(&APIClientPublicKey{}).
		Where("company_id = ? AND client_id = ? AND id = ?", companyID, clientID, publicKeyID).
		Updates(map[string]any{
			"status":     "revoked",
			"revoked_at": &now,
		}).Error
}

func marshalScopes(scopes []string) datatypes.JSON {
	data, _ := json.Marshal(scopes)
	return data
}
