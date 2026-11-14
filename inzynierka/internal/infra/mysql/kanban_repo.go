package mysql

import (
	"context"
	"database/sql"
	"strings"

	du "erpbadger/internal/domain/kanban"
)

type KanbanRepo struct {
	db *sql.DB
}

func NewKanbanRepo(db *sql.DB) du.Repository {
	return &KanbanRepo{db: db}
}

func (r *KanbanRepo) GetTickets(ctx context.Context) ([]du.Ticket, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, title, description, type, status, assignees FROM tickets")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets := []du.Ticket{}
	for rows.Next() {
		var t du.Ticket
		var assigneesStr string
		if err := rows.Scan(&t.ID, &t.Title, &t.Desc, &t.Type, &t.Status, &assigneesStr); err != nil {
			return nil, err
		}
		if assigneesStr != "" {
			t.Assignees = strings.Split(assigneesStr, ",")
		} else {
			t.Assignees = []string{}
		}
		tickets = append(tickets, t)
	}
	return tickets, rows.Err()
}

func (r *KanbanRepo) CreateTicket(ctx context.Context, t *du.Ticket) error {
	assigneesStr := strings.Join(t.Assignees, ",")
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO tickets (title, description, type, status, assignees) VALUES (?, ?, ?, ?, ?)",
		t.Title, t.Desc, t.Type, t.Status, assigneesStr,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = int(id)
	return nil
}

func (r *KanbanRepo) UpdateTicket(ctx context.Context, id int, t *du.Ticket) error {
	assigneesStr := strings.Join(t.Assignees, ",")
	_, err := r.db.ExecContext(ctx,
		"UPDATE tickets SET title=?, description=?, type=?, status=?, assignees=? WHERE id=?",
		t.Title, t.Desc, t.Type, t.Status, assigneesStr, id,
	)
	if err != nil {
		return err
	}
	t.ID = id
	return nil
}

func (r *KanbanRepo) DeleteTicket(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM tickets WHERE id=?", id)
	return err
}

func (r *KanbanRepo) GetCalendar(ctx context.Context) ([]du.CalendarEntry, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT person, date_str, status FROM calendar")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []du.CalendarEntry{}
	for rows.Next() {
		var entry du.CalendarEntry
		if err := rows.Scan(&entry.Person, &entry.DateStr, &entry.Status); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (r *KanbanRepo) UpdateCalendar(ctx context.Context, entry *du.CalendarEntry) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO calendar (person, date_str, status) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE status=?",
		entry.Person, entry.DateStr, entry.Status, entry.Status,
	)
	return err
}
