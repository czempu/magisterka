package whitelist

import "context"

// Repository ukrywa szczegóły, skąd bierzemy listę IP
type Repository interface {
    IsAllowed(ctx context.Context, ip string) (bool, error)
}

