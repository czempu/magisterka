package kanban

import (
	"context"
	"errors"
	"strings"

	du "erpbadger/internal/domain/kanban"
)

type Usecase interface {
	GetTickets(ctx context.Context) ([]du.Ticket, error)
	CreateTicket(ctx context.Context, t *du.Ticket) error
	UpdateTicket(ctx context.Context, id int, t *du.Ticket) error
	DeleteTicket(ctx context.Context, id int) error
	GetCalendar(ctx context.Context) ([]du.CalendarEntry, error)
	UpdateCalendar(ctx context.Context, entry *du.CalendarEntry) error
}

type usecase struct {
	repo du.Repository
}

func NewUsecase(r du.Repository) Usecase {
	return &usecase{repo: r}
}

func (u *usecase) GetTickets(ctx context.Context) ([]du.Ticket, error) {
	return u.repo.GetTickets(ctx)
}

func (u *usecase) CreateTicket(ctx context.Context, t *du.Ticket) error {
	if strings.TrimSpace(t.Title) == "" {
		return errors.New("tytuł zadania nie może być pusty")
	}
	return u.repo.CreateTicket(ctx, t)
}

func (u *usecase) UpdateTicket(ctx context.Context, id int, t *du.Ticket) error {
	if id <= 0 {
		return errors.New("nieprawidłowy identyfikator zadania")
	}
	return u.repo.UpdateTicket(ctx, id, t)
}

func (u *usecase) DeleteTicket(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("nieprawidłowy identyfikator zadania")
	}
	return u.repo.DeleteTicket(ctx, id)
}

func (u *usecase) GetCalendar(ctx context.Context) ([]du.CalendarEntry, error) {
	return u.repo.GetCalendar(ctx)
}

func (u *usecase) UpdateCalendar(ctx context.Context, entry *du.CalendarEntry) error {
	if strings.TrimSpace(entry.Person) == "" || strings.TrimSpace(entry.DateStr) == "" {
		return errors.New("osoba oraz data są wymagane")
	}
	return u.repo.UpdateCalendar(ctx, entry)
}
