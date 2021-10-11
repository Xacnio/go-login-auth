package controllers

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"go-login-auth/database"
	"go-login-auth/utils"
	"os"
	"strconv"
)

func UserInfo(c *fiber.Ctx) error {
	val, _ := c.Locals("User-ID").(uint)
	user, _ := database.GetUserById(val)
	return c.JSON(user)
}

func Login(c *fiber.Ctx) error {
	type LoginInfo struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	data := new(LoginInfo)
	if err := c.BodyParser(&data); err != nil {
		return err
	}

	/* Password Hash (Salt: SC_KEY + email&password custom salt) */
	saltP := fmt.Sprintf("#%s}&{%s#", data.Email, data.Password)
	data.Password = utils.HashPassword(data.Password, saltP)

	id := database.CheckUser(data.Email, data.Password)
	if id == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "user not found",
		})
	}
	user, err := database.GetUserById(id)
	if err != nil {
		return err
	} else {
		ts, err := utils.CreateAccessToken(user.Id)
		if err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
		}
		saveErr := utils.CreateAuthData(user.Id, ts)
		if saveErr != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(saveErr.Error())
		}
		tokens := map[string]string{
			"access_token":  b64.StdEncoding.EncodeToString([]byte(ts.AccessToken)),
			"refresh_token": b64.StdEncoding.EncodeToString([]byte(ts.RefreshToken)),
		}
		return c.Status(fiber.StatusOK).JSON(tokens)
	}
}

func Refresh(c *fiber.Ctx) error {
	// Get a new token with refresh token
	mapToken := map[string]string{}
	if err := c.BodyParser(&mapToken); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
	}
	refreshToken, _ := b64.StdEncoding.DecodeString(mapToken["refresh_token"])
	token, err := jwt.Parse(string(refreshToken), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})
	// Check signature
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON("Refresh token expired, revoked or incorrect")
	}
	// Is token valid?
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(err)
	}
	// Since token is valid, get the uuid
	claims, ok := token.Claims.(jwt.MapClaims) // the token claims should conform to MapClaims
	if ok && token.Valid {
		refreshUuid, ok := claims["refresh_uuid"].(string) // convert the interface to string
		if !ok {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(err)
		}
		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON("Error occurred")
		}
		// Delete the previous refresh token
		deleted, delErr := utils.DeleteAuthData(refreshUuid)
		if delErr != nil || deleted == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON("unauthorized")
		}
		// Create new pairs of refresh and access tokens
		ts, createErr := utils.CreateAccessToken(uint(userId))
		if createErr != nil {
			return c.Status(fiber.StatusForbidden).JSON(createErr.Error())
		}
		// Save the tokens metadata to redis
		saveErr := utils.CreateAuthData(uint(userId), ts)
		if saveErr != nil {
			return c.Status(fiber.StatusForbidden).JSON(saveErr.Error())
		}
		tokens := map[string]string{
			"access_token":  b64.StdEncoding.EncodeToString([]byte(ts.AccessToken)),
			"refresh_token": b64.StdEncoding.EncodeToString([]byte(ts.RefreshToken)),
		}
		return c.Status(fiber.StatusCreated).JSON(tokens)
	} else {
		return c.Status(fiber.StatusUnauthorized).JSON("refresh token expired or revoked")
	}
}

func Logout(c *fiber.Ctx) error {
	metadata, err := utils.ExtractTokenMetadata(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": 0, "message": "unauthorized"})
	}
	delErr := utils.DeleteTokens(metadata)
	if delErr != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": 0, "message": delErr.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": 1, "message": "Successfully logged out"})
}
