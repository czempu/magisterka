package server

import (
	"net/http"
	"strings"
	"erpbadger/internal/interface/http/middleware"
	"erpbadger/internal/domain/whitelist"
	authhttp "erpbadger/internal/interface/http/auth"
	leaveshttp "erpbadger/internal/interface/http/leaves"
	moduleshttp "erpbadger/internal/interface/http/modules"
	permissionshttp "erpbadger/internal/interface/http/permissions"
	rcphttp "erpbadger/internal/interface/http/rcp"
	usershttp "erpbadger/internal/interface/http/users"
	viewshttp "erpbadger/internal/interface/http/views"
	organizationhttp "erpbadger/internal/interface/http/organization"
	approvalhttp "erpbadger/internal/interface/http/approval"
	workflowhttp "erpbadger/internal/interface/http/workflow"
	employmenthttp "erpbadger/internal/interface/http/employment"
	schedulehttp "erpbadger/internal/interface/http/schedule"
	statshttp "erpbadger/internal/interface/http/stats"
	sicknesshttp "erpbadger/internal/interface/http/sickness"
	kanbanhttp "erpbadger/internal/interface/http/kanban"
	i18nhttp "erpbadger/internal/interface/http/i18n"
	"erpbadger/internal/usecase/auth"
	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
	"erpbadger/config"
)

type Modules struct {
	Auth *authhttp.Handlers
	Leaves *leaveshttp.Handlers
	Modules *moduleshttp.Handlers
	Permissions *permissionshttp.Handlers
	Rcp *rcphttp.Handlers
	Users *usershttp.Handlers
	Views *viewshttp.Handlers
	Organization *organizationhttp.Handlers
	Approval *approvalhttp.Handlers
	Workflow *workflowhttp.Handlers
	Employment  *employmenthttp.Handlers
    	Schedule    *schedulehttp.Handlers
	Stats	*statshttp.Handlers
	Sickness *sicknesshttp.Handlers
	Kanban *kanbanhttp.Handlers
	I18n *i18nhttp.Handlers
}


func RegisterRoutes(e *echo.Echo, m Modules, ipRepo whitelist.Repository, authUC auth.Usecase, jwtSecret []byte) {

	//
	//
	// sekcja endpointów publicznych
	//
	//

	e.Static("/public", "public")
	if !config.DevMode {
		e.Pre(mw.HTTPSWWWRedirect())
	}
	e.Use(middleware.IPFilterMiddleware(ipRepo))
	e.GET("/indevelopment", m.Views.DevModePage)
	e.GET("/", m.Views.HomePage)
	e.GET("home", m.Views.HomePage)
	e.GET("login", m.Views.LoginPage)
	e.GET("logout", m.Auth.Logout)
	e.GET("/notfound", m.Views.NotFoundPage)

	// i18n API
	if m.I18n != nil {
		e.GET("/api/i18n/languages", m.I18n.GetLanguages)
		e.GET("/api/i18n/:lang", m.I18n.GetTranslations)
	}




e.HTTPErrorHandler = func(err error, c echo.Context) {
	// rozpoznaj 404
	if he, ok := err.(*echo.HTTPError); ok && he.Code == http.StatusNotFound {
		path := c.Request().URL.Path

		// /api -> normalny 404 JSON
		if strings.HasPrefix(path, "/api") {
			e.DefaultHTTPErrorHandler(err, c)
			return
		}

		// poza /api -> redirect na /
		if path != "/" {
			_ = c.Redirect(http.StatusFound, "/notfound")
			return
		}
	}

	// pozostałe błędy standardowo
	e.DefaultHTTPErrorHandler(err, c)
}






	auth := e.Group("/auth")
	auth.POST("/login/token", m.Auth.LoginToken)
	auth.POST("/login/cookie", m.Auth.LoginCookie)

	// zmienna decydując czy dla braku uprawnień (widok) przekierowanie na stronę z informacją o braku uprawnień (rest api) odpowiedź JSON 
	isView := true

	//
	// middleware.Authz()
	//
	// funkcja sprawdzająca czy użytkownik próbujący dostać sie do zasobu posiada wymagane uprawnienia
	//
	//


	view := e.Group("/app")
	view.Use(middleware.AuthJWTWeb(jwtSecret), middleware.RequireSessionWeb())
	view.GET("/dashboard" , m.Views.DashboardPage, middleware.Authz(authUC, isView, "dashboard:read:own"),)
	view.GET("/users",m.Views.UsersPage, middleware.Authz(authUC, isView, "leaves:read:own"),)
	view.GET("/leaves", m.Views.LeavesPage, middleware.Authz(authUC, isView, "leaves:read:own"),)
	view.GET("/leaves/:idleave", m.Views.LeavesInfoPage, middleware.Authz(authUC, isView, "leaves:read:own"),)
	view.GET("/leaves/add", m.Views.LeavesAddPage, middleware.Authz(authUC, isView, "leaves:update:own"),)
	view.GET("/organization" , m.Views.OrganizationView, middleware.Authz(authUC, isView, "organization:read:own"),)
	view.GET("/approvals" , m.Views.ApprovalsPage, middleware.Authz(authUC, isView, "approvals:read:own"),)
 	view.GET("/schedule" , m.Views.ScheduleUserPageMe, middleware.Authz(authUC, isView, "schedule:read:own"),)
	view.GET("/user/:id/schedule" , m.Views.ScheduleUserPage, middleware.Authz(authUC, isView, "schedule:read:team"),)
	view.GET("/user/:id/role" , m.Views.UserRolePage, middleware.Authz(authUC, isView, "permissions:read:own"),) 
	view.GET("/user/:id/details" , m.Views.UserDetailsPage, middleware.Authz(authUC, isView, "users:read:own"),)
	view.GET("/sick" , m.Views.SickLeavePage, middleware.Authz(authUC, isView, "sick:read:own"),)
	view.GET("/user/:id/employment" , m.Views.UserEmploymentPage, middleware.Authz(authUC, isView, "users:read:own"),)
	view.GET("/user/:id/schedule/week/:week" , m.Views.UserScheduleWeeklyPage, middleware.Authz(authUC, isView, "schedule:read:own"),)
	view.GET("/user/:id/schedule/weeks" , m.Views.UserScheduleWeeksPage, middleware.Authz(authUC, isView, "schedule:read:own"),)
	view.GET("/user/:id/schedule/overrides" , m.Views.UserScheduleOverridesPage, middleware.Authz(authUC, isView, "schedule:read:own"),)
	view.GET("/permissions" , m.Views.PermissionsView, middleware.Authz(authUC, isView, "permissions:read:org"),)
	view.GET("/unauthorized" , m.Views.UnauthorizedView,)
	view.GET("/leavebalance" , m.Views.LeaveBalancePage, middleware.Authz(authUC, isView, "leavebalance:read:own"),)
	view.GET("/timepolicy" , m.Views.TimePolicyPage, middleware.Authz(authUC, isView, "timepolicy:read:org"),)
	view.GET("/workflow", m.Views.WorkflowPage, middleware.Authz(authUC, isView, "workflow:read:own"),)
		view.GET("/passwordchange", m.Views.PasswordChangePage,)
	view.GET("/kaban", m.Views.KabanPage, middleware.Authz(authUC, isView, "dashboard:read:own"),)

	//
	//
	//  sekcja API
	// 
	//
	

	isView = false
	api := e.Group("/api")
	api.Use(middleware.AuthJWT(jwtSecret),  middleware.RequireSession() )
	api.GET("/hello", m.Views.ApiHello, middleware.Authz(authUC,isView, "*:*:org"),)
	api.GET("/users", m.Users.Users, middleware.Authz(authUC,isView, "users:read:own"))
	api.GET("/request", m.Leaves.Request, middleware.Authz(authUC,isView, "leaves:read:own"))
	api.GET("/request/:id", m.Leaves.RequestInfo, middleware.Authz(authUC,isView, "leaves:read:own"))
	api.GET("/leave-types", m.Leaves.LeaveTypes, middleware.Authz(authUC,isView, "leaves:read:own"))
	api.POST("/request/add", m.Leaves.RequestAdd , middleware.Authz(authUC,isView, "leaves:update:own"))


	//raporty live - dashboard
	stats := api.Group("/stats")
	stats.GET("/employees/chart", m.Stats.EmployeeStatsChart)

	//api.GET("/cron/rcp/daily-totals", m.Rcp.RecalculateDailyTotals)


	workflow := api.Group("/workflow")
	workflow.GET("", m.Workflow.Workflows, middleware.Authz(authUC,isView, "workflow:read:org"))
	workflow.GET("/id/:id", m.Workflow.GetWorkflowByID, middleware.Authz(authUC,isView, "workflow:read:org"))
	workflow.GET("/:resource", m.Workflow.GetWorkflowForResource, middleware.Authz(authUC,isView, "workflow:read:org"))
	workflow.POST("", m.Workflow.CreateWorkflow, middleware.Authz(authUC,isView, "workflow:update:org"))
	workflow.PATCH("/:id/steps", m.Workflow.UpdateWorkflowSteps, middleware.Authz(authUC,isView, "workflow:update:org"))

	// organizacje
	org := api.Group("/org")
	org.GET("", m.Organization.OrgUnits, middleware.Authz(authUC,isView, "organization:read:own"))
	org.POST("/add", m.Organization.AddOrgUnit, middleware.Authz(authUC,isView, "organization:update:org"))
	org.POST("/assign", m.Organization.AssignUser, middleware.Authz(authUC,isView, "organization:update:org"))
	org.POST("/manager", m.Organization.SetManager, middleware.Authz(authUC,isView, "organization:update:org"))
	org.GET("/manager/:userid", m.Organization.GetManagerForUser, middleware.Authz(authUC,isView, "organization:read:own"))
	org.POST("/update", m.Organization.UpdateOrgUnit, middleware.Authz(authUC,isView, "organization:update:org"))
	org.GET("/manager/unit/:orgunitid", m.Organization.GetManagerForUnit, middleware.Authz(authUC,isView, "organization:read:own"))
	org.GET("/user-memberships", m.Organization.ListUserMemberships, middleware.Authz(authUC,isView, "organization:read:own"))
	org.DELETE("/:orgid", m.Organization.DeleteOrg, middleware.Authz(authUC,isView, "organization:update:org"))

	system := api.Group("/system")
	system.GET("/modules", m.Modules.SystemModules)
	system.GET("/roles", m.Permissions.SystemRoles, middleware.Authz(authUC,isView, "permissions:read:own"),)

	me := api.Group("/me")
	me.GET("/modules", m.Modules.MeModules)
	me.POST("/work/start", m.Rcp.MeWorkStart, middleware.Authz(authUC,isView, "schedule:time:own"),)
	me.POST("/work/stop", m.Rcp.MeWorkStop, middleware.Authz(authUC,isView, "schedule:time:own"),)
	me.GET("/work/status", m.Rcp.MeWorkStatus, middleware.Authz(authUC,isView, "schedule:read:own"),)
	me.GET("/work/timesheet", m.Rcp.MeWorkTimesheet, middleware.Authz(authUC,isView, "schedule:read:own"),)	

	approval := api.Group("/approval")
	//approval.POST("/create", m.Approval.Create, middleware.Authz(authUC,isView, "approvals:update:own"),)
	approval.POST("/:id/decision", m.Approval.Decision, middleware.Authz(authUC,isView, "approvals:update:org"),)
	approval.GET("/pending/:approver", m.Approval.Pending, middleware.Authz(authUC,isView, "approvals:read:org"),)

	user := api.Group("/user")
	user.POST("/add", m.Users.AddUser)
	user.POST("/:id/delete", m.Users.DeleteUser)
	user.GET("/:id/modules", m.Modules.UserModules)
	user.GET("/:id/role", m.Permissions.UserRole, middleware.Authz(authUC,isView, "permissions:read:own"))
	user.GET("/:id/details",m.Users.UserDetails)
	user.POST("/:id/role/change", m.Permissions.UserRoleChange, middleware.Authz(authUC,isView, "permissions:update:org"))
	user.POST("/:id/details/edit", m.Users.UserDetailsEdit, middleware.Authz(authUC,isView, "users:update:own"))
	user.GET("/:id/pii",m.Users.GetUserPII, middleware.Authz(authUC,isView, "users:read:own"))

	roles := api.Group("/roles")
	roles.GET("", m.Permissions.RolesList, middleware.Authz(authUC,isView, "permissions:read:own"))
	roles.GET("/:id",m.Permissions.RoleGet, middleware.Authz(authUC,isView, "permissions:read:own"))
	roles.POST("", m.Permissions.RoleCreate, middleware.Authz(authUC,isView, "permissions:update:org"))
	roles.PUT("/:id", m.Permissions.RoleUpdate, middleware.Authz(authUC,isView, "permissions:update:org"))
	roles.DELETE("/:id", m.Permissions.RoleDelete, middleware.Authz(authUC,isView, "permissions:update:org"))


	// --- permissions --- uprawnienia ok
	api.GET("/permissions", m.Permissions.PermissionsList, middleware.Authz(authUC,isView, "permissions:read:own")) // endpoint zwraca wszystkie uprawnienia
	api.GET("/permissions/:id", m.Permissions.PermissionGet, middleware.Authz(authUC,isView, "permissions:read:own")) // endpoint zwraca uprawnienia o wybranym id
	api.GET("/roles/:roleId/permissions",  m.Permissions.RolePermissionsList, middleware.Authz(authUC,isView, "permissions:read:own")) // endpoint zwraca uprawnienia dla roli o danym id

	api.POST("/user/:id/changepassword",  m.Users.PasswordChange) // endpoint zmiany hasła użytkownika

		api.POST("/user/:id/resetpassword",  m.Users.PasswordReset,middleware.Authz(authUC,isView, "users:update:org")) // endpoint zmiany hasła użytkownika



	// uprawnienia router + widoki
	// --- role <-> permission ---
	api.POST("/roles/:roleId/permissions/:permId", m.Permissions.RoleAssignPermission, middleware.Authz(authUC,isView, "permissions:update:org"))
	api.DELETE("/roles/:roleId/permissions/:permId", m.Permissions.RoleRemovePermission, middleware.Authz(authUC,isView, "permissions:update:org"))
	// --- user <-> role ---
	api.POST("/users/:userId/roles/:roleId", m.Permissions.UserAssignRole, middleware.Authz(authUC,isView, "permissions:update:org"))
	api.DELETE("/users/:userId/roles/:roleId", m.Permissions.UserRemoveRole, middleware.Authz(authUC,isView, "permissions:update:org"))



	// EMPLOYMENT - wymiar godzinowy / etat
	emp := api.Group("/user/:id/employment")
	emp.GET("", m.Employment.ListTerms, middleware.Authz(authUC,isView, "users:read:own"))       
	emp.POST("", m.Employment.AddTerm, middleware.Authz(authUC,isView, "users:update:org"))            // dodanie nowego okresu zatrudnienia
	emp.PUT("/:term_id", m.Employment.UpdateTerm, middleware.Authz(authUC,isView, "users:update:org")) // aktualizacja konkretnego okresu
	api.DELETE("/employment/:term_id", m.Employment.DeleteTerm, middleware.Authz(authUC,isView, "users:update:org"))






	api.GET("/time-policies", m.Schedule.ListTimePolicies, middleware.Authz(authUC,isView, "timepolicy:read:org"))
	api.POST("/time-policies", m.Schedule.SaveTimePolicy, middleware.Authz(authUC,isView, "timepolicy:update:org"))
	api.GET("/time-policies/active", m.Schedule.GetActiveTimePolicy, middleware.Authz(authUC,isView, "users:update:own"))
	api.DELETE("/time-policies/:tpid", m.Schedule.DeleteTimePolicy, middleware.Authz(authUC,isView, "timepolicy:update:org"))

    // SCHEDULE - grafik użytkownika
	schedule := api.Group("/schedule")
	schedule.GET("/:userId/:month", m.Schedule.GetMonthly, middleware.Authz(authUC,isView, "schedule:read:own"))
	//schedule.GET("/monthly/users", m.Schedule.GetMonthlyForAllUsers)
	schedule.GET("/raport/:user_id/:month", m.Schedule.MonthlyReport, middleware.Authz(authUC,isView, "schedule:read:own"))
	schedule.GET("/2/:userId/:month", m.Schedule.GetMonthlyNew, middleware.Authz(authUC,isView, "schedule:read:own"))
	sched := api.Group("/user/:id/schedule")
	sched.GET("/weekly", m.Schedule.GetWeeks, middleware.Authz(authUC,isView, "schedule:read:own"))   // pobierz grafik tygodniowy
	sched.GET("/weekly/:week", m.Schedule.GetWeekly, middleware.Authz(authUC,isView, "schedule:read:own"))   // pobierz grafik tygodniowy
	sched.PUT("/weekly/:week", m.Schedule.UpdateWeekly, middleware.Authz(authUC,isView, "schedule:update:org")) // aktualizuj grafik tygodniowy
	sched.POST("/weekly/:week", m.Schedule.NewWeekly, middleware.Authz(authUC,isView, "schedule:update:org")) // zapisz/aktualizuj grafik tygodniowy
	sched.DELETE("/weekly/:week", m.Schedule.DeleteWeekly, middleware.Authz(authUC,isView, "schedule:update:org")) // zapisz/aktualizuj grafik tygodniowy


	sched.GET("/overrides", m.Schedule.ListOverrides, middleware.Authz(authUC,isView, "schedule:read:own")) // pobierz wyjątki dzienne
	sched.POST("/overrides", m.Schedule.AddOverride, middleware.Authz(authUC,isView, "schedule:update:org"))  // dodaj wyjątek (np. dzień wolny lub inna zmiana)
	schedule.DELETE("/overrides/:oid", m.Schedule.DeleteOverride, middleware.Authz(authUC,isView, "schedule:update:org"))  // dodaj wyjątek (np. dzień wolny lub inna zmiana)


	sick := api.Group("/sickness")

	// uprawnienia sprawdzane router / widoki
	sick.GET("", m.Sickness.List, middleware.Authz(authUC,isView, "sick:read:own"))              // GET /api/sickness?employee_id=&status=
	sick.GET("/:id", m.Sickness.Get, middleware.Authz(authUC,isView, "sick:read:own"))           // GET /api/sickness/:id
	sick.POST("/user/:id", m.Sickness.Create, middleware.Authz(authUC,isView, "sick:update:org"))           // POST /api/sickness
	sick.DELETE("/:id", m.Sickness.Delete, middleware.Authz(authUC,isView, "sick:update:org"))     // DELETE /api/sickness/:id

	// KANBAN - zadania i kalendarz (chronione JWT)
	if m.Kanban != nil {
		api.GET("/tickets", m.Kanban.GetTickets)
		api.POST("/tickets", m.Kanban.CreateTicket)
		api.PUT("/tickets/:id", m.Kanban.UpdateTicket)
		api.DELETE("/tickets/:id", m.Kanban.DeleteTicket)
		api.GET("/calendar", m.Kanban.GetCalendar)
		api.POST("/calendar", m.Kanban.UpdateCalendar)
	}
}

