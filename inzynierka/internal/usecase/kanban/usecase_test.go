package kanban_test

import (
	"context"
	"testing"

	du "erpbadger/internal/domain/kanban"
	kanbanuc "erpbadger/internal/usecase/kanban"
)

type mockKanbanRepo struct {
	tickets   []du.Ticket
	calendars []du.CalendarEntry
}

func (m *mockKanbanRepo) GetTickets(ctx context.Context) ([]du.Ticket, error) {
	return m.tickets, nil
}
func (m *mockKanbanRepo) CreateTicket(ctx context.Context, t *du.Ticket) error {
	t.ID = 1
	m.tickets = append(m.tickets, *t)
	return nil
}
func (m *mockKanbanRepo) UpdateTicket(ctx context.Context, id int, t *du.Ticket) error {
	return nil
}
func (m *mockKanbanRepo) DeleteTicket(ctx context.Context, id int) error {
	return nil
}
func (m *mockKanbanRepo) GetCalendar(ctx context.Context) ([]du.CalendarEntry, error) {
	return m.calendars, nil
}
func (m *mockKanbanRepo) UpdateCalendar(ctx context.Context, entry *du.CalendarEntry) error {
	m.calendars = append(m.calendars, *entry)
	return nil
}

func TestCreateTicket_Validation(t *testing.T) {
	mock := &mockKanbanRepo{}
	uc := kanbanuc.NewUsecase(mock)

	// Pusty tytuł powinien zwrócić błąd
	emptyTicket := &du.Ticket{Title: "   "}
	err := uc.CreateTicket(context.Background(), emptyTicket)
	if err == nil {
		t.Error("expected error for empty ticket title, got nil")
	}

	// Poprawny tytuł
	validTicket := &du.Ticket{Title: "Poprawić raporty"}
	err = uc.CreateTicket(context.Background(), validTicket)
	if err != nil {
		t.Errorf("unexpected error for valid ticket: %v", err)
	}
	if validTicket.ID != 1 {
		t.Errorf("expected ticket ID 1, got %d", validTicket.ID)
	}
}

func TestDeleteTicket_Validation(t *testing.T) {
	mock := &mockKanbanRepo{}
	uc := kanbanuc.NewUsecase(mock)

	err := uc.DeleteTicket(context.Background(), 0)
	if err == nil {
		t.Error("expected error for ID 0, got nil")
	}

	err = uc.DeleteTicket(context.Background(), 5)
	if err != nil {
		t.Errorf("unexpected error for valid ID: %v", err)
	}
}
