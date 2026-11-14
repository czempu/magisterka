//////////////////////////////////////////////////////////////////////////////////
//
// Autor: Marcel Czempik
//
// Program: Main.go
//
// Inicjacja elementów systemu
//
//
//
///////////////////////////////////////////////////////////////////////////////////
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	n "net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"erpbadger/config"

	//"github.com/robfig/cron/v3"

	//router
	"erpbadger/internal/interface/http/server"

	//infra
	domaini18n "erpbadger/internal/domain/i18n"
	i18ninfra "erpbadger/internal/infra/i18n"
	mailerinfra "erpbadger/internal/infra/mailer"
	"erpbadger/internal/infra/mysql"
	pdfinfra "erpbadger/internal/infra/pdf"
	securityinfra "erpbadger/internal/infra/security"
	i18nhttp "erpbadger/internal/interface/http/i18n"
	"erpbadger/internal/usecase/approval"
	"erpbadger/internal/usecase/auth"
	"erpbadger/internal/usecase/employment"
	"erpbadger/internal/usecase/leaves"
	"erpbadger/internal/usecase/modules"
	"erpbadger/internal/usecase/organization"
	"erpbadger/internal/usecase/permissions"
	"erpbadger/internal/usecase/rcp"
	"erpbadger/internal/usecase/schedule"
	"erpbadger/internal/usecase/sickness"
	"erpbadger/internal/usecase/stats"
	"erpbadger/internal/usecase/users"
	"erpbadger/internal/usecase/workflow"
	kanbanuc "erpbadger/internal/usecase/kanban"

	//handlery
	approvalhttp "erpbadger/internal/interface/http/approval"
	authhttp "erpbadger/internal/interface/http/auth"
	employmenthttp "erpbadger/internal/interface/http/employment"
	leaveshttp "erpbadger/internal/interface/http/leaves"
	moduleshttp "erpbadger/internal/interface/http/modules"
	organizationhttp "erpbadger/internal/interface/http/organization"
	permissionshttp "erpbadger/internal/interface/http/permissions"
	rcphttp "erpbadger/internal/interface/http/rcp"
	schedulehttp "erpbadger/internal/interface/http/schedule"
	sicknesshttp "erpbadger/internal/interface/http/sickness"
	statshttp "erpbadger/internal/interface/http/stats"
	usershttp "erpbadger/internal/interface/http/users"
	viewshttp "erpbadger/internal/interface/http/views"
	workflowhttp "erpbadger/internal/interface/http/workflow"
	kanbanhttp "erpbadger/internal/interface/http/kanban"

	echoSwagger "github.com/swaggo/echo-swagger"
	_ "erpbadger/docs/v1"
)

type TemplateRenderer struct {
	templates  *template.Template
	translator domaini18n.Translator
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if m, ok := data.(echo.Map); ok {
		if _, has := m["Lang"]; !has {
			if lang, ok := c.Get("lang").(string); ok && lang != "" {
				m["Lang"] = lang
			} else {
				m["Lang"] = "pl"
			}
		}
	}
	return t.templates.ExecuteTemplate(w, name, data)
}

// funkcja pomocnicza dla szablonów
func toJson(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func newRenderer(translator domaini18n.Translator) *TemplateRenderer {
	tmpl := template.New("").Funcs(template.FuncMap{
		"toJson": toJson,
		"t": func(key string, lang ...string) string {
			l := "pl"
			if len(lang) > 0 && lang[0] != "" {
				l = lang[0]
			}
			return translator.T(l, key)
		},
	})

	// wczytanie głównych template'ów
	tmpl = template.Must(tmpl.ParseGlob("templates/*.html"))

	// doładowanie partiali
	tmpl = template.Must(tmpl.ParseGlob("templates/partials/*.html"))

	return &TemplateRenderer{
		templates:  tmpl,
		translator: translator,
	}
}

func loadEnvFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			if os.Getenv(k) == "" {
				os.Setenv(k, v)
			}
		}
	}
}

func main() {

	config.DevMode = true
	config.Proxy = false
	
	loadEnvFile(".env")

	resendKey := os.Getenv("RESEND_API_KEY")
	if resendKey == "" {
		resendKey = "dev_placeholder_key"
	}
	appSecret := os.Getenv("APP_SECRET")
	if appSecret == "" {
		appSecret = "dev-secret-key-1234567890"
	}
	
	// Oddzielne dane do baz!
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "CZEMPU:Czempu2025@tcp(127.0.0.1:3306)/RCP?allowNativePasswords=true&parseTime=true"
	}
	kanbanDsn := os.Getenv("KANBAN_DB_DSN") 
	if kanbanDsn == "" {
		kanbanDsn = "CZEMPU:Czempu2025@tcp(127.0.0.1:3306)/kalendarz?allowNativePasswords=true&parseTime=true"
	}

	config.ResendApiKey = resendKey
	config.Secret = appSecret
	
	// 1. POŁĄCZENIE Z GŁÓWNĄ BAZĄ ERPBADGER
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Błąd połączenia z główną bazą: %v", err)
	}
	if pingErr := db.Ping(); pingErr != nil {
		if config.DevMode {
			log.Printf("[DevMode] Uwaga: Brak połączenia z główną bazą MySQL (%v). Uruchamianie w trybie podglądu UI.", pingErr)
		} else {
			log.Fatalf("Błąd PING do głównej DB: %v", pingErr)
		}
	}

	// 2. POŁĄCZENIE Z NIEZALEŻNĄ BAZĄ KANBANA
	kanbanDb, err := sql.Open("mysql", kanbanDsn)
	if err != nil {
		log.Fatalf("Błąd połączenia z bazą Kanbana: %v", err)
	}
	if pingErr := kanbanDb.Ping(); pingErr != nil {
		if config.DevMode {
			log.Printf("[DevMode] Uwaga: Brak połączenia z bazą Kanbana (%v).", pingErr)
		} else {
			log.Fatalf("Błąd PING do bazy Kanbana: %v", pingErr)
		}
	}

	// @title ERPBadger API
	// @version 1.0
	// @description Documentation of ERPBadger API.
	// @contact.email contact@erpbadger.com
	e := echo.New()
	// --- [NOWE] redirector na :80 -> https://www.erpbadger.com ---
	// @securityDefinitions.apikey BearerAuth
	// @in header
	// @name Authorization

	// @host www.localhost:8080
	// @schemes https http

	//cron na ten moment niepotrzebny
	//c := cron.New()
	//c.AddFunc("*/5 * * * *", func() { fmt.Println("Co 5 minut") })
	//c.Start()

	//inicjacja repository
	ipRepo := mysql.NewMySQLWhitelistRepo(db)
	authRepo := mysql.NewAuthRepo(db)
	leavesRepo := mysql.NewLeavesRepo(db)
	modulesRepo := mysql.NewModulesRepo(db)
	permissionsRepo := mysql.NewPermissionsRepo(db)
	rcpRepo := mysql.NewRcpRepo(db)
	usersRepo := mysql.NewUsersRepo(db)
	organizationRepo := mysql.NewOrgRepo(db)
	approvalRepo := mysql.NewApprovalRepo(db)
	workflowRepo := mysql.NewWorkflowRepo(db)
	employmentRepo := mysql.NewEmploymentRepo(db)
	scheduleRepo := mysql.NewScheduleRepo(db)
	statsRepo := mysql.NewStatsRepo(db)
	sicknessRepo := mysql.NewSicknessRepo(db)
	kanbanRepo := mysql.NewKanbanRepo(kanbanDb)

	translator, err := i18ninfra.NewEngine()
	if err != nil {
		log.Fatalf("Błąd inicjalizacji silnika i18n: %v", err)
	}

	hasher := securityinfra.NewBcryptHasher(0)
	pwdGen := securityinfra.NewRandomPasswordGenerator()
	mailerSvc := mailerinfra.NewResendMailer(resendKey)
	pdfGen := pdfinfra.NewMonthlyPDFGenerator("tmp/raports")

	// inicjacja use case'ów 
	authUC := auth.NewUsecase(authRepo, hasher)
	modulesUC := modules.NewUsecase(modulesRepo)
	permissionsUC := permissions.NewUsecase(permissionsRepo)
	rcpUC := rcp.NewUsecase(rcpRepo)
	usersUC := users.NewUsecase(usersRepo, authUC, mailerSvc, hasher, pwdGen)
	organizationUC := organization.NewUsecase(organizationRepo)
	workflowUC := workflow.NewUsecase(workflowRepo)
	approvalUC := approval.NewUsecase(approvalRepo, organizationRepo, workflowUC, permissionsUC)
	leavesUC := leaves.NewUsecase(leavesRepo, approvalUC, authUC)
	employmentUC := employment.NewUsecase(employmentRepo)
	scheduleUC := schedule.NewUsecase(scheduleRepo, pdfGen)
	statsUC := stats.NewUsecase(statsRepo)
	sicknessUC := sickness.NewUsecase(sicknessRepo, approvalUC)
	kanbanUC := kanbanuc.NewUsecase(kanbanRepo)

	jwtSecret := []byte(appSecret)

	// inicjacja Handler'ów
	viewsh := viewshttp.New(authUC, usersUC)
	statsh := statshttp.New(statsUC)
	authh := authhttp.New(authUC, jwtSecret)
	leavesh := leaveshttp.New(leavesUC)
	modulesh := moduleshttp.New(modulesUC)
	permissionsh := permissionshttp.New(permissionsUC)
	rcph := rcphttp.New(rcpUC)
	usersh := usershttp.New(usersUC)
	organizationh := organizationhttp.New(organizationUC)
	approvalh := approvalhttp.New(approvalUC)
	workflowh := workflowhttp.New(workflowUC)
	employmenth := employmenthttp.New(employmentUC)
	scheduleh := schedulehttp.New(scheduleUC, authUC)
	sicknessh := sicknesshttp.New(sicknessUC)
	kanbanh := kanbanhttp.New(kanbanUC)
	i18nh := i18nhttp.New(translator)

	e.Use(i18nhttp.Middleware(translator))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/app/docs/*", echoSwagger.WrapHandler)
	e.GET("/app/docs", func(c echo.Context) error {
		return c.Redirect(n.StatusMovedPermanently, "/app/docs/index.html")
	})

	renderer := newRenderer(translator)
	e.Renderer = renderer

	for _, t := range renderer.templates.Templates() {
		fmt.Println("Załadowany szablon:", t.Name())
	}

	server.RegisterRoutes(e, server.Modules{
		Views:        viewsh,
		Auth:         authh,
		Leaves:       leavesh,
		Modules:      modulesh,
		Permissions:  permissionsh,
		Rcp:          rcph,
		Users:        usersh,
		Organization: organizationh,
		Approval:     approvalh,
		Workflow:     workflowh,
		Employment:   employmenth,
		Schedule:     scheduleh,
		Stats:        statsh,
		Sickness:     sicknessh,
		Kanban:       kanbanh,
		I18n:         i18nh,
	}, ipRepo, authUC, jwtSecret)

	if !config.Proxy {
		if config.DevMode {
			fmt.Println("Devmode")
			log.Println("Serwer HTTP działa na porcie 8080")
			if err := e.Start(":8080"); err != nil {
				log.Fatalf("Błąd uruchamiania serwera HTTP: %v", err)
			}
		} else {
			go n.ListenAndServe(":80", n.RedirectHandler("https://www.erpbadger.com", n.StatusMovedPermanently))

			certFile := "/etc/letsencrypt/live/erpbadger.com/fullchain.pem"
			keyFile := "/etc/letsencrypt/live/erpbadger.com/privkey.pem"

			log.Println("Serwer HTTPS działa na porcie 443")
			if err := e.StartTLS(":443", certFile, keyFile); err != nil {
				log.Fatalf("Błąd uruchamiania serwera HTTPS: %v", err)
			}
		}
	} else {
		log.Println("Serwer HTTP działa na porcie 8080")
		if err := e.Start(":8080"); err != nil {
			log.Fatalf("Błąd uruchamiania serwera HTTP: %v", err)
		}
	}
}
