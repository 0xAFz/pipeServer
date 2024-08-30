package api

import (
	"crypto/hmac"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"pipe/internal/config"
	"pipe/internal/entity"
	"pipe/pkg/utils"
	"sort"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/labstack/echo/v4"
	"gopkg.in/telebot.v3"
)

func (w *WebApp) index(c echo.Context) error {
	log.Println("request from: ", c.Request().RequestURI)
	return c.JSON(http.StatusOK, map[string]any{
		"status": "ok",
	})
}

func (w *WebApp) getMe(c echo.Context) error {
	log.Println("request from: ", c.Request().RequestURI)
	authUser := c.Get("user").(telebot.User)

	u, err := w.App.Account.GetUserByID(authUser.ID)
	if err != nil {
		if err == gocql.ErrNotFound {
			newUser := entity.User{ID: authUser.ID, PrivateID: utils.GenerateRandomPrivateID(), CreatedAt: time.Now()}
			err = w.App.Account.CreateUser(newUser)
			if err != nil {
				log.Println("error when creating new user: ", err)
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"error": "Failed to create user",
				})
			}
			log.Println("new user created: ", newUser)
			return c.JSON(http.StatusCreated, newUser)
		}

		log.Println("failed to get user: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to get user",
		})
	}

	return c.JSON(http.StatusOK, u)
}

func (w *WebApp) getUser(c echo.Context) error {
	log.Println("request from: ", c.Request().RequestURI)
	privateID := c.Param("private_id")

	if privateID == "" {
		log.Println("nil private id")
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "private id can't be empty",
		})
	}

	u, err := w.App.Account.GetUserByPrivateID(privateID)
	if err != nil {
		if err == gocql.ErrNotFound {
			log.Println("user not found: ", privateID)
			return c.JSON(http.StatusNotFound, map[string]any{
				"error": "user not found",
			})
		}

		log.Println("failed to get user: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to get user",
		})
	}

	outUser := entity.User{PrivateID: u.PrivateID, PubKey: u.PubKey}

	return c.JSON(http.StatusOK, outUser)
}

func (w *WebApp) getMessages(c echo.Context) error {
	log.Println("request from: ", c.Request().RequestURI)
	authUser := c.Get("user").(telebot.User)

	messages, err := w.App.Message.GetUserMessages(authUser.ID)
	if err != nil {
		log.Println("failed to get user messages: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to get user messages",
		})
	}

	return c.JSON(http.StatusOK, messages)
}

func (w *WebApp) sendMessage(c echo.Context) error {
	log.Println("request from: ", c.Request().RequestURI)

	var text entity.Text
	err := c.Bind(&text)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "message can't be empty",
		})
	}

	privateID := c.Param("private_id")
	authUser := c.Get("user").(telebot.User)

	if privateID == "" {
		log.Println("nil private id")
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "private id can't be empty",
		})
	}

	u, err := w.App.Account.GetUserByPrivateID(privateID)
	if err != nil {
		if err == gocql.ErrNotFound {
			log.Println("user not found: ", privateID)
			return c.JSON(http.StatusNotFound, map[string]any{
				"error": "user not found",
			})
		}

		log.Println("failed to get user: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to get user",
		})
	}

	message := entity.Message{ID: gocql.TimeUUID(), FromUser: authUser.ID, ToUser: u.ID, Text: text.Message, Date: time.Now().Unix()}
	err = w.App.Message.Send(message)
	if err != nil {
		log.Println("failed to send message: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to send message",
		})
	}

	_, err = w.bot.Send(&telebot.Chat{ID: u.ID}, "یه پیام جدید داری.", &telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{
			{
				{
					Text:   "Open",
					WebApp: &telebot.WebApp{URL: config.AppConfig.ClientURL},
				},
			},
		},
	})

	if err != nil {
		log.Println("can't send notif to user: ", err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "message sent.",
	})
}

func (w *WebApp) deleteAccount(c echo.Context) error {
	log.Println("request from: ", c.Request().RequestURI)
	authUser := c.Get("user").(telebot.User)

	u, err := w.App.Account.GetUserByID(authUser.ID)
	if err != nil {
		if err == gocql.ErrNotFound {
			log.Println("user not found: ", authUser.ID)
			return c.JSON(http.StatusNotFound, map[string]any{
				"error": "user not found",
			})
		}

		log.Println("failed to get user: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to get user",
		})
	}

	if err := w.App.Account.DeleteUser(u); err != nil {
		log.Println("failed to dalete user: ", err)

		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to delete user",
		})
	}

	_, err = w.bot.Send(&telebot.Chat{ID: authUser.ID}, "حساب کاربری شما با موفقیت حذف شد. توجه داشته باشید که اگر دوباره وارد مینی اپ شوید حساب کاربری جدیدی برای شما ساخته می شود.")
	if err != nil {
		log.Println("can't send notif to user: ", err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "ok",
	})
}

func (w *WebApp) setPubKey(c echo.Context) error {
	log.Println("request from: ", c.Request().RequestURI)

	var pubkey entity.PubKey
	err := c.Bind(&pubkey)
	if err != nil {
		log.Println("set pubkey bad request")
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "pubkey can't be empty",
		})
	}

	authUser := c.Get("user").(telebot.User)

	u, err := w.App.Account.GetUserByID(authUser.ID)
	if err != nil {
		if err == gocql.ErrNotFound {
			log.Println("user not found: ", authUser.ID)
			return c.JSON(http.StatusNotFound, map[string]any{
				"error": "user not found",
			})
		}

		log.Println("failed to get user: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to get user",
		})
	}

	u.PubKey = pubkey.Value

	if err := w.App.Account.SetPubKey(u); err != nil {
		log.Println("failed to update user pubkey: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to update user pubkey",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "ok",
	})
}

func (w *WebApp) withAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		initData := c.Request().Header.Get("Authorization")

		if initData == "" {
			log.Println("auth header missed")
			return c.JSON(http.StatusUnauthorized, map[string]any{
				"error": "authorization required",
			})
		}

		authScheme := strings.Split(initData, " ")

		authParamsLen := len(authScheme)
		if authParamsLen < 2 || authParamsLen > 2 {
			log.Println("auth scheme is not valid")
			return c.JSON(http.StatusUnauthorized, map[string]any{
				"error": "authorization scheme is not valid",
			})
		}

		if strings.Compare(authScheme[0], "tma") != 0 {
			log.Println("auth scheme is not valid")
			return c.JSON(http.StatusUnauthorized, map[string]any{
				"error": "authorization scheme is not valid",
			})
		}

		isValid, err := w.validateInitData(authScheme[1], config.AppConfig.Token)
		if err != nil {
			log.Println("auth failed with error: ", err)
			return c.JSON(http.StatusUnauthorized, map[string]any{
				"status": isValid,
				"error":  "authorization failed",
			})
		}

		if !isValid {
			log.Println("auth failed")
			return c.JSON(http.StatusUnauthorized, map[string]any{
				"status": isValid,
				"error":  "authorization failed",
			})
		}

		parsed, _ := url.ParseQuery(initData)
		var user telebot.User
		if err := json.Unmarshal([]byte(parsed.Get("user")), &user); err != nil {
			log.Println("error when unmarshaling init data user")
			return err
		}

		c.Set("user", user)
		log.Println(user)

		return next(c)
	}
}

func (w *WebApp) validateInitData(inputData, botToken string) (bool, error) {
	initData, err := url.ParseQuery(inputData)
	if err != nil {
		log.Println("couldn't parse web app input data")
		return false, err
	}

	dataCheckString := make([]string, 0, len(initData))
	for k, v := range initData {
		if k == "hash" {
			continue
		}
		if len(v) > 0 {
			dataCheckString = append(dataCheckString, fmt.Sprintf("%s=%s", k, v[0]))
		}
	}

	sort.Strings(dataCheckString)

	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(botToken))

	hHash := hmac.New(sha256.New, secret.Sum(nil))
	hHash.Write([]byte(strings.Join(dataCheckString, "\n")))

	hash := hex.EncodeToString(hHash.Sum(nil))

	if initData.Get("hash") != hash {
		return false, nil
	}

	return true, nil
}
