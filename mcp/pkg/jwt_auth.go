package pkg

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/dolthub/dolt/go/libraries/utils/jwtauth"
)

// withBearerAuth enforces Authorization: Bearer <token>
func withBearerAuth(logger *zap.Logger, next http.Handler, jwkClaimsMap map[string]string, jwksUrl string) (http.Handler, error) {
	pr, err := getJWTProvider(jwkClaimsMap, jwksUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to get JWT provider: %w", err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// auth will be "" if the "Authorization" header is not set
		auth := r.Header.Get("Authorization")

		var token string
		if strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
			token = strings.TrimSpace(token)
		} else {
			vals := r.URL.Query()

			for key, arr := range vals {
				if key == "jwt" && len(arr) == 1 {
					token = strings.TrimSpace(arr[0])
					break
				}
			}

			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}

		// validate token
		valid, _, err := validateJWT(logger, pr, token, time.Now())
		if err != nil || !valid {
			logger.Info("unable to authorize jwt", zap.Bool("valid", valid), zap.Error(err))
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		}

		next.ServeHTTP(w, r)
	}), nil
}

func validateJWT(logger *zap.Logger, pr jwtauth.JWTProvider, token string, reqTime time.Time) (bool, *jwtauth.Claims, error) {
	vd, err := jwtauth.NewJWTValidator(pr)
	if err != nil {
		logger.Info("unable to create JWTValidator", zap.Error(err))
		return false, nil, fmt.Errorf("unable to get JWT validator: %w", err)
	}

	privClaims, err := vd.ValidateJWT(token, reqTime)
	if err != nil {
		logger.Info("unable to validate jwt", zap.Error(err))
		return false, nil, fmt.Errorf("unable to validate JWT token: %w", err)
	}

	logger.Info("MCP Auth with JWT", zap.String("id", privClaims.ID), zap.String("iss", privClaims.Issuer), zap.String("sub", privClaims.Subject), zap.String("on_behalf_of", privClaims.OnBehalfOf))

	return true, privClaims, nil
}

func getJWTProvider(expectedClaimsMap map[string]string, url string) (jwtauth.JWTProvider, error) {
	pr := jwtauth.JWTProvider{URL: url}
	for name, claim := range expectedClaimsMap {
		switch name {
		case "iss":
			pr.Issuer = claim
		case "aud":
			pr.Audience = claim
		case "sub":
			pr.Subject = claim
		default:
			return pr, errors.New("ValidateJWT: Unexpected expected claim found in user identity")
		}
	}
	return pr, nil
}
