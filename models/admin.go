package models

import (
	"authserver/database"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// Admin struct
type Admin struct {
	ID                     uint64
	Email                  string
	Password               string
	HashedPassword         string
	Role                   string
	Status                 string
	TokenRow               uint64
	AccessToken            string
	AccessTokenExpiration  int64
	RefreshToken           string `json:"refresh_token"`
	RefreshTokenExpiration int64
	RemoteAddress          string
}

// IsAuthenticated verifies a username and password combination exists in the database
func (admin *Admin) IsAuthenticated() (int, error) {
	query := "SELECT id, email, password, role, status FROM admins WHERE email=?"
	stmt, err := database.Client.Prepare(query)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Internal server error")
	}
	defer stmt.Close()

	// Does admin with given email exist?
	result := stmt.QueryRow(admin.Email)
	if err := result.Scan(&admin.ID, &admin.Email, &admin.HashedPassword, &admin.Role, &admin.Status); err != nil {
		return http.StatusUnauthorized, fmt.Errorf("Please provide valid credentials")
	}

	// Has account been "soft" deleted
	if admin.Status == "deleted" {
		return http.StatusUnauthorized, fmt.Errorf("Please provide valid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.HashedPassword), []byte(admin.Password)); err != nil {
		return http.StatusUnauthorized, fmt.Errorf("Please provide valid credentials")
	}

	// Has account been suspended
	if admin.Status == "suspended" {
		return http.StatusForbidden, fmt.Errorf("Account has been suspended")
	}

	return 0, nil
}

// createExpireTime takes an environment variable containing expiration in minutes and returns token expire time
func createExpireTime(mins string) int64 {
	// use the TOKEN_EXPIRE env variable to calculate the expiration time in Unix from now.
	expire, err := strconv.ParseInt(mins, 10, 64)
	if err != nil {
		expire = 0
	}
	return time.Now().Add(time.Minute * time.Duration(expire)).Unix()
}

// GetAuthToken sets the auth token to be used
func (admin *Admin) GetAuthToken() error {
	expireUnix := createExpireTime(os.Getenv("ADMIN_TOKEN_EXPIRE"))
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = admin.ID
	claims["exp"] = expireUnix
	claims["role"] = admin.Role
	signedToken, err := token.SignedString([]byte(os.Getenv("ADMIN_TOKEN_SECRET")))
	if err != nil {
		return err
	}
	admin.AccessToken = signedToken
	admin.AccessTokenExpiration = expireUnix
	return nil
}

// GetRefreshToken sets the auth token to be used
func (admin *Admin) GetRefreshToken() error {
	expireUnix := createExpireTime(os.Getenv("ADMIN_REFRESH_TOKEN_EXPIRE"))
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	claims := refreshToken.Claims.(jwt.MapClaims)
	claims["sub"] = admin.ID
	claims["exp"] = expireUnix
	signedToken, err := refreshToken.SignedString([]byte(os.Getenv("ADMIN_REFRESH_SECRET")))
	if err != nil {
		return err
	}
	admin.RefreshToken = signedToken
	admin.RefreshTokenExpiration = expireUnix
	return nil
}

// StoreSession adds the token details to the database
func (admin *Admin) StoreSession() (int, error) {
	query := "INSERT INTO admin_sessions(admin_id, access_token, access_token_expiration, refresh_token, refresh_token_expiration, remote_address) VALUES (?, ?, ?, ?, ?, ?);"
	stmt, err := database.Client.Prepare(query)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Internal server error")
	}
	defer stmt.Close()

	_, err = stmt.Exec(admin.ID, admin.AccessToken, admin.AccessTokenExpiration, admin.RefreshToken, admin.RefreshTokenExpiration, admin.RemoteAddress)
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError, fmt.Errorf("Internal server error")
	}
	return 0, nil
}

// Logout logs out an admin by deleting the token and refresh token from the database. If not found in DB do not return error, just log out.
func (admin *Admin) Logout() {
	query := "DELETE FROM admin_sessions WHERE admin_id=? AND access_token=?"
	stmt, err := database.Client.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(admin.ID, admin.AccessToken)
	return
}

// ValidateRefreshToken ...
func (admin *Admin) ValidateRefreshToken() (int, error) {

	_, err := jwt.Parse(admin.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ADMIN_REFRESH_SECRET")), nil
	})
	if err != nil {
		return http.StatusUnauthorized, err
	}

	query := `SELECT
							admins.id,
							admins.status,
							admins.role,
							admin_sessions.revoked,
							admin_sessions.id
						FROM admins, admin_sessions
						WHERE admins.id=admin_sessions.admin_id
							AND admin_sessions.refresh_token=?`
	stmt, err := database.Client.Prepare(query)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Database error")
	}
	defer stmt.Close()

	var status string
	var tokenRevoked bool
	result := stmt.QueryRow(admin.RefreshToken)
	if err := result.Scan(&admin.ID, &status, &admin.Role, &tokenRevoked, &admin.TokenRow); err != nil {
		return http.StatusUnauthorized, fmt.Errorf("Unauthorized")
	}

	if status == "suspended" {
		return http.StatusForbidden, fmt.Errorf("Account has been suspended")
	}

	if status == "deleted" {
		return http.StatusUnauthorized, fmt.Errorf("Unauthorized")
	}

	if tokenRevoked {
		return http.StatusUnauthorized, fmt.Errorf("Unauthorized")
	}

	return 0, nil
}

// UpdateSession updates the access token and refresh token on refreshing
func (admin *Admin) UpdateSession() (int, error) {
	query := "UPDATE admin_sessions SET access_token=?, access_token_expiration=?, refresh_token=?, refresh_token_expiration=?, remote_address=? WHERE id=?"
	stmt, err := database.Client.Prepare(query)
	if err != nil {
		fmt.Println(err)
		return http.StatusInternalServerError, fmt.Errorf("Internal server error")
	}
	defer stmt.Close()

	_, err = stmt.Exec(admin.AccessToken, admin.AccessTokenExpiration, admin.RefreshToken, admin.RefreshTokenExpiration, admin.RemoteAddress, admin.TokenRow)
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError, fmt.Errorf("Internal server error")
	}
	return 0, nil
}
