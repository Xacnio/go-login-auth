package utils

import (
	b64 "encoding/base64"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/twinj/uuid"
	"go-login-auth/database"
	"os"
	"strconv"
	"strings"
	"time"
)

type AccessDetails struct {
	AccessUuid string
	UserId     uint
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

func CreateAccessToken(userid uint) (*TokenDetails, error) {
	/* Token Details */
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 15).Unix()
	td.AccessUuid = uuid.NewV4().String()
	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = td.AccessUuid + "++" + strconv.FormatUint(uint64(userid), 10)

	/* Create Access Token from Data with JWT */
	var err error
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = userid
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS384, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	/* Create Refresh Token from Data with JWT */
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userid
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS384, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		return nil, err
	}
	return td, nil
}

func CreateAuthData(userid uint, td *TokenDetails) error {
	// Store the token data in Redis
	at := time.Unix(td.AtExpires, 0)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	errAccess := database.RedisClient.Set(td.AccessUuid, strconv.FormatUint(uint64(userid), 10), at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}
	errRefresh := database.RedisClient.Set(td.RefreshUuid, strconv.FormatUint(uint64(userid), 10), rt.Sub(now)).Err()
	if errRefresh != nil {
		return errRefresh
	}
	return nil
}

func ExtractAccessToken(c *fiber.Ctx) string {
	// Fetch Authorization Data from HTTP header & decode with base64 & "Bearer <token>"
	bearToken := c.Get(fiber.HeaderAuthorization)
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		sDec, _ := b64.StdEncoding.DecodeString(strArr[1])
		return string(sDec)
	}
	return ""
}

func VerifyAccessToken(r *fiber.Ctx) (*jwt.Token, error) {
	// Try to verify signature of the token
	tokenString := ExtractAccessToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func ExtractTokenMetadata(c *fiber.Ctx) (*AccessDetails, error) {
	// Decode the access token & extract user info with decoded access token
	token, err := VerifyAccessToken(c)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}
		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			return nil, err
		}
		return &AccessDetails{
			AccessUuid: accessUuid,
			UserId:     uint(userId),
		}, nil
	}
	return nil, err
}

func FetchAuthData(authD *AccessDetails) (uint, error) {
	// Verify access token from Redis & Check the User ID in access token equals User ID in redis data
	userid, err := database.RedisClient.Get(authD.AccessUuid).Result()
	if err != nil {
		return 0, errors.New("unauthorized")
	}
	userID, _ := strconv.ParseUint(userid, 10, 64)
	if authD.UserId != uint(userID) {
		return 0, errors.New("unauthorized")
	}
	return uint(userID), nil
}

func DeleteAuthData(givenUuid string) (int64, error) {
	// Delete access token data from Redis
	deleted, err := database.RedisClient.Del(givenUuid).Result()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

func DeleteTokens(authD *AccessDetails) error {
	// Delete Access/Refresh Tokens in Redis (Logout)
	refreshUuid := fmt.Sprintf("%s++%d", authD.AccessUuid, authD.UserId)
	deletedAt, err := database.RedisClient.Del(authD.AccessUuid).Result()
	if err != nil {
		return err
	}
	deletedRt, err := database.RedisClient.Del(refreshUuid).Result()
	if err != nil {
		return err
	}
	if deletedAt != 1 || deletedRt != 1 {
		return errors.New("something went wrong")
	}
	return nil
}
