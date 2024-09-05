package api

import (
	"crypto/hmac"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"pipe/internal/config"
	"pipe/internal/entity"
	"pipe/pkg/utils"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/labstack/echo/v4"
	"github.com/redis/rueidis"
	"gopkg.in/telebot.v3"
)

func (w *WebApp) index(c echo.Context) error {
	log.Printf("Handling request to index endpoint from URI: %s\n", c.Request().RequestURI)
	return c.JSON(http.StatusOK, map[string]any{
		"status": "ok",
	})
}

func (w *WebApp) getMe(c echo.Context) error {
	log.Printf("Handling request to getMe endpoint from URI: %s\n", c.Request().RequestURI)
	authUser := c.Get("user").(telebot.User)

	u, err := w.App.Account.GetUserByID(authUser.ID)
	if err != nil {
		if err == gocql.ErrNotFound {
			log.Printf("User not found for ID: %d. Creating new user.\n", authUser.ID)
			newUser := entity.User{ID: authUser.ID, PrivateID: utils.GenerateRandomPrivateID(), CreatedAt: time.Now()}
			err = w.App.Account.CreateUser(newUser)
			if err != nil {
				log.Printf("Error creating new user for ID: %d, Error: %v\n", authUser.ID, err)
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"error": "Failed to create user",
				})
			}
			log.Printf("New user created successfully: %+v\n", newUser)
			return c.JSON(http.StatusCreated, newUser)
		}
		log.Printf("Failed to retrieve user for ID: %d, Error: %v\n", authUser.ID, err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to get user",
		})
	}

	log.Printf("User retrieved successfully for ID: %d\n", authUser.ID)
	return c.JSON(http.StatusOK, u)
}

func (w *WebApp) getUser(c echo.Context) error {
	log.Printf("Handling request to getUser endpoint from URI: %s\n", c.Request().RequestURI)
	privateID := c.Param("privateID")

	if privateID == "" {
		log.Println("Private ID is missing in request")
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Private ID can't be empty",
		})
	}

	u, err := w.App.Account.GetUserByPrivateID(privateID)
	if err != nil {
		if err == gocql.ErrNotFound {
			log.Printf("User not found for PrivateID: %s\n", privateID)
			return c.JSON(http.StatusNotFound, map[string]any{
				"error": "User not found",
			})
		}
		log.Printf("Failed to retrieve user for PrivateID: %s, Error: %v\n", privateID, err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to get user",
		})
	}

	log.Printf("User retrieved successfully for PrivateID: %s\n", privateID)
	outUser := entity.User{PrivateID: u.PrivateID, PubKey: u.PubKey}
	return c.JSON(http.StatusOK, outUser)
}

func (w *WebApp) getMessages(c echo.Context) error {
	log.Printf("Handling request to getMessages endpoint from URI: %s\n", c.Request().RequestURI)
	authUser := c.Get("user").(telebot.User)

	messages, err := w.App.Message.GetUserMessages(authUser.ID)
	if err != nil {
		log.Printf("Failed to retrieve messages for UserID: %d, Error: %v\n", authUser.ID, err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to get user messages",
		})
	}

	log.Printf("Messages retrieved successfully for UserID: %d\n", authUser.ID)
	return c.JSON(http.StatusOK, messages)
}

func (w *WebApp) sendMessage(c echo.Context) error {
	log.Printf("Handling request to sendMessage endpoint from URI: %s\n", c.Request().RequestURI)

	var text entity.Text
	err := c.Bind(&text)
	if err != nil {
		log.Println("Failed to bind request body to Text entity")
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Message can't be empty",
		})
	}

	if len(strings.TrimSpace(text.Message)) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Message can't be empty",
		})
	}

	privateID := c.Param("privateID")
	authUser := c.Get("user").(telebot.User)

	if privateID == "" {
		log.Println("Private ID is missing in request")
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Private ID can't be empty",
		})
	}

	u, err := w.App.Account.GetUserByPrivateID(privateID)
	if err != nil {
		if err == gocql.ErrNotFound {
			log.Printf("User not found for PrivateID: %s\n", privateID)
			return c.JSON(http.StatusNotFound, map[string]any{
				"error": "User not found",
			})
		}
		log.Printf("Failed to retrieve user for PrivateID: %s, Error: %v\n", privateID, err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to get user",
		})
	}

	message := entity.Message{ID: gocql.TimeUUID(), FromUser: authUser.ID, ToUser: u.ID, Text: text.Message, Date: time.Now().Unix()}
	if err := w.App.Message.Send(message); err != nil {
		log.Printf("Failed to send message from UserID: %d to UserID: %d, Error: %v\n", authUser.ID, u.ID, err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to send message",
		})
	}

	outMessage := entity.Message{ID: message.ID, Text: message.Text, Date: message.Date}
	messageJSON, err := json.Marshal(outMessage)
	if err != nil {
		log.Printf("Failed to serialize message to JSON, Error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to serialize message",
		})
	}

	log.Printf("Serialized JSON: %#v, Type: %T", string(messageJSON), string(messageJSON))
	if err := w.App.Message.AddToRedis(c.Request().Context(), u.ID, string(messageJSON)); err != nil {
		log.Printf("Failed to add message to Redis, Error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to add message to Redis",
		})
	}

	log.Printf("Message sent successfully from UserID: %d to UserID: %d\n", authUser.ID, u.ID)
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
		log.Printf("Failed to send notification to UserID: %d, Error: %v\n", u.ID, err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "Message sent.",
	})
}

func (w *WebApp) deleteAccount(c echo.Context) error {
	log.Printf("Handling request to deleteAccount endpoint from URI: %s\n", c.Request().RequestURI)
	authUser := c.Get("user").(telebot.User)

	u, err := w.App.Account.GetUserByID(authUser.ID)
	if err != nil {
		if err == gocql.ErrNotFound {
			log.Printf("User not found for ID: %d\n", authUser.ID)
			return c.JSON(http.StatusNotFound, map[string]any{
				"error": "User not found",
			})
		}
		log.Printf("Failed to retrieve user for ID: %d, Error: %v\n", authUser.ID, err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to get user",
		})
	}

	if err := w.App.Account.DeleteUser(u); err != nil {
		log.Printf("Failed to delete user for ID: %d, Error: %v\n", u.ID, err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to delete user",
		})
	}

	log.Printf("User deleted successfully for ID: %d\n", authUser.ID)
	_, err = w.bot.Send(&telebot.Chat{ID: authUser.ID}, "حساب کاربری شما با موفقیت حذف شد. توجه داشته باشید که اگر دوباره وارد مینی اپ شوید حساب کاربری جدیدی برای شما ساخته می شود.")
	if err != nil {
		log.Printf("Failed to send account deletion notification to UserID: %d, Error: %v", authUser.ID, err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "ok",
	})
}

func (w *WebApp) setPubKey(c echo.Context) error {
	log.Printf("Handling request to setPubKey endpoint from URI: %s\n", c.Request().RequestURI)

	var pubkey entity.PubKey
	err := c.Bind(&pubkey)
	if err != nil {
		log.Println("Failed to bind request body to PubKey entity")
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "PubKey can't be empty",
		})
	}

	if len(strings.TrimSpace(pubkey.Value)) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "PubKey can't be empty",
		})
	}

	authUser := c.Get("user").(telebot.User)

	u, err := w.App.Account.GetUserByID(authUser.ID)
	if err != nil {
		if err == gocql.ErrNotFound {
			log.Printf("User not found for ID: %d\n", authUser.ID)
			return c.JSON(http.StatusNotFound, map[string]any{
				"error": "User not found",
			})
		}
		log.Printf("Failed to retrieve user for ID: %d, Error: %v\n", authUser.ID, err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to get user",
		})
	}

	u.PubKey = pubkey.Value

	if err := w.App.Account.SetPubKey(u); err != nil {
		log.Printf("Failed to update PubKey for UserID: %d, Error: %v\n", u.ID, err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to update PubKey",
		})
	}

	log.Printf("PubKey updated successfully for UserID: %d\n", authUser.ID)
	return c.JSON(http.StatusOK, map[string]any{
		"status": "ok",
	})
}

func (w *WebApp) getUpdates(c echo.Context) error {
	log.Printf("Handling request to getUpdates endpoint from URI: %s\n", c.Request().RequestURI)

	timeoutStr := c.QueryParam("timeout")
	if timeoutStr == "" {
		timeoutStr = "0.0"
	}
	timeout, err := strconv.ParseFloat(timeoutStr, 64)
	if err != nil {
		log.Printf("Invalid timeout value: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid timeout value")
	}
	log.Printf("The timeout was set to %f seconds\n", timeout)

	authUser := c.Get("user").(telebot.User)

	messagesJSON, err := w.App.Message.GetRedisMessages(c.Request().Context(), authUser.ID, 0, -1)
	if err != nil {
		log.Printf("Failed to get messages for user: %d, Error: %v\n", authUser.ID, err)
		if errors.Is(err, rueidis.Nil) {
			return c.JSON(http.StatusNoContent, map[string]any{
				"error": "Messages are empty",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to get messages",
		})
	}
	log.Printf("Messages JSON: %#v Items length: %d\n", messagesJSON, len(messagesJSON))

	if len(messagesJSON) < 1 {
		log.Printf("Messages retrieved from redis are empty: %#v for user: %d\n", messagesJSON, authUser.ID)
		return c.JSON(http.StatusNoContent, map[string]any{
			"error": "Messages are empty",
		})
	}

	var messages []entity.Message
	for _, msgJSON := range messagesJSON[1:] {
		var msg entity.Message
		if err := json.Unmarshal([]byte(msgJSON), &msg); err != nil {
			log.Printf("Failed to deserialize message: %s, Error: %v\n", msgJSON, err)
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"error": "Failed to deserialize message",
			})
		}
		messages = append(messages, msg)
	}

	if len(messages) > 0 {
		log.Printf("Messages retrieved from redis: %#v for user: %d\n", messages, authUser.ID)
		return c.JSON(http.StatusOK, messages)
	}

	log.Println("User not have messages in redis. Waiting for new messages..")
	newMessagesJSON, err := w.App.Message.ListenForNewMessage(c.Request().Context(), authUser.ID, timeout)
	if err != nil {
		log.Printf("Failed to get new messages from redis, Error: %v\n", err)
		if errors.Is(err, rueidis.Nil) {
			return c.JSON(http.StatusNoContent, map[string]any{
				"error": "Messages are empty",
			})
		}
		return c.JSON(http.StatusNoContent, map[string]any{
			"error": "timeout",
		})
	}
	log.Printf("New messages JSON: %#v, Items length: %d\n", newMessagesJSON, len(newMessagesJSON))

	var newMessages []entity.Message
	for _, msgJSON := range newMessagesJSON[1:] {
		var msg entity.Message
		if err := json.Unmarshal([]byte(msgJSON), &msg); err != nil {
			log.Printf("Failed to deserialize message: %s, Error: %v\n", msgJSON, err)
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"error": "Failed to deserialize message",
			})
		}
		newMessages = append(newMessages, msg)
	}

	log.Printf("New message retrieved from redis %#v\n", newMessages)
	return c.JSON(http.StatusOK, newMessages)
}

func (w *WebApp) withAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		initData := c.Request().Header.Get("Authorization")

		if initData == "" {
			log.Println("Authorization header is missing")
			return c.JSON(http.StatusUnauthorized, map[string]any{
				"error": "Authorization required",
			})
		}

		authScheme := strings.Split(initData, " ")

		if len(authScheme) != 2 {
			log.Println("Invalid authorization scheme format")
			return c.JSON(http.StatusUnauthorized, map[string]any{
				"error": "Authorization scheme is not valid",
			})
		}

		if authScheme[0] != "tma" {
			log.Println("Invalid authorization scheme")
			return c.JSON(http.StatusUnauthorized, map[string]any{
				"error": "Authorization scheme is not valid",
			})
		}

		isValid, err := w.validateInitData(authScheme[1], config.AppConfig.Token)
		if err != nil {
			log.Printf("Authorization failed with error: %v\n", err)
			return c.JSON(http.StatusUnauthorized, map[string]any{
				"status": isValid,
				"error":  "Authorization failed",
			})
		}

		if !isValid {
			log.Println("Authorization failed due to invalid data")
			return c.JSON(http.StatusUnauthorized, map[string]any{
				"status": isValid,
				"error":  "Authorization failed",
			})
		}

		parsed, err := url.ParseQuery(initData)
		if err != nil {
			log.Println("Failed to parse init data from authorization header")
			return err
		}

		var user telebot.User
		if err := json.Unmarshal([]byte(parsed.Get("user")), &user); err != nil {
			log.Println("Error unmarshalling user data from init data")
			return err
		}

		c.Set("user", user)
		log.Printf("User authenticated successfully: %+v", user)

		return next(c)
	}
}

func (w *WebApp) validateInitData(inputData, botToken string) (bool, error) {
	initData, err := url.ParseQuery(inputData)
	if err != nil {
		log.Printf("Failed to parse web app input data: %v\n", err)
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
		log.Printf("Hash mismatch: %s\n", initData.Get("hash"))
		return false, nil
	}

	log.Println("Init data validated successfully")
	return true, nil
}
