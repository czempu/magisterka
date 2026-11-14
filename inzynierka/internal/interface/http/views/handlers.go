package handler

import (
    "net/http"
	"strconv"	
    "fmt"
    "github.com/labstack/echo/v4"
    "github.com/golang-jwt/jwt/v5"
    "erpbadger/internal/interface/http/middleware"

    "erpbadger/internal/usecase/auth"
    "erpbadger/internal/usecase/users"
)






func getUser(c echo.Context) int {
    idParam := c.Param("id")
    id, err := strconv.Atoi(idParam)
    if err != nil {
    }
    return id 

}

var p = fmt.Println

type Handlers struct {
authSvc auth.Usecase 
usersSvc users.Usecase
}

func New(authSvc auth.Usecase, usersSvc users.Usecase) *Handlers {
	return &Handlers{authSvc: authSvc, usersSvc: usersSvc}
}

///////////////////
//
// testy
//
///////////////////


func (h *Handlers) ApiHello(c echo.Context) error {  // system 
    userToken := c.Get("user").(*jwt.Token)
    claims := userToken.Claims.(jwt.MapClaims)
    username := claims["username"].(string)

   // sd := c.Get("session").(*middleware.SessionData)
    //fmt.Println(sd.Username)
 //fmt.Println(sd.UserID)

    return c.JSON(http.StatusOK, echo.Map{"message": "Hello, " + username + "!"})
}

func (h *Handlers) canAccess(c echo.Context, resource string) bool {

	targetuserid := c.Param("id")

	targetuseridint,_ := strconv.Atoi(targetuserid)
	

	scope,_, err := h.authSvc.MaxScopeFor(c.Request().Context(), resource)
	if err != nil {
		return false
	}

    switch scope {
    case "own":
        is, _ := h.authSvc.CheckIsYourself(c.Request().Context(), targetuseridint)
        return is
    case "team":
        is, _ := h.authSvc.CheckIsUserInMyDepartment(c.Request().Context(), targetuseridint)
        return is
    case "org":
        return true
    default:
        return false
    }
}

func (h *Handlers) renderWithAuth(c echo.Context, status int, view string, title string, data map[string]any) error {
    sd := c.Get("session").(*middleware.SessionData)

	id := c.Param("id")
	idint, _ := strconv.Atoi(id)

    perms, err := h.authSvc.GetAllPermissions(c.Request().Context())
    if err != nil {
        perms = nil
    }

    // Budujemy mapę uprawnień
    permissionsSet := make(map[string]bool)
    for _, p := range perms {
        key := fmt.Sprintf("%s:%s:%s", p.Resource, p.Action, p.Scope)
        permissionsSet[key] = true
    }

    PII, err := h.usersSvc.GetUserPII(c.Request().Context(), sd.UserID)
    if err != nil {
    }

    TargetPII, err := h.usersSvc.GetUserPII(c.Request().Context(), idint)
    if err != nil {
    }






    // Przygotowujemy bazowy kontekst danych
    viewData := map[string]any{
        "Title":          title,
	"User":           
	struct{ 
		ID int 
		Username string
		Name string
		Surname string
	}{ID: sd.UserID, Username: sd.Username, Name: PII.Name,Surname: PII.Surname},
	"TargetUser": 	  struct{ ID string
				Name string
				Surname string}{ID: id, Name: TargetPII.Name, Surname: TargetPII.Surname },
        "Permissions":    perms,
        "PermissionsSet": permissionsSet,
    }

    // Łączymy bazowe dane z dodatkowymi danymi przekazanymi do funkcji
    for k, v := range data {
        viewData[k] = v
    }

    return c.Render(status, view, viewData)
}

// Przykład handlera z dodatkowymi danymi:
//func (h *Handlers) UserProfilePage(c echo.Context) error {
//    extraData := map[string]any{
//        "ProfileInfo": "Jakieś dodatkowe dane",
//    }
//    return h.renderWithAuth(c, http.StatusOK, "profile.html", "Mój Profil", extraData)
//}


///////////////////////
//
// widoki do przeniesienia całe handler views
//
///////////////////////


func (h *Handlers) DashboardPage(c echo.Context) error {
return h.renderWithAuth(c, http.StatusOK, "dashboard.html", "Dashboard", nil)
}

func (h *Handlers) LeaveBalancePage(c echo.Context) error {
return h.renderWithAuth(c, http.StatusOK, "leave_balance.html", "Bilans Urlopów", nil)    
}

func (h *Handlers) SickLeavePage(c echo.Context) error {
return h.renderWithAuth(c, http.StatusOK, "sick_leave_add.html", "Choroby", nil)   
}

func (h *Handlers) TimePolicyPage(c echo.Context) error {
return h.renderWithAuth(c, http.StatusOK, "time_policies.html", "Polityka Czasu Pracy", nil)
}
func (h *Handlers) KabanPage(c echo.Context) error {
return h.renderWithAuth(c, http.StatusOK, "kaban.html", "Kaban", nil)
}



func (h *Handlers) PermissionsView(c echo.Context) error {
	return h.renderWithAuth(c, http.StatusOK, "permissions.html", "Uprawnienia", nil)
}
func (h *Handlers) OrganizationView(c echo.Context) error {
	return h.renderWithAuth(c, http.StatusOK, "organization.html", "Organizacja", nil)
}


func (h *Handlers) ApprovalsPage(c echo.Context) error {
	return h.renderWithAuth(c, http.StatusOK, "approvals.html", "Akceptacje", nil)
}


func (h *Handlers) WorkflowPage(c echo.Context) error {
	return h.renderWithAuth(c, http.StatusOK, "workflow_admin.html", "Obiegi Akceptacji", nil)
}

func (h *Handlers)  UsersPage(c echo.Context) error {
	return h.renderWithAuth(c, http.StatusOK, "users.html", "Użytkownicy", nil)
}

func (h *Handlers)  PasswordChangePage(c echo.Context) error {
	return h.renderWithAuth(c, http.StatusOK, "passwordchange.html", "Zmiana Hasła", nil)
}







//////////////////////////////////////////////////

func (h *Handlers) ScheduleUserPage(c echo.Context) error {

	if h.canAccess(c, "schedule:read") {
	return h.renderWithAuth(c, http.StatusOK, "schedule_user2.html", "Grafik", nil)
	}
	 return c.Redirect(http.StatusSeeOther, "/app/unauthorized") 

}

func (h *Handlers) ScheduleUserPageMe(c echo.Context) error {
	return h.renderWithAuth(c, http.StatusOK, "schedule_user2.html", "Grafik", nil)
}


func (h *Handlers) UserRolePage(c echo.Context) error {
return h.renderWithAuth(c, http.StatusOK, "userrole.html", "Rola Użytkownika", nil)
}

func (h *Handlers) UserEmploymentPage(c echo.Context) error {

	if h.canAccess(c, "schedule:read") {
		return h.renderWithAuth(c, http.StatusOK, "user_employment.html", "Zatrudnienie Użytkownika", nil)
	}
	return c.Redirect(http.StatusSeeOther, "/app/unauthorized") 

}

func (h *Handlers) UserScheduleWeeklyPage(c echo.Context) error {
	week := c.Param("week")
	extraData := map[string]any{
        	"Week": struct{ WEEK string }{WEEK: week},
    	}
	return h.renderWithAuth(c, http.StatusOK, "user_schedule_weekly.html", "Plan Tygodniowy Użytkownika", extraData)
}



func (h *Handlers) UserScheduleWeeksPage(c echo.Context) error {


	if h.canAccess(c, "schedule:read") {
return h.renderWithAuth(c, http.StatusOK, "user_schedule_weeks.html", "Tygodniowe Planowanie Użytkownika", nil)
}
	 return c.Redirect(http.StatusSeeOther, "/app/unauthorized") 




}
func (h *Handlers) UserScheduleOverridesPage(c echo.Context) error {
	if h.canAccess(c, "schedule:read") {
		return h.renderWithAuth(c, http.StatusOK, "user_schedule_overrides.html", "Dni Wyjątki Użytkownika", nil)
	}
	return c.Redirect(http.StatusSeeOther, "/app/unauthorized") 
}




func (h *Handlers) UserDetailsPage(c echo.Context) error {
return h.renderWithAuth(c, http.StatusOK, "userdetails.html", "Edycja Użytkownika", nil)
}

func (h *Handlers) LeavesInfoPage(c echo.Context) error {

	idleave := c.Param("idleave")
	extraData := map[string]any{
		"Leave": struct{ ID string }{ID: idleave},
    	}

	return h.renderWithAuth(c, http.StatusOK, "leavesinfo.html", "Urlopy Szczegóły", extraData)
}

func (h *Handlers) LeavesPage(c echo.Context) error {
return h.renderWithAuth(c, http.StatusOK, "leaves.html", "Urlopy", nil)
}

func (h *Handlers) LeavesAddPage(c echo.Context) error {
return h.renderWithAuth(c, http.StatusOK, "leavesadd.html", "Dodawanie Wniosku Urlopowego", nil)
}










/////////////////////////////////////////////
//
//
//
//
// sekcja publiczna
//
//
//
//
//
//////////////////////////////////////////////





func (h *Handlers) UnauthorizedView(c echo.Context) error {
	return c.File("templates/unauthorized.html")
}


func (h *Handlers) NotFoundPage(c echo.Context) error {
	return c.File("templates/notfound.html")
}

//syf
func (h *Handlers) SettingsSystemPage(c echo.Context) error {
	return c.File("templates/settings-system.html")
}  


//////////////////////////////////////////////////
func (h *Handlers) TestPage(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", map[string]any{
			"Title": "Test",
		})
}


func (h *Handlers) HomePage(c echo.Context) error {
		fmt.Println("home page")
		return c.Render(http.StatusOK, "index.html", map[string]any{
			"Title": "Strona główna",
		})
}

func (h *Handlers) LoginPage(c echo.Context) error {
		return c.Render(http.StatusOK, "login.html", map[string]any{
			"Title": "login",
		})
}


func (h *Handlers) DevModePage(c echo.Context) error {
		return c.Render(http.StatusOK, "indevelopment.html", map[string]any{
			"Title": "WTR Hub - dev mode",
		})
}

//////////////////////
//
// koniec widoków
//
//////////////////////


