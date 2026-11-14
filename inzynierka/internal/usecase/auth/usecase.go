package auth

import (
    "context"
    "errors"
	"fmt"
	"strings"

	du "erpbadger/internal/domain/auth"
	"erpbadger/internal/domain/security"
	"golang.org/x/crypto/bcrypt"
)

var p = fmt.Print

type Usecase interface {

  // sprawdza user+pass, zwraca userID lub błąd
    CheckUserPassword(ctx context.Context, username, password string) (int, error)
CheckPermission(ctx context.Context, required ...string) (*du.PermCtx, error)
    // NOWA METODA – lista perms dla dashboardu, widoków itd.
    GetAllPermissions(ctx context.Context) ([]du.EffectivePermission, error)
    MaxScopeFor(ctx context.Context, required string) (string,int, error)
CheckIsUserInMyDepartment(ctx context.Context, targetUserID int) (bool, error)
CheckIsYourself(ctx context.Context, targetUserID int)	(bool, error)
CheckUserPasswordById(ctx context.Context, id int, password string) (int, error)
}

type usecase struct {
	repo   du.Repository
	hasher security.PasswordHasher
}

// NewUsecase – wstrzykujesz tutaj repo oraz hasher
func NewUsecase(r du.Repository, hasher security.PasswordHasher) Usecase {
	return &usecase{repo: r, hasher: hasher}
}


// ===========================
// Permission Checking (NEW)
// przeniesione z middleware
// ===========================

func (u *usecase) CheckIsYourself(ctx context.Context, targetUserID int) (bool, error) {
	// 1. Pobieramy ID użytkownika wykonującego zapytanie (observer / "1") z kontekstu
	uid, ok := ctx.Value("userID").(int)
	if !ok || uid == 0 {
		return false, errors.New("Brak Autoryzacji")
	}

	// Opcjonalnie: Jeśli sprawdzamy samego siebie, od razu zwracamy true
	if uid == targetUserID {
		return true, nil
	}
		return false, errors.New("Brak Uprawnień")
}



func (u *usecase) CheckIsUserInMyDepartment(ctx context.Context, targetUserID int) (bool, error) {
	// 1. Pobieramy ID użytkownika wykonującego zapytanie (observer / "1") z kontekstu
	uid, ok := ctx.Value("userID").(int)
	if !ok || uid == 0 {
		return false, errors.New("unauthorized")
	}

	// Opcjonalnie: Jeśli sprawdzamy samego siebie, od razu zwracamy true
	if uid == targetUserID {
		return true, nil
	}

	// 2. Wywołujemy repozytorium z Twoim SQLem
	// uid = user odpytujący (1)
	// targetUserID = user sprawdzany (4)
	isSameOrg, err := u.repo.IsUserInSamePrimaryOrg(ctx, uid, targetUserID)
	if err != nil {
		// Tutaj warto zalogować prawdziwy błąd bazy danych
		return false, errors.New("check department failed")
	}

	return isSameOrg, nil
}
// CheckPermission sprawdza, czy user z ctx ma dostęp do required perms
func (u *usecase) CheckPermission(ctx context.Context, required ...string) (*du.PermCtx, error) {
	// userID z kontekstu — ustawiany przez RequireSession()
	uid, ok := ctx.Value("userID").(int)
	if !ok || uid == 0 {
		return nil, errors.New("unauthorized")
	}

	perms, err := u.repo.ForUser(ctx, uid)
	if err != nil {
		return nil, errors.New("permission load failed")
	}

	order := map[string]int{"own": 1, "team": 2, "org": 3}

	parse := func(t string) (string, string, string) {
		parts := strings.Split(t, ":")
		if len(parts) != 3 {
			return t, "*", "own"
		}
		return parts[0], parts[1], parts[2]
	}

	var chosen du.EffectivePermission
	best := 0
	okFound := false

	for _, r := range required {
		res, act, sc := parse(r)
		need := order[sc]

		for _, ep := range perms {
			if (ep.Resource == res || ep.Resource == "*") &&
				(ep.Action == act || ep.Action == "*") {

				if g := order[ep.Scope]; g >= need && g > best {
					best = g
					chosen = ep
					okFound = true
				}
			}
		}
	}

	if !okFound {
		return nil, errors.New("forbidden")
	}

	return &du.PermCtx{
		Scope:     chosen.Scope,
		Condition: chosen.Condition,
	}, nil
}

func (u *usecase) MaxScopeFor(ctx context.Context, required string) (string,int, error) {
	// userID z kontekstu — ustawiany przez RequireSession()
	uid, ok := ctx.Value("userID").(int)
	if !ok || uid == 0 {
		return "",0, errors.New("unauthorized")
	}

	perms, err := u.repo.ForUser(ctx, uid)
	if err != nil {
		return "",0, errors.New("permission load failed")
	}

	order := map[string]int{"own": 1, "team": 2, "org": 3}

	parse := func(t string) (string, string, error) {
		parts := strings.Split(t, ":")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", errors.New("invalid required permission format (expected resource:action)")
		}
		return parts[0], parts[1], nil
	}

	// Tie-breaker: preferuj bardziej konkretne dopasowania (bez wildcardów)
	specificity := func(ep du.EffectivePermission, res, act string) int {
		s := 0
		if ep.Resource != "*" && ep.Resource == res {
			s++
		}
		if ep.Action != "*" && ep.Action == act {
			s++
		}
		return s
	}

	res, act, err := parse(required)
	if err != nil {
		return "",0, err
	}

	bestScope := 0
	bestSpec := -1
	found := false

	for _, ep := range perms {
		if (ep.Resource == res || ep.Resource == "*") &&
			(ep.Action == act || ep.Action == "*") {

			g := order[ep.Scope]
			if g == 0 {
				continue // nieznany scope
			}

			spec := specificity(ep, res, act)

			if g > bestScope || (g == bestScope && spec > bestSpec) {
				bestScope = g
				bestSpec = spec
				found = true
			}
		}
	}

	if !found {
		return "",0, errors.New("forbidden")
	}

	// mapowanie int -> string
	switch bestScope {
	case 3:
		return "org",uid, nil
	case 2:
		return "team",uid, nil
	default:
		return "own",uid, nil
	}
}



func (u *usecase) checkPassword(hash, password string) bool {
	if u.hasher != nil {
		return u.hasher.Compare(hash, password) == nil
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// CheckUserPassword – logika: pobierz encję, porównaj hasła
func (u *usecase) CheckUserPassword(ctx context.Context, username, password string) (int, error) {
	usr, err := u.repo.GetByUsername(ctx, username)
	if err != nil {
		return 0, err
	}

	if u.checkPassword(usr.Password, password) {
		return usr.ID, nil
	}
	return 0, errors.New("invalid credentials")
}

func (u *usecase) CheckUserPasswordById(ctx context.Context, id int, password string) (int, error) {
	usr, err := u.repo.GetById(ctx, id)
	if err != nil {
		return 0, err
	}

	if u.checkPassword(usr.Password, password) {
		return usr.ID, nil
	}
	return 0, errors.New("invalid credentials")
}


func (u *usecase) GetAllPermissions(ctx context.Context) ([]du.EffectivePermission, error) {
    uid, ok := ctx.Value("userID").(int)
    if !ok || uid == 0 {
        return nil, errors.New("unauthorized")
    }

    return u.repo.ForUser(ctx, uid)
}

