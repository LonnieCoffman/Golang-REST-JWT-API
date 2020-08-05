package auth

import (
	"authserver/database"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type Auth struct {
	ID                uint64
	AccessTokenString string
	AccessToken       *jwt.Token
}

var (
	rolelvl = map[string]int{
		"superadmin": 3,
		"admin":      2,
		"client":     1,
	}
	authDetails = Auth{}
)

func SuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if status, err := validateAdminRole(c, "superadmin"); err != nil {
			c.JSON(status, gin.H{"success": false, "message": err.Error()})
			c.Abort()
		}
		c.Next()
	}
}

func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if status, err := validateAdminRole(c, "admin"); err != nil {
			c.JSON(status, gin.H{"success": false, "message": err.Error()})
			c.Abort()
		}
		c.Next()
	}
}

func validateAdminRole(c *gin.Context, role string) (int, error) {

	if err := VerifyToken(c.Request); err != nil {
		return http.StatusUnauthorized, err
	}

	if err := getAdminID(); err != nil {
		return http.StatusUnauthorized, err
	}

	query := `SELECT
							admins.role,
							admins.status,
							admin_sessions.revoked
						FROM admins, admin_sessions
						WHERE admins.id=?
							AND admins.id=admin_sessions.admin_id
							AND admin_sessions.access_token=?`
	stmt, err := database.Client.Prepare(query)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Database error")
	}
	defer stmt.Close()

	var dbRole, status string
	var tokenRevoked bool
	result := stmt.QueryRow(authDetails.ID, authDetails.AccessTokenString)
	if err := result.Scan(&dbRole, &status, &tokenRevoked); err != nil {
		return http.StatusUnauthorized, fmt.Errorf("Unauthorized")
	}

	if rolelvl[dbRole] < rolelvl[role] {
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

	c.Set("AdminID", authDetails.ID)
	c.Set("AdminToken", authDetails.AccessTokenString)

	return 0, nil
}

func getAdminID() error {
	claims, ok := authDetails.AccessToken.Claims.(jwt.MapClaims)
	if ok && authDetails.AccessToken.Valid {
		adminID, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			return err
		}
		authDetails.ID = adminID
	}
	return nil
}

// verifyToken validates the token signing method and secret
func VerifyToken(r *http.Request) error {
	extractToken(r)
	token, err := jwt.Parse(authDetails.AccessTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ADMIN_TOKEN_SECRET")), nil
	})
	if err != nil {
		return err
	}
	authDetails.AccessToken = token
	return nil
}

// extractToken parses the header and extracts the token
func extractToken(r *http.Request) {
	auth := r.Header.Get("Authorization")
	strArr := strings.Split(auth, "Bearer ")
	if len(strArr) == 2 {
		authDetails.AccessTokenString = strArr[1]
	}
}
