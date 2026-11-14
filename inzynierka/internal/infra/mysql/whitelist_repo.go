package mysql

import (
    "context"
    "database/sql"
    "erpbadger/internal/domain/whitelist"
)

type mysqlWhitelistRepo struct {
    db *sql.DB
}

func NewMySQLWhitelistRepo(db *sql.DB) whitelist.Repository {
    return &mysqlWhitelistRepo{db: db}
}

func (r *mysqlWhitelistRepo) IsAllowed(ctx context.Context, ip string) (bool, error) {
    const q = `SELECT 1 FROM whitelist WHERE IP = ? LIMIT 1`
    var dummy int
    err := r.db.QueryRowContext(ctx, q, ip).Scan(&dummy)
    if err == sql.ErrNoRows {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}

