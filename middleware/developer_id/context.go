package developer_id

import (
	"context"
)

type developerIDKey struct{}

var DeveloperIDKey = developerIDKey{}

// GetDeveloperIDFromContext 从 context 中获取开发者 ID
func GetDeveloperIDFromContext(ctx context.Context) string {
	if developerID := ctx.Value(DeveloperIDKey); developerID != nil {
		if id, ok := developerID.(string); ok {
			return id
		}
	}
	return ""
}

// WithDeveloperID 将开发者 ID 存入 context
func WithDeveloperID(ctx context.Context, developerID string) context.Context {
	return context.WithValue(ctx, DeveloperIDKey, developerID)
}
