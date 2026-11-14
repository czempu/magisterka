package mysql

import (
    "context"
    "database/sql"
    "erpbadger/internal/domain/stats"
)

type StatsRepo struct {
    db *sql.DB
}

func NewStatsRepo(db *sql.DB) stats.Repository {
    return &StatsRepo{db: db}
}




func (r *StatsRepo) GetEmpStats(ctx context.Context) (stats.EmpStats, error) {
	// Zapytanie SQL z dwoma podzapytaniami, zwracające jeden wiersz z dwiema kolumnami:
	// 'active_entries' (z timesheet_entries) i 'all_users' (z users).
	q := `
		SELECT
			(SELECT COUNT(*) FROM timesheet_entries WHERE started_at IS NOT NULL AND ended_at IS NULL) AS active_entries,
			(SELECT COUNT(*) FROM users) AS all_users
	`

	var results stats.EmpStats

	// Używamy QueryRow, ponieważ oczekujemy tylko jednego wiersza wynikowego.
	// Następnie używamy Scan do przypisania wartości z kolumn do pól struktury EmpStats.
	err := r.db.QueryRowContext(ctx, q).Scan(
		&results.Active, // Mapowanie 'active_entries' do pola Active
		&results.Total,  // Mapowanie 'all_users' do pola Total
	)

	if err != nil {
		// Sprawdzamy, czy błąd to brak wierszy, choć w tym przypadku COUNT() zawsze zwraca 0,
		// więc taki błąd jest mało prawdopodobny, ale obsługa błędów jest konieczna.
		if err == sql.ErrNoRows {
			// Zwracamy pustą statystykę i brak błędu, jeśli z jakiegoś powodu nie było wierszy
			// (w tym przypadku to niemożliwe, bo COUNT() zawsze zwraca wiersz).
			return stats.EmpStats{}, nil
		}
		return stats.EmpStats{}, err
	}

	return results, nil
}





