
package usershttp

import (
    "net/http"
	"strconv"	
    "fmt"
	"erpbadger/config"
    "github.com/labstack/echo/v4"
    uc "erpbadger/internal/usecase/users"
	do "erpbadger/internal/domain/users"
)






func getUser(c echo.Context) int {
    idParam := c.Param("id")
    id, err := strconv.Atoi(idParam)
    if err != nil {
    }
    return id 

}

var p = fmt.Println


var jwtSecret = []byte(config.Secret)

type Handlers struct {
    svc uc.Usecase
}
func New(svc uc.Usecase) *Handlers { return &Handlers{svc: svc} }



type AddUserNewData struct{        // moduł użytkownicy
    Username string `json:"username"`
    Email string `json:"email"`
}

type TargetUserID struct {       // moduł użytkownicy
    UserID int `json:"userid"`
}

type UserDetails struct {
	ID int	`json:"id"`
	Username string `json:"username"`
	Email string `json:"email"`
	IsEmployee int `json:"isemployee"`
	CountryCode string `json:"countrycode"` 
	NominalDayMinutes int `json:"NominalDayMinutes"`
	Timezone string `json:"timezone"`
}


// UserItem reprezentuje pojedynczego użytkownika w odpowiedzi /api/users.
type UserItem struct {
	ID       int    `json:"id" example:"1"`
	Username string `json:"username" example:"czempu"`
	Name     string `json:"Name" example:"Marcel"`
	Surname  string `json:"Surname" example:"Czempik"`
}

// UsersResponse to odpowiedź zawierająca listę użytkowników.
type UsersResponse struct {
	Data []UserItem `json:"data"`
}

// ErrorResponse opisuje standardowy błąd API.
type ErrorResponse struct {
	Error string `json:"error" example:"invalid credentials"`
}
// Users godoc
// @Summary      Users
// @Description  Zwraca listę użytkowników względem poziomu uprawnień.
// @Tags         Users
// @Produce      json
// @Success      200  {object}  UsersResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /api/users [get]
func (h *Handlers) Users(c echo.Context) error {


    mods, err := h.svc.GetUsersInfo(c.Request().Context())
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": mods})
}


//users
func (h *Handlers) AddUser(c echo.Context) error {

    var req AddUserNewData
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "błędne dane"})
    }

    _, err := h.svc.AddUser(c.Request().Context(), req.Username, req.Email)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
  
    return c.JSON(http.StatusOK, echo.Map{"message": "Dodano użytkownika"})
}


func (h *Handlers) PasswordChange(c echo.Context) error {

    var req do.UserPassChange
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "błędne dane"})
    }
	
	if len(req.NewPassword) > 32 {
		 return c.JSON(http.StatusBadRequest, echo.Map{"error": "hasło jest za długie (maksymalnie 32 znaki)"})
	}
	if len(req.NewPassword) < 4 {
		 return c.JSON(http.StatusBadRequest, echo.Map{"error": "hasło jest za krótkie (minimalnie 5 znaków)"})
	}



	idstring := c.Param("id")
	userid, erratoi := strconv.Atoi(idstring)
	if erratoi != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "nieprawidłowy id"})
	}




    err := h.svc.PasswordChange(c.Request().Context(), userid, req.OldPassword, req.NewPassword)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
  
    return c.JSON(http.StatusOK, echo.Map{"message": "hasło zaktualizowane"})
}


func (h *Handlers) PasswordReset(c echo.Context) error {

	idstring := c.Param("id")
	userid, erratoi := strconv.Atoi(idstring)
	if erratoi != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "nieprawidłowy id"})
	}

    result, err := h.svc.ResetPassword(c.Request().Context(), userid)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
  
    return c.JSON(http.StatusOK, echo.Map{"message": result})
}




//users
func (h *Handlers) DeleteUser(c echo.Context) error {

    var req TargetUserID
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "błędne dane"})
    }

    //userToken := c.Get("user").(*jwt.Token)
    //claims := userToken.Claims.(jwt.MapClaims)
    //userID := int(claims["id"].(float64))

    info, err := h.svc.DeleteUser(c.Request().Context(), req.UserID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": info})
}


func (h *Handlers) GetUserPII(c echo.Context) error {
	
    data, err := h.svc.GetUserPII(c.Request().Context(), getUser(c))
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": data})

}

//users
func (h *Handlers) UserDetails(c echo.Context) error {

    data, err := h.svc.UserDetails(c.Request().Context(), getUser(c))
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": data})
}


//users
func (h *Handlers) UserDetailsEdit(c echo.Context) error {
 var req do.UserDetails
 if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "błędne dane"})
    }

	if len(req.Username) > 50 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Przekroczono dopuszczalną długość nazwy użytkownika (50)"})
	}
	if len(req.Email) > 100 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Przekroczono dopuszczalną długość maila (100)"})
	}
	if len(req.Name) > 100 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Przekroczono dopuszczalną długość imienia (100)"})
	}
	if len(req.Surname) > 100 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Przekroczono dopuszczalną długość nazwiska (100)", "message": "Przekroczono dopuszczalną długość nazwiska (100)"})
	}




    username, err := h.svc.UserDetailsEdit(c.Request().Context(), getUser(c), req)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": username})
}


















