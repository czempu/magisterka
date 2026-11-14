package approval

import (
	"context"
	"fmt"
	"time"
	"strconv"
	du "erpbadger/internal/domain/approval"
	org "erpbadger/internal/domain/organization"
	wfuc "erpbadger/internal/usecase/workflow"
	peuc "erpbadger/internal/usecase/permissions"
)

type Usecase interface {
	CreateApproval(ctx context.Context, resource string, targetID int64, requesterID int) (string, error)
	Decision(ctx context.Context, stepID int64, approverID int, decision string, note *string) (string, error)
	ListPending(ctx context.Context, approverID int) ([]du.ApprovalStepList, error)
}

type usecase struct {
	repo        du.Repository
	orgRepo     org.Repository
	workflowSvc wfuc.Usecase
	permissionsSvc peuc.Usecase
}

func NewUsecase(r du.Repository, o org.Repository, wf wfuc.Usecase, pe peuc.Usecase) Usecase {
	return &usecase{
		repo:        r,
		orgRepo:     o,
		workflowSvc: wf,
		permissionsSvc: pe,
	}
}





// -----------------------------
// CREATE (tworzy instancję procesu wg workflow)
// -----------------------------
func (u *usecase) CreateApproval(ctx context.Context, resource string, targetID int64, requesterID int) (string, error) {
	// 1) pobierz definicję workflowa
	workflowDef, err := u.workflowSvc.GetWorkflowForResource(ctx, resource)

	// Jeśli brak workflowa (lub NULL w approval_policies) → auto-accept
	if err != nil || workflowDef.ID == 0 {
		now := time.Now()
		req := du.ApprovalRequest{
			Resource:  resource,
			TargetID:  targetID,
			Requester: requesterID,
			Status:    "approved",
			CreatedAt: now,
			ClosedAt:  &now,
		}

		if _, err := u.repo.CreateRequest(ctx, req); err != nil {
			return "", fmt.Errorf("nie udało się utworzyć auto-approved requesta: %w", err)
		}
		return "Brak workflow — wniosek automatycznie zaakceptowany", nil
	}

	// Helper: parsowanie slug -> int (z polityką "jeśli brak/źle: pomiń krok")
	parseSlugInt := func(slug *string) (int, bool) {
		if slug == nil {
			return 0, false
		}
		v, err := strconv.Atoi(*slug)
		if err != nil {
			return 0, false
		}
		return v, true
	}

	// 2) Najpierw przygotuj kroki, które DA SIĘ utworzyć (reszta pomijana)
	type preparedStep struct {
		stepNo            int
		approverType      string
		approverID        any   
		approverOrgUnitID *int
		approverRoleCode  *int
	}

	prepared := make([]preparedStep, 0, len(workflowDef.Steps))

	for _, step := range workflowDef.Steps {
		switch step.ApproverType {

		case "manager_of_requester":
			managerID, err := u.orgRepo.GetManagerForUser(ctx, requesterID)
			if err != nil || managerID == nil {
				// brak użytkownika → pomijamy krok
				continue
			}
			prepared = append(prepared, preparedStep{
				stepNo:       step.StepNo,
				approverType: step.ApproverType,
				approverID:   managerID,
			})

		case "manager_of_unit":
			unitID, ok := parseSlugInt(step.Slug)
			if !ok {
				// brak/niepoprawny slug → pomijamy krok
				continue
			}
			managerID, err := u.orgRepo.GetManagerForUnit(ctx, unitID)
			if err != nil || managerID == nil {
				// brak użytkownika → pomijamy krok
				continue
			}
			prepared = append(prepared, preparedStep{
				stepNo:       step.StepNo,
				approverType: step.ApproverType,
				approverID:   managerID,
			})

		case "specific_user":
			userID, ok := parseSlugInt(step.Slug)
			if !ok {
				continue
			}
			uid := userID
			prepared = append(prepared, preparedStep{
				stepNo:       step.StepNo,
				approverType: step.ApproverType,
				approverID:   &uid,
			})

		case "user_of_unit":
			unitID, ok := parseSlugInt(step.Slug)
			if !ok {
				continue
			}
			uid := unitID
			prepared = append(prepared, preparedStep{
				stepNo:            step.StepNo,
				approverType:      step.ApproverType,
				approverID:        nil,   // brak konkretnego usera
				approverOrgUnitID: &uid,
			})

		case "role_global":
			roleID, ok := parseSlugInt(step.Slug)
			if !ok {
				continue
			}
			rid := roleID
			prepared = append(prepared, preparedStep{
				stepNo:           step.StepNo,
				approverType:     step.ApproverType,
				approverID:       nil,
				approverRoleCode: &rid,
			})

		default:
			// Nieznany typ → pomijamy (żeby "nic się nie  ")
			continue
		}
	}

	// 3) Jeśli po pominięciach nie ma żadnych kroków → auto-approve
	if len(prepared) == 0 {
		now := time.Now()
		req := du.ApprovalRequest{
			Resource:  resource,
			TargetID:  targetID,
			Requester: requesterID,
			Status:    "approved",
			CreatedAt: now,
			ClosedAt:  &now,
		}
		if _, err := u.repo.CreateRequest(ctx, req); err != nil {
			return "", fmt.Errorf("nie udało się utworzyć auto-approved requesta (brak kroków po resolvowaniu): %w", err)
		}
		return "Workflow istnieje, ale nie udało się wyznaczyć żadnego akceptującego — wniosek automatycznie zaakceptowany", nil
	}

	// 4) Utwórz główny rekord jako pending
	req := du.ApprovalRequest{
		Resource:  resource,
		TargetID:  targetID,
		Requester: requesterID,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	requestID, err := u.repo.CreateRequest(ctx, req)
	if err != nil {
		return "", err
	}

	// 5) Zapisz kroki: pierwszy utworzony = pending, reszta = waiting
	first := true
	for _, ps := range prepared {
		status := "waiting"
		if first {
			status = "pending"
		}

		step := du.ApprovalStep{
			RequestID:           requestID,
			StepNo:              ps.stepNo,
			ApproverID:          nil,
			Status:              status,
			Approver_type:       ps.approverType,
			Approver_org_unitID: ps.approverOrgUnitID,
			Approver_role_code:  ps.approverRoleCode,
		}

		// zachowujemy Twoją logikę typów ApproverID (managerID bywa już wskaźnikiem / specific_user jest *int)
		if ps.approverID != nil {
			if v, ok := ps.approverID.(*int); ok {
				step.ApproverID = v
			} else {
		
				if vv, ok := ps.approverID.(interface{ }); ok {
					_ = vv // placeholder aby nie psuć kompilacji w zależności od faktycznego typu
				}
			}
		}

		if _, err := u.repo.CreateStep(ctx, step); err != nil {
			return "", err
		}
		first = false
	}

	return fmt.Sprintf("Utworzono wniosek: %s (%d kroki akceptacji)", workflowDef.Name, len(prepared)), nil
}
// -----------------------------
// DECISION (zatwierdzenie / odrzucenie kroku)
// -----------------------------
func (u *usecase) Decision(ctx context.Context, stepID int64, approverID int, decision string, note *string) (string, error) {
	if decision != "approved" && decision != "rejected" {
		return "", fmt.Errorf("błędna decyzja: %s", decision)
	}
	

	step, err := u.repo.GetStepByID(ctx, stepID) 
	if err != nil {
		return "", fmt.Errorf("nie znaleziono kroku: %w", err)
	}

	if step.Status != "pending" {
		return "", fmt.Errorf("krok został już rozpatrzony (status: %s)", step.Status)
	}

	
	// AUTORYZACJA 


	switch step.Approver_type {
		case "manager_of_requester":
			if step.ApproverID == nil || *step.ApproverID != approverID {
				return "", fmt.Errorf("brak uprawnień do akceptacji tego kroku")
			}



	if err := u.repo.UpdateStepStatus(ctx, stepID, decision, note); err != nil {
		return "", err
	}

	if decision == "rejected" {
		if err := u.repo.UpdateRequestStatus(ctx, step.RequestID, "rejected"); err != nil {
			return "", err
		}
		return "wniosek odrzucony", nil
	}


	nextStep, err := u.repo.GetNextStep(ctx, step.RequestID, step.StepNo)
	if err != nil {
		// brak kolejnego -> zakończ workflow
		if err := u.repo.UpdateRequestStatus(ctx, step.RequestID, "approved"); err != nil {
			return "", err
		}
		return "workflow zakończony – zaakceptowano", nil
	}

	if err := u.repo.UpdateStepStatusByID(ctx, nextStep.ID, "pending", nil); err != nil {
		return "", fmt.Errorf("nie udało się aktywować kolejnego kroku: %w", err)
	}









		case "role_global":

		if step.Approver_role_code == nil {
    				return "", fmt.Errorf("krok nie posiada przypisanej roli")
			}
			have, err := u.permissionsSvc.DoesUserHaveRole(ctx, approverID, *step.Approver_role_code)
			if err != nil || !have {
				return "", fmt.Errorf("brak uprawnień do akceptacji tego kroku")
			}


			if err := u.repo.UpdateStepStatusAndApproverID(ctx, stepID, decision, note, approverID); err != nil {
				return "", err
			}

	if decision == "rejected" {
		if err := u.repo.UpdateStepStatusAndApproverID(ctx, step.RequestID, "rejected", note, approverID); err != nil {
			return "", err
		}
		return "wniosek odrzucony", nil
	}


	nextStep, err := u.repo.GetNextStep(ctx, step.RequestID, step.StepNo)
	if err != nil {
		// brak kolejnego -> zakończ workflow
		if err := u.repo.UpdateRequestStatus(ctx, step.RequestID, "approved"); err != nil {
			return "", err
		}
		return "workflow zakończony – zaakceptowano", nil
	}

	if err := u.repo.UpdateStepStatusByID(ctx, nextStep.ID, "pending", nil); err != nil {
		return "", fmt.Errorf("nie udało się aktywować kolejnego kroku: %w", err)
	}






		case "manager_of_unit":
			if step.ApproverID == nil || *step.ApproverID != approverID {
				return "", fmt.Errorf("brak uprawnień do akceptacji tego kroku")
			}


	if err := u.repo.UpdateStepStatus(ctx, stepID, decision, note); err != nil {
		return "", err
	}

	if decision == "rejected" {
		if err := u.repo.UpdateRequestStatus(ctx, step.RequestID, "rejected"); err != nil {
			return "", err
		}
		return "wniosek odrzucony", nil
	}

	nextStep, err := u.repo.GetNextStep(ctx, step.RequestID, step.StepNo)
	if err != nil {
		// brak kolejnego -> zakończ workflow
		if err := u.repo.UpdateRequestStatus(ctx, step.RequestID, "approved"); err != nil {
			return "", err
		}
		return "workflow zakończony – zaakceptowano", nil
	}

	if err := u.repo.UpdateStepStatusByID(ctx, nextStep.ID, "pending", nil); err != nil {
		return "", fmt.Errorf("nie udało się aktywować kolejnego kroku: %w", err)
	}



		case "specific_user":
			if step.ApproverID == nil || *step.ApproverID != approverID {
				return "", fmt.Errorf("brak uprawnień do akceptacji tego kroku")
			}



	if err := u.repo.UpdateStepStatus(ctx, stepID, decision, note); err != nil {
		return "", err
	}

	if decision == "rejected" {
		if err := u.repo.UpdateRequestStatus(ctx, step.RequestID, "rejected"); err != nil {
			return "", err
		}
		return "wniosek odrzucony", nil
	}

	nextStep, err := u.repo.GetNextStep(ctx, step.RequestID, step.StepNo)
	if err != nil {
		// brak kolejnego -> zakończ workflow
		if err := u.repo.UpdateRequestStatus(ctx, step.RequestID, "approved"); err != nil {
			return "", err
		}
		return "workflow zakończony – zaakceptowano", nil
	}

	if err := u.repo.UpdateStepStatusByID(ctx, nextStep.ID, "pending", nil); err != nil {
		return "", fmt.Errorf("nie udało się aktywować kolejnego kroku: %w", err)
	}



		case "user_of_unit":

var unitInt int
// Zmień cały blok w case "user_of_unit":
if step.Approver_org_unitID != nil {
    unitInt = *step.Approver_org_unitID
} else {
    return "", fmt.Errorf("brak ID jednostki w kroku")
}




	is, err:= u.orgRepo.IsUserNowInUnit(ctx, approverID, unitInt)

	fmt.Println(is)
	fmt.Println(err)
			if err != nil || !is {
				return "", fmt.Errorf("brak uprawnień do akceptacji tego kroku")
			}


			if err := u.repo.UpdateStepStatusAndApproverID(ctx, stepID, decision, note, approverID); err != nil {
				return "", err
			}

	if decision == "rejected" {
		if err := u.repo.UpdateStepStatusAndApproverID(ctx, step.RequestID, "rejected", note, approverID); err != nil {
			return "", err
		}
		return "wniosek odrzucony", nil
	}

	nextStep, err := u.repo.GetNextStep(ctx, step.RequestID, step.StepNo)
	if err != nil {
		if err := u.repo.UpdateRequestStatus(ctx, step.RequestID, "approved"); err != nil {
			return "", err
		}
		return "workflow zakończony – zaakceptowano", nil
	}

	if err := u.repo.UpdateStepStatusByID(ctx, nextStep.ID, "pending", nil); err != nil {
		return "", fmt.Errorf("nie udało się aktywować kolejnego kroku: %w", err)
	}














		default:
		fmt.Printf("⚠️ nieznany typ approvera: %s\n", step.Approver_type)



	return "zatwierdzono krok – aktywowano następny", nil
}

	return "zatwierdzono krok – aktywowano następny", nil
}
// -----------------------------
// LIST PENDING (oczekujące do akceptacji dla usera)
// -----------------------------
func (u *usecase) ListPending(ctx context.Context, approverID int) ([]du.ApprovalStepList, error) {
	return u.repo.GetPendingStepsForApprover(ctx, approverID)
}

