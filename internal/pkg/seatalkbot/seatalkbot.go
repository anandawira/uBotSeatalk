package seatalkbot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
	"go.uber.org/ratelimit"

	"github.com/anandawira/uBotSeatalk/internal/data/dao"
	"github.com/anandawira/uBotSeatalk/internal/data/model"
	"github.com/anandawira/uBotSeatalk/internal/pkg/log"
)

const (
	host               = "https://openapi.seatalk.io"
	endpointGetSubs    = host + "/messaging/v2/get_bot_subscriber_list"
	endpointSendToSubs = host + "/messaging/v2/single_chat"

	pageSize = 100
)

var (
	appID     = os.Getenv("app_id")
	appSecret = os.Getenv("app_secret")
)

type client struct {
	httpClient     http.Client
	accessToken    string
	rateLimitSubs  ratelimit.Limiter
	ratelimitGroup ratelimit.Limiter
}

func NewClient() dao.ChatBotDAO {
	c := &client{
		httpClient:     http.Client{Timeout: 3 * time.Second},
		accessToken:    "",
		rateLimitSubs:  ratelimit.New(300, ratelimit.Per(1*time.Minute)),
		ratelimitGroup: ratelimit.New(100, ratelimit.Per(1*time.Minute)),
	}

	err := c.updateAccessToken()
	if err != nil {
		panic(err)
	}

	return c
}

func (c *client) GetListSubscriber(ctx context.Context) ([]string, error) {
	var cursor string
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointGetSubs, http.NoBody)
		if err != nil {
			log.ErrorCtx(ctx, "create request error",
				"error", err,
			)
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+c.accessToken)

		q := url.Values{}
		q.Set("page_size", strconv.Itoa(pageSize))
		if cursor != "" {
			q.Set("cursor", cursor)
		}

		req.URL.RawQuery = q.Encode()

		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.ErrorCtx(ctx, "http request error",
				"error", err,
			)
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.ErrorCtx(ctx, "http response code is not 200",
				"code", resp.StatusCode,
			)
			return nil, errors.New("http response code not 200")
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.ErrorCtx(ctx, "read resp body error",
				"error", err,
			)
			return nil, err
		}

		var response model.SeatalkGetBotSubscriberListResponse
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			log.ErrorCtx(ctx, "unmarshal resp body error",
				"error", err,
			)
			return nil, err
		}

		if response.Code != 0 {
			log.ErrorCtx(ctx, "error code not 0",
				"code", response.Code,
			)
			return nil, errors.New("error code not 0")
		}

		return response.Subscribers.EmployeeCodes, nil
	}

}

// TODO: Use generics to simplify http call

func (c *client) SendChatToSubscriber(ctx context.Context, employeeCode string, message json.RawMessage) error {
	reqBody, err := json.Marshal(model.SendToSubsRequestBody{
		EmployeeCode: employeeCode,
		Message:      message,
	})

	if err != nil {
		log.ErrorCtx(ctx, "marshal json error",
			"error", err,
		)
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointSendToSubs, bytes.NewReader(reqBody))
	if err != nil {
		log.ErrorCtx(ctx, "create request error",
			"error", err,
		)
		return err
	}

	c.rateLimitSubs.Take()

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.ErrorCtx(ctx, "http request error",
			"error", err,
		)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.ErrorCtx(ctx, "http response code is not 200",
			"code", resp.StatusCode,
		)
		return errors.New("http response code not 200")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if gjson.Get(string(respBody), "code").Int() != 0 {
		log.ErrorCtx(ctx, "code in response body is not 0",
			"resp_body", respBody,
		)
		return errors.New("code in response body is not 0")
	}

	return nil
}

func (c *client) runAccessTokenScheduler() {
	ticker := time.NewTicker(7000 * time.Second)
	for {
		select {
		case <-ticker.C:
			for i := 0; i < 3; i++ { // can retry max 3 times
				if err := c.updateAccessToken(); err != nil {
					log.Error("refresh access token error", "error", err)
				} else {
					break // success, do not retry
				}
			}
		}
	}
}

func (c *client) updateAccessToken() error {
	reqBody, err := json.Marshal(model.AccessTokenReqBody{
		AppID:     appID,
		AppSecret: appSecret,
	})

	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post("https://openapi.seatalk.io/auth/app_access_token", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("status code not 200")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	accessToken := gjson.Get(string(respBody), "app_access_token")
	if !accessToken.Exists() {
		return fmt.Errorf("access token not exist. resp_body = %s", respBody)
	}

	c.accessToken = accessToken.String()

	return nil
}
