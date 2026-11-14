package middleware

import (
	//"errors"
	"net/http"
	"strings"
	//"fmt"
	"github.com/labstack/echo/v4"
	echojwt "github.com/labstack/echo-jwt/v4"
	"erpbadger/internal/domain/whitelist"
	"github.com/golang-jwt/jwt/v5"
	"erpbadger/config"
	"erpbadger/internal/usecase/auth"
	"context"
)

type SessionData struct {
    UserID   int
    Username string
}

func Authz(uc auth.Usecase,viewmode bool, required ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

		ctx := c.Request().Context()
		pc, err := uc.CheckPermission(ctx, required...)
		
		if err != nil {
                // Jeśli uprawnienia nie są spełnione:

		// W przypadku interfejsu API zwracamy odpowiedź w postaci JSON
                	if !viewmode {
                    		return c.JSON(http.StatusForbidden, map[string]string{"message": "Brak Uprawnień", "error": "Brak Uprawnień"})
                	}
                // W przypadku interfejsu UI przekierowujemy użytkownika na stronę z informacją o braku uprawnień
		return c.Redirect(http.StatusSeeOther, "/app/unauthorized") 
		}
			c.Set("permctx", pc)
			return next(c)
		}
	}
}



func RequireSession() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {

            tok, ok := c.Get("user").(*jwt.Token)
            if !ok || tok == nil || !tok.Valid {
                return echo.NewHTTPError(http.StatusUnauthorized, "Nieprawidłowy token")
            }

            claims, ok := tok.Claims.(jwt.MapClaims)
            if !ok {
                return echo.NewHTTPError(http.StatusUnauthorized, "Brak sesji")
            }

            uid := int(claims["id"].(float64))
            uname, _ := claims["username"].(string)

            sd := &SessionData{UserID: uid, Username: uname}
            c.Set("session", sd)

            // najważniejsze:
            ctx := context.WithValue(c.Request().Context(), "userID", uid)
            c.SetRequest(c.Request().WithContext(ctx))

            return next(c)
        }
    }
}


func RequireSessionWeb() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {

            tok, ok := c.Get("user").(*jwt.Token)
            if !ok || tok == nil || !tok.Valid {
          		return c.Redirect(http.StatusFound, "/login")     
	    }

            claims, ok := tok.Claims.(jwt.MapClaims)
            if !ok {
                return c.Redirect(http.StatusFound, "/login")
            }

            uid := int(claims["id"].(float64))
            uname, _ := claims["username"].(string)

            sd := &SessionData{UserID: uid, Username: uname}
            c.Set("session", sd)

            // najważniejsze:
            ctx := context.WithValue(c.Request().Context(), "userID", uid)
            c.SetRequest(c.Request().WithContext(ctx))

            return next(c)
        }
    }
}

func RedirectToHTTPS(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // sprawdzenie schematu




        if c.Scheme() == "http" && config.DevMode == false {
            // budujemy URL na https
            target := "https://" + c.Request().Host + c.Request().URL.String()
            return c.Redirect(http.StatusMovedPermanently, target)
        }
        return next(c)
    }
}



func IPFilterMiddleware(repo whitelist.Repository) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            ip := c.RealIP()
   
            if strings.HasPrefix(c.Request().URL.Path, "/indevelopment") {
                return next(c)
            }


  	ok, err := repo.IsAllowed(c.Request().Context(), ip)
            
	    if config.DevMode {
		ok = true
		err = nil
	    }


	if err != nil {
	return c.String(http.StatusInternalServerError, "błąd whitelist")
        }




            if !ok {
               		return c.Redirect(http.StatusMovedPermanently, "/indevelopment")         
       	}
            return next(c)
        }
    }
}


func AuthJWT(secret []byte) echo.MiddlewareFunc {
    return echojwt.WithConfig(echojwt.Config{
        SigningKey:  secret,
        TokenLookup: "header:Authorization:Bearer ,cookie:token",
        ContextKey:  "user",
        NewClaimsFunc: func(c echo.Context) jwt.Claims {
            return jwt.MapClaims{}
        },
        ErrorHandler: func(c echo.Context, err error) error {
            return c.JSON(http.StatusUnauthorized, map[string]string{
                "message": "missing or invalid jwt",
                "error":   "unauthorized",
            })
        },
    })
}

func AuthJWTWeb(secret []byte) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey:  secret,
		TokenLookup: "header:Authorization:Bearer ,cookie:token",
		ContextKey:  "user",         // token będzie dostępny w c.Get("user")
  		NewClaimsFunc: func(c echo.Context) jwt.Claims {
      			return jwt.MapClaims{}
  		},
		ErrorHandler: func(c echo.Context, err error) error {
			// ważne: unikaj pętli przekierowań, jeśli /login też jest w tej grupie
			return c.Redirect(http.StatusFound, "/login")
		},
	})
}




