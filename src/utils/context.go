package utils

import "context"

func CtxValueOrDefault[K any](ctx context.Context, key any, defaultValue K) K {
	value := ctx.Value(key)
	if value == nil {
		return defaultValue
	}
	return value.(K)
}
