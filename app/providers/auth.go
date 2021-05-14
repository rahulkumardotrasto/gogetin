package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

//Auth Wrapper for authentication
type Auth struct{}

var (
	authKey = []byte("gogetin-key")
)

//Authenticate ..validates token
func (auth *Auth) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("bearerToken")
		fmt.Println(c.Request.Header.Get("token"))
		authorizationHeader := c.Request.Header.Get("token")
		if authorizationHeader != "" {
			bearerToken := strings.Split(authorizationHeader, " ")

			if len(bearerToken) == 2 {
				token, error := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("An error occured")
					}
					return authKey, nil
				})
				if error != nil {
					c.SecureJSON(http.StatusUnauthorized, gin.H{
						"success": false,
						"message": "Error while authenticating.",
					})
					c.Abort()
				}
				if !token.Valid {
					c.SecureJSON(http.StatusUnauthorized, gin.H{
						"success": false,
						"message": "Unauthorized access or token is expired.",
					})
					c.Abort()
				} else {
					claims, _ := token.Claims.(jwt.MapClaims)
					marchalledClaims, _ := json.Marshal(claims)
					unmarshalledClaims := map[string]string{}
					json.Unmarshal(marchalledClaims, &unmarshalledClaims)
					c.Set("user_id", unmarshalledClaims["jti"])
					c.Next()
				}
			} else {
				c.SecureJSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "User is unauthorized.",
				})
				c.Abort()
			}
		} else {
			c.SecureJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "User is unauthorized.",
			})
			c.Abort()
		}
	}
}
