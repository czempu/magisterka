package authhttp

import (
    "net/http"
    "time"
	"strconv"	
    "fmt"
    "github.com/golang-jwt/jwt/v5"
    "github.com/labstack/echo/v4"
    uc "erpbadger/internal/usecase/auth"

)






func getUser(c echo.Context) int {
    idParam := c.Param("id")
    id, err := strconv.Atoi(idParam)
    if err != nil {
    }
    return id 

}


// Helper do kasowania ciastek
func clearCookie(c echo.Context, name string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true, 
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
}

func (h *Handlers) Logout(c echo.Context) error {
	clearCookie(c, "token")
	clearCookie(c, "logged")
	// Po wylogowaniu leć na /login
	return c.Redirect(http.StatusSeeOther, "/")
}









var p = fmt.Println

type Handlers struct {
    svc uc.Usecase
    jwtSecret []byte
}
func New(svc uc.Usecase, jwtSecret []byte) *Handlers { return &Handlers{svc: svc, jwtSecret: jwtSecret} }


type LoginRequest struct {    // logowanie 
    User string `json:"user"`
    Pass string `json:"pass"`
}



// LoginResponse opisuje poprawną odpowiedź z tokenem JWT.
type LoginResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn   int64  `json:"expires_in"   example:"604800"`
	TokenType   string `json:"token_type"   example:"Bearer"`
}

// ErrorResponse opisuje standardowy błąd API.
type ErrorResponse struct {
	Error string `json:"error" example:"invalid credentials"`
}


// < login dto >


// HANDLER (Echo) + adnotacje Swaggera

// Login/token godoc
// @Summary      Login/token (issue JWT)
// @Description  Zwraca token dostępu po poprawnej autoryzacji.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      LoginRequest  true  "Login payload"
// @Success      200      {object}  LoginResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /auth/login/token [post]
func (h *Handlers) LoginToken(c echo.Context) error {  // auth
    var req LoginRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "błędne dane"})
    }

    userID, err := h.svc.CheckUserPassword(c.Request().Context(), req.User, req.Pass)
    if err != nil {
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Nieprawidłowy login lub hasło!"})
    }

    // wygeneruj JWT
    exp := time.Now().Add(7 * 24 * time.Hour)
    claims := jwt.MapClaims{
        "username": req.User,
        "id":       userID,
        "exp":      exp.Unix(),
    }
    tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := tok.SignedString(h.jwtSecret)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": "błąd generowania tokena"})
    }

    // ZWRÓĆ token w odpowiedzi (dla klientów używających Bearer)
    return c.JSON(http.StatusOK, echo.Map{
        "access_token": tokenString,
        "token_type":   "Bearer",
        "expires_in":   int((7 * 24 * time.Hour).Seconds()),
    })



}











// Login/cookie godoc
// @Summary      Login/cookie (issue JWT)
// @Description  Zwraca ciasteczko po poprawnej autoryzacji.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      LoginRequest  true  "Login payload"
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /auth/login/cookie [post]
func (h *Handlers) LoginCookie(c echo.Context) error {  // auth
    var req LoginRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "błędne dane"})
    }

    userID, err := h.svc.CheckUserPassword(c.Request().Context(), req.User, req.Pass)
    if err != nil {
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Nieprawidłowy login lub hasło!"})
    }

    // wygeneruj JWT
    exp := time.Now().Add(7 * 24 * time.Hour)
    claims := jwt.MapClaims{
        "username": req.User,
        "id":       userID,
        "exp":      exp.Unix(),
    }
    tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := tok.SignedString(h.jwtSecret)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": "błąd generowania tokena"})
    }


    // (opcjonalnie) ustaw HttpOnly cookie 
    c.SetCookie(&http.Cookie{
        Name:     "token",
        Value:    tokenString,
        Path:     "/",
        HttpOnly: true,
        Secure:   true,                 
	SameSite: http.SameSiteLaxMode, 
	MaxAge:   int((7 * 24 * time.Hour).Seconds()),
    })
//dla strony głównej - ciasteczko sprawdza czy zalogowany jeśli tak to pokazuje dashboard
   c.SetCookie(&http.Cookie{
    Name:     "logged",
    Value:    "1",
    Path:     "/",
    HttpOnly: false, 
    Secure:   true,  
    SameSite: http.SameSiteLaxMode,
    MaxAge:   int((7 * 24 * time.Hour).Seconds()),
})





	return c.JSON(http.StatusOK, echo.Map{
        "access_token": "token",
        "token_type":   "cookie",
        "expires_in":   int((7 * 24 * time.Hour).Seconds()),
    })

}







































































