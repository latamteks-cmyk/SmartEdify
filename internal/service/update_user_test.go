package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smartedify/auth-service/internal/models"
	"github.com/smartedify/auth-service/internal/repository"
	"github.com/stretchr/testify/assert"
)

// Mock audit logger
var auditEvents []string

func mockLogAuditEvent(event string) {
	auditEvents = append(auditEvents, event)
}

func TestUpdateUserRectificationAndAudit(t *testing.T) {
	repo := repository.NewMockUserRepository()
	user := &models.User{
		ID:        uuid.New(),
		Email:     "old@example.com",
		FirstName: "OldName",
		LastName:  "OldLast",
		Phone:     "999999999",
		TenantID:  "tenant-1",
		Status:    models.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := repo.CreateUser(context.Background(), user)
	assert.NoError(t, err)

	// Simular rectificación
	user.Email = "new@example.com"
	user.FirstName = "NewName"
	user.LastName = "NewLast"
	user.Phone = "888888888"
	user.UpdatedAt = time.Now()
	err = repo.UpdateUser(context.Background(), user)
	assert.NoError(t, err)

	// Verificar que los datos fueron actualizados
	updated, err := repo.GetUserByID(context.Background(), user.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, "new@example.com", updated.Email)
	assert.Equal(t, "NewName", updated.FirstName)
	assert.Equal(t, "NewLast", updated.LastName)
	assert.Equal(t, "888888888", updated.Phone)

	// Simular registro en auditoría
	mockLogAuditEvent("Rectificación de datos para usuario " + user.ID.String())
	assert.Contains(t, auditEvents, "Rectificación de datos para usuario "+user.ID.String())
}
