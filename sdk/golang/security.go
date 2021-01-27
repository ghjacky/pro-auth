package authsdk

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	JWT_EXPIRE_SECONDS = 500
)

func GenerateJWTToken(id int64, secret string) string {
	strconv.FormatInt(id, 10)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"time":  fmt.Sprintf("%d", time.Now().Unix()),
		"id":    fmt.Sprintf("%d", id),
		"nonce": fmt.Sprintf("%d", rand.Int31()),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	return tokenString
}

func ParseJWTToken(token, secret string) string {
	ans := ""
	_, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("Unexpected claims type")
		}

		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", claims["id"])
		}

		timeStampString, ok := claims["time"].(string)
		if !ok {
			return nil, fmt.Errorf("Unexpected timeStamp format, it should be string")
		}
		timeStamp, err := strconv.Atoi(timeStampString)
		if err != nil {
			return nil, fmt.Errorf("Unexpected timeStamp length, it should be int64")
		}

		if time.Now().Unix()-int64(timeStamp) > JWT_EXPIRE_SECONDS {
			return nil, fmt.Errorf("Unexpected timeStamp in claims: %d", timeStamp)
		}
		ans = ans + fmt.Sprintf("time: %d\n", timeStamp)

		clientIDString, ok := claims["id"].(string)
		if !ok {
			return nil, fmt.Errorf("Unexpected client id: %v", clientIDString)
		}
		clientID, err := strconv.Atoi(clientIDString)
		if err != nil {
			return nil, fmt.Errorf("Unexpected timeStamp length, it should be int64")
		}

		ans = ans + fmt.Sprintf("id: %d\n", clientID)

		return []byte(secret), nil
	})
	if err != nil {
		return err.Error()
	}

	return ans
}
