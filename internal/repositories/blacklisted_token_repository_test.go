package repositories

import (
	"testing"

	"github.com/array/banking-api/internal/models"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestBlacklistedTokenRepository_GetByJTI_NotFound(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.BlacklistedToken{}))

	repo := NewBlacklistedTokenRepository(db)

	token, err := repo.GetByJTI("missing-jti")
	require.NoError(t, err)
	require.Nil(t, token)
}
