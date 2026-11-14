package sickness

import (
	"context"
	"fmt"
	"time"

	du "erpbadger/internal/domain/sickness"
	apuc "erpbadger/internal/usecase/approval"
)

type Usecase interface {
	List(ctx context.Context, employeeID *int, status *string) ([]du.Sickness, error)
	Get(ctx context.Context, id int) (*du.Sickness, error)
	Create(ctx context.Context, employeeID int,fromDate, toDate, legalBasis, certificateNumber string,percentCoverage int, isAccident bool,) (string, error)
	Delete(ctx context.Context, id int) error

	// jeśli chcesz mieć też ręczny update statusu z modułu:
	UpdateStatus(ctx context.Context, id int, status string) error
}

type usecase struct {
	repo        du.Repository
	approvalSvc apuc.Usecase
}

func NewUsecase(r du.Repository, ap apuc.Usecase) Usecase {
	return &usecase{repo: r, approvalSvc: ap}
}

func (u *usecase) List(ctx context.Context, employeeID *int, status *string) ([]du.Sickness, error) {
	return u.repo.List(ctx, employeeID, status)
}

func (u *usecase) Get(ctx context.Context, id int) (*du.Sickness, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *usecase) Create(
	ctx context.Context,
	employeeID int,
	fromDate, toDate, legalBasis, certificateNumber string,
	percentCoverage int,
	isAccident bool,
) (string, error) {

	if employeeID == 0 {
		return "", fmt.Errorf("employeeID jest wymagany")
	}
	if fromDate == "" || toDate == "" {
		return "", fmt.Errorf("from_date i to_date są wymagane")
	}
	if legalBasis == "" {
		return "", fmt.Errorf("legal_basis jest wymagana")
	}

	// prosta walidacja zakresu dat
	fromT, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		return "", fmt.Errorf("niepoprawny format from_date, oczekiwano YYYY-MM-DD")
	}
	toT, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		return "", fmt.Errorf("niepoprawny format to_date, oczekiwano YYYY-MM-DD")
	}
	if toT.Before(fromT) {
		return "", fmt.Errorf("to_date nie może być wcześniejsza niż from_date")
	}

	// domyślne % – w PL zazwyczaj 80, albo 100 przy wypadku
	if percentCoverage == 0 {
		if isAccident {
			percentCoverage = 100
		} else {
			percentCoverage = 80
		}
	}

	if len(legalBasis) > 255 {
		return "", fmt.Errorf("Przekroczono dopuszczalną długość podstawy prawnej (255)")
	}
	if len(certificateNumber) > 64 {
		return "", fmt.Errorf("Przekroczono dopuszczalną długość numeru zaświadczenia (64)")
	}



	s := du.Sickness{
		EmployeeID:             employeeID,
		FromDate:               fromDate,
		ToDate:                 toDate,
		LegalBasis:             legalBasis,
		CertificateNumber:      certificateNumber,
		PercentCoverage:        percentCoverage,
		IsOccupationalAccident: isAccident,
		Status:                 "pending",
	}


	exist, eerr := u.repo.SickExist(ctx, s)
	if eerr != nil {
		return "", eerr
	}
	if exist {
		return "błąd approval create", fmt.Errorf("Kolizja z inną chorobą pracownika")
	}


	id, err := u.repo.Create(ctx, s)
	if err != nil {
		return "", err
	}

	// PODPIĘCIE DO WORKFLOW – resource "l4"
	// target_id = ID rekordu chorobowego, requester = employeeID
	resp, err := u.approvalSvc.CreateApproval(ctx, "sick", int64(id), employeeID)
	if err != nil {
		return "błąd approval create", err
	}

	return resp, nil
}

func (u *usecase) Delete(ctx context.Context, id int) error {
	if id == 0 {
		return fmt.Errorf("id jest wymagane")
	}
	return u.repo.Delete(ctx, id)
}

func (u *usecase) UpdateStatus(ctx context.Context, id int, status string) error {
	if id == 0 {
		return fmt.Errorf("id jest wymagane")
	}
	if status == "" {
		return fmt.Errorf("status jest wymagany")
	}
	return u.repo.UpdateStatus(ctx, id, status)
}

