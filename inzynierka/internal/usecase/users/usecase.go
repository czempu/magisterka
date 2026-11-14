package users

import (
	"context"
	"fmt"

	"erpbadger/internal/domain/mailer"
	"erpbadger/internal/domain/security"
	du "erpbadger/internal/domain/users"
	auth "erpbadger/internal/usecase/auth"
)

var p = fmt.Print



// Usecase – interfejs dla handlera/middleware
type Usecase interface {
	GetUserPII(ctx context.Context, userID int) (du.PII, error)
	UserDetails(ctx context.Context, userID int) (du.UserDetails,error)
	UserDetailsEdit(ctx context.Context, userID int, userd du.UserDetails) (string, error)
      AddUser(ctx context.Context, username, email string) (string, error)
DeleteUser(ctx context.Context, userID int) (string, error)
 GetUsersInfo(ctx context.Context) ([]du.UsersInfo, error)
ResetPassword(ctx context.Context, id int) (string,error)
PasswordChange(ctx context.Context, id int, OldPassword,NewPassword string) (error)
 UserExist(ctx context.Context,username, email string, id *int) (bool, error) 
}

type usecase struct {
	repo    du.Repository
	authSvc auth.Usecase
	mailer  mailer.Mailer
	hasher  security.PasswordHasher
	pwdGen  security.PasswordGenerator
}

// NewUsecase – wstrzykujesz repo, auth, mailer oraz komponenty bezpieczeństwa
func NewUsecase(
	r du.Repository,
	au auth.Usecase,
	m mailer.Mailer,
	hasher security.PasswordHasher,
	pwdGen security.PasswordGenerator,
) Usecase {
	return &usecase{
		repo:    r,
		authSvc: au,
		mailer:  m,
		hasher:  hasher,
		pwdGen:  pwdGen,
	}
}

func (u *usecase) DeleteUser(ctx context.Context, userid int) (string, error) {
	info, err := u.repo.DeleteUser(ctx, userid)
	if err != nil {
		return "", err
	}
	return info, err
}

func (u *usecase) AddUser(ctx context.Context, username, email string) (string, error) {
	exist, eerr := u.repo.UserExists(ctx, username, email, nil)
	if eerr != nil {
		return "", eerr
	}
	if exist {
		return "", fmt.Errorf("Użytkownik już istnieje!")
	}

	password, err := u.pwdGen.Generate(8)
	if err != nil {
		return "", fmt.Errorf("błąd generowania hasła: %w", err)
	}

	hash, err := u.hasher.Hash(password)
	if err != nil {
		return "", fmt.Errorf("błąd hashowania hasła: %w", err)
	}

	if err := u.repo.AddUser(ctx, username, email, hash); err != nil {
		return "", err
	}

	msg := "Witamy w systemie ERPBadger: <b>" + username + "</b>! <br><br> Twoje hasło to: <b>" + password + "</b><br><br><br> Po pierwszym logowaniu zalecamy zmianę hasła w profilu użytkownika. <br><br> Z wyrazami szacunku, <br><br> Zespół ERPBadger."

	if u.mailer != nil {
		_ = u.mailer.SendMail(ctx, "ERPBadger <noreply@erpbadger.com>", []string{email}, msg, "Rejestracja w systemie ERPBadger")
	}

	return username, nil
}

func (u *usecase) PasswordChange(ctx context.Context, id int, oldpass, password string) error {
	userDetails, userDetailsErr := u.repo.UserDetails(ctx, id)
	if userDetailsErr != nil {
		return userDetailsErr
	}

	is, erris := u.authSvc.CheckIsYourself(ctx, id)
	if !is {
		return fmt.Errorf("Użytkownik może zmieniać tylko własne hasło!")
	}
	if erris != nil {
		return fmt.Errorf("Nie powiodła sie walidacja użytkownika")
	}

	_, erris2 := u.authSvc.CheckUserPasswordById(ctx, id, oldpass)
	if erris2 != nil {
		return fmt.Errorf("stare hasło jest nieprawidłowe")
	}

	hash, err := u.hasher.Hash(password)
	if err != nil {
		return fmt.Errorf("błąd hashowania nowego hasła: %w", err)
	}

	err = u.repo.PasswordChange(ctx, id, hash)
	if err != nil {
		return err
	}

	msg := "Wykryliśmy zmianę hasła, jeśli to nie ty je zmieniałeś skontaktuj sie z administratorem. <br><br> Z wyrazami szacunku, <br><br> Zespół ERPBadger."

	if u.mailer != nil {
		_ = u.mailer.SendMail(ctx, "ERPBadger <noreply@erpbadger.com>", []string{userDetails.Email}, msg, "Zmiana hasła w systemie ERPBadger")
	}

	return nil
}

func (u *usecase) ResetPassword(ctx context.Context, id int) (string, error) {
	userDetails, userDetailsErr := u.repo.UserDetails(ctx, id)
	if userDetailsErr != nil {
		return "Błąd podczas pobierania informacji o użytkowniku.", userDetailsErr
	}

	password, err := u.pwdGen.Generate(8)
	if err != nil {
		return "Błąd przy generowaniu hasła", err
	}

	hash, err := u.hasher.Hash(password)
	if err != nil {
		return "Błąd przy hashowaniu hasła", err
	}

	if err := u.repo.PasswordChange(ctx, id, hash); err != nil {
		return "Błąd przy resetowaniu hasła", err
	}

	msg := "Administrator zresetował twoje hasło. <br><br><br> twoje wygenerowane hasło to: <b>" + password + "</b> <br><br> Z wyrazami szacunku, <br><br> Zespół ERPBadger."

	if u.mailer != nil {
		_ = u.mailer.SendMail(ctx, "ERPBadger <noreply@erpbadger.com>", []string{userDetails.Email}, msg, "Zresetowane hasło - ERPBadger")
	}

	return "Hasło zostało zaktualizowane", nil
}

func (u *usecase) UserDetailsEdit(ctx context.Context, userID int, userd du.UserDetails) (string, error) {


	fmt.Println(userd.ID)
 	exist, eerr := u.repo.UserExists(ctx, userd.Username, userd.Email, &userd.ID) 
	if eerr != nil {
		return "", eerr
	}
	if exist {
		return "", fmt.Errorf("Użytkownik już istnieje!")
	}

	err := u.repo.UserDetailsEdit(ctx, userID, userd)
	
	if err != nil {
		return "", err
	}

	return "zaktualizowano", nil
}





func (u *usecase) GetUserPII(ctx context.Context, userID int) (du.PII, error) {
	data, err := u.repo.GetUserPII(ctx, userID)
	return data, err
}


func (u *usecase) UserExist(ctx context.Context,username, email string, id *int) (bool, error) {
	exist, err := u.repo.UserExists(ctx, username, email, id)
	return exist, err
}







func (u *usecase) UserDetails(ctx context.Context, userID int) (du.UserDetails, error) {
	UD, _ := u.repo.UserDetails(ctx, userID)
	
	p(UD)

	return UD, nil
}




func (u *usecase) GetUsersInfo(ctx context.Context) ([]du.UsersInfo, error) {

	scope,userid, err := u.authSvc.MaxScopeFor(ctx, "users:read")
	if err != nil {
		return nil,fmt.Errorf("Brak Uprawnień")
	}

	return u.repo.GetUsersInfo(ctx, scope, userid)
}














