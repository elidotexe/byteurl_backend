package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Auth struct {
	Issuer       string
	Audience     string
	Secret       string
	TokenExpiry  time.Duration
	CookieDomain string
	CookiePath   string
	CookieName   string
}

type JWTUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Token string `json:"token"`
}

type Claims struct {
	jwt.RegisteredClaims
}

func (j *Auth) GenerateTokenPair(user *JWTUser) (string, error) {
	// Create a token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = user.Email
	claims["sub"] = fmt.Sprint(user.ID)
	claims["aud"] = j.Audience
	claims["iss"] = j.Issuer
	claims["iat"] = time.Now().UTC().Unix()
	claims["typ"] = "JWT"

	// Set the expiry for JWT
	claims["exp"] = time.Now().UTC().Add(j.TokenExpiry).Unix()

	// Create a signed token
	signedAccessToken, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		return "", err
	}

	// Return TokenPairs
	return signedAccessToken, nil
}

func (j *Auth) GetRefreshCookie(token string) *http.Cookie {
	return &http.Cookie{
		Name:     j.CookieName,
		Path:     j.CookiePath,
		Value:    token,
		Expires:  time.Now().Add(j.TokenExpiry),
		MaxAge:   int(j.TokenExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   j.CookieDomain,
		HttpOnly: true,
		Secure:   true,
	}
}

func (j *Auth) GetTokenFromHeaderAndVerify(w http.ResponseWriter, r *http.Request) (string, *Claims, error) {
	w.Header().Add("Vary", "Authorization")

	// Get the token from the header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", nil, fmt.Errorf("missing Authorization header")
	}

	// Split the token parts
	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 {
		return "", nil, fmt.Errorf("invalid Authorization header format")
	}
	if headerParts[0] != "Bearer" {
		return "", nil, fmt.Errorf("invalid Authorization header format")
	}

	token := headerParts[1]

	claims := &Claims{}

	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(j.Secret), nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "token is expired by") {
			return "", nil, errors.New("expired token")
		}

		return "", nil, errors.New("invalid token")
	}

	if claims.Issuer != j.Issuer {
		return "", nil, fmt.Errorf("invalid issuer")
	}

	return token, claims, nil
}
