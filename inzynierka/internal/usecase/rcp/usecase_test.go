package rcp_test

import (
	"context"
	"testing"
	"time"

	du "erpbadger/internal/domain/rcp"
	rcpuc "erpbadger/internal/usecase/rcp"
)

type mockRcpRepo struct {
	plannedMinutes  int
	workedMinutes   int
	breakMinutes    int
	lastUpserted    *du.DailyTotal
}

func (m *mockRcpRepo) GetMeWorkStatus(ctx context.Context, userID int) ([]du.WorkStatus, error) {
	return nil, nil
}
func (m *mockRcpRepo) PostMeWorkStart(ctx context.Context, userID int) ([]du.WorkStart, error) {
	return nil, nil
}
func (m *mockRcpRepo) PostMeWorkStop(ctx context.Context, userID int) ([]du.WorkStop, error) {
	return nil, nil
}
func (m *mockRcpRepo) GetMeWorkTimesheet(ctx context.Context, userID int) ([]du.WorkTimesheet, error) {
	return nil, nil
}
func (m *mockRcpRepo) ListActiveEmployees(day time.Time) ([]int, error) {
	return []int{1}, nil
}
func (m *mockRcpRepo) GetPlannedMinutes(userID int, day time.Time) (int, error) {
	return m.plannedMinutes, nil
}
func (m *mockRcpRepo) GetWorkedAndBreakMinutes(userID int, day time.Time) (int, int, error) {
	return m.workedMinutes, m.breakMinutes, nil
}
func (m *mockRcpRepo) UpsertDailyTotal(dt du.DailyTotal) error {
	m.lastUpserted = &dt
	return nil
}

func TestRecalculateDailyTotal_Overtime(t *testing.T) {
	mock := &mockRcpRepo{
		plannedMinutes: 480, // 8h
		workedMinutes:  600, // 10h
		breakMinutes:   30,
	}
	uc := rcpuc.NewUsecase(mock)

	now := time.Now()
	err := uc.RecalculateDailyTotalsForDate(context.Background(), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.lastUpserted == nil {
		t.Fatal("expected daily total to be upserted")
	}
	if mock.lastUpserted.WorkedMinutes != 480 {
		t.Errorf("expected worked minutes 480, got %d", mock.lastUpserted.WorkedMinutes)
	}
	if mock.lastUpserted.OvertimeMinutes != 120 {
		t.Errorf("expected overtime minutes 120, got %d", mock.lastUpserted.OvertimeMinutes)
	}
	if mock.lastUpserted.AbsenceMinutes != 0 {
		t.Errorf("expected absence minutes 0, got %d", mock.lastUpserted.AbsenceMinutes)
	}
	if mock.lastUpserted.BreakMinutes != 30 {
		t.Errorf("expected break minutes 30, got %d", mock.lastUpserted.BreakMinutes)
	}
}

func TestRecalculateDailyTotal_Absence(t *testing.T) {
	mock := &mockRcpRepo{
		plannedMinutes: 480, // 8h
		workedMinutes:  300, // 5h
		breakMinutes:   15,
	}
	uc := rcpuc.NewUsecase(mock)

	now := time.Now()
	err := uc.RecalculateDailyTotalsForDate(context.Background(), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.lastUpserted.WorkedMinutes != 300 {
		t.Errorf("expected worked minutes 300, got %d", mock.lastUpserted.WorkedMinutes)
	}
	if mock.lastUpserted.OvertimeMinutes != 0 {
		t.Errorf("expected overtime minutes 0, got %d", mock.lastUpserted.OvertimeMinutes)
	}
	if mock.lastUpserted.AbsenceMinutes != 180 {
		t.Errorf("expected absence minutes 180, got %d", mock.lastUpserted.AbsenceMinutes)
	}
}

func TestRecalculateDailyTotal_DayOff(t *testing.T) {
	mock := &mockRcpRepo{
		plannedMinutes: 0,   // Dzień wolny od pracy
		workedMinutes:  240, // Przepracowano 4h
		breakMinutes:   0,
	}
	uc := rcpuc.NewUsecase(mock)

	now := time.Now()
	err := uc.RecalculateDailyTotalsForDate(context.Background(), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.lastUpserted.WorkedMinutes != 0 {
		t.Errorf("expected worked minutes 0, got %d", mock.lastUpserted.WorkedMinutes)
	}
	if mock.lastUpserted.OvertimeMinutes != 240 {
		t.Errorf("expected overtime minutes 240, got %d", mock.lastUpserted.OvertimeMinutes)
	}
	if mock.lastUpserted.AbsenceMinutes != 0 {
		t.Errorf("expected absence minutes 0, got %d", mock.lastUpserted.AbsenceMinutes)
	}
}
