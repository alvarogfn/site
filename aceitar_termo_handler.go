package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/candidatos-info/site/db"
	"github.com/candidatos-info/site/exception"
	"github.com/candidatos-info/site/token"
	"github.com/labstack/echo"
)

func newAceitarTermoFormHandler(dbClient *db.Client, year int) echo.HandlerFunc {
	return func(c echo.Context) error {
		encodedAccessToken := c.FormValue("token")
		if encodedAccessToken == "" {
			log.Printf("empty token")
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Código de acesso inválido",
				"Success":  false,
			})
		}
		accessTokenBytes, err := base64.StdEncoding.DecodeString(encodedAccessToken)
		if err != nil {
			log.Printf("error decoding access token %s", string(encodedAccessToken))
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
				"Success":  false,
			})
		}
		if !tokenService.IsValid(string(accessTokenBytes)) {
			log.Printf("invalid access token")
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Código de acesso inválido",
				"Success":  false,
			})
		}
		claims, err := token.GetClaims(string(accessTokenBytes))
		if err != nil {
			log.Printf("failed to extract email from token claims, error %v\n", err)
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
				"Success":  false,
			})
		}
		email := claims["email"]
		fmt.Println(email)
		foundCandidate, err := dbClient.GetCandidateByEmail(email, year)
		if err != nil {
			log.Printf("failed find candidate on DB, error %v\n", err)
			switch {
			case err != nil && err.(*exception.Exception).Code == exception.NotFound:
				return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
					"ErrorMsg": fmt.Sprintf("Não encontramos um cadastro de candidatura através do email %s. Por favor verifique se o email está correto.", email),
					"Success":  false,
				})
			case err != nil:
				return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
					"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
					"Success":  false,
				})
			}
		}
		foundCandidate.AcceptedTerms = time.Now()
		if _, err := dbClient.UpdateCandidateProfile(foundCandidate); err != nil {
			log.Printf("failed to update candidate with time that terms were accepted, error %v", err)
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
				"Success":  false,
			})
		}
		return c.Redirect(http.StatusSeeOther, "/atualizar-candidato?token="+string(encodedAccessToken))
	}
}
