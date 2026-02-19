package worker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestValidateAPIKey_NoAuth(t *testing.T) {
	// When no key is configured, all requests pass
	err := validateAPIKey(context.Background(), "")
	assert.NoError(t, err)
}

func TestValidateAPIKey_MissingMetadata(t *testing.T) {
	err := validateAPIKey(context.Background(), "secret-key")
	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestValidateAPIKey_MissingKey(t *testing.T) {
	md := metadata.New(map[string]string{"other-header": "value"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	err := validateAPIKey(ctx, "secret-key")
	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestValidateAPIKey_InvalidKey(t *testing.T) {
	md := metadata.New(map[string]string{apiKeyHeader: "wrong-key"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	err := validateAPIKey(ctx, "secret-key")
	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestValidateAPIKey_ValidKey(t *testing.T) {
	md := metadata.New(map[string]string{apiKeyHeader: "secret-key"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	err := validateAPIKey(ctx, "secret-key")
	assert.NoError(t, err)
}
