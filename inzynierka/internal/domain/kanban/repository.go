package kanban

import "context"

type Repository interface {
	GetTickets(ctx context.Context) ([]Ticket, error)
	CreateTicket(ctx context.Context, t *Ticket) error
	UpdateTicket(ctx context.Context, id int, t *Ticket) error
	DeleteTicket(ctx context.Context, id int) error
	GetCalendar(ctx context.Context) ([]CalendarEntry, error)
	UpdateCalendar(ctx context.Context, entry *CalendarEntry) error
}
