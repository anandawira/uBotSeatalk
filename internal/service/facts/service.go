package facts

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
	"golang.org/x/sync/errgroup"

	"github.com/anandawira/uBotSeatalk/internal/data/dao"
	"github.com/anandawira/uBotSeatalk/internal/data/model"
	"github.com/anandawira/uBotSeatalk/internal/pkg/log"
)

const host = "https://uselessfacts.jsph.pl"

type Service interface {
	SendTodayFact(ctx context.Context) error
	SendRandomFactToUser(ctx context.Context, employeeCode string) error
}

type service struct {
	httpClient http.Client
	botClient  dao.ChatBotDAO
}

func NewService(botClient dao.ChatBotDAO) Service {
	return &service{
		httpClient: http.Client{
			Transport:     nil,
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       3 * time.Second,
		},
		botClient: botClient,
	}
}

func (s *service) SendTodayFact(ctx context.Context) error {
	factText, err := s.getFact(ctx, "/api/v2/facts/today")
	if err != nil {
		return err
	}

	message, err := model.NewTextMessage(factText)
	if err != nil {
		return err
	}

	employeeCodes, err := s.botClient.GetListSubscriber(ctx)
	if err != nil {
		return err
	}

	group := errgroup.Group{}
	for _, employeeCode := range employeeCodes {
		group.Go(func() error {
			return s.botClient.SendChatToSubscriber(ctx, employeeCode, message)
		})
	}

	return group.Wait()
}

func (s *service) SendRandomFactToUser(ctx context.Context, employeeCode string) error {
	fact, err := s.getFact(ctx, "/api/v2/facts/random")
	if err != nil {
		return err
	}

	message, err := model.NewTextMessage(fact)
	if err != nil {
		log.ErrorCtx(ctx, "create message error",
			"error", err,
			"fact", fact,
		)
		return err
	}

	return s.botClient.SendChatToSubscriber(ctx, employeeCode, message)
}

func (s *service) getFact(ctx context.Context, path string) (string, error) {
	resp, err := s.httpClient.Get(host + path)
	if err != nil {
		log.ErrorCtx(ctx, "http call error",
			"error", err,
		)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.ErrorCtx(ctx, "http status not 200", "status_code", resp.StatusCode)
		return "", errors.New("status code not 200")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.ErrorCtx(ctx, "reading response body error",
			"error", err,
		)
		return "", err
	}

	text := gjson.Get(string(respBody), "text")
	if !text.Exists() {
		log.ErrorCtx(ctx, "text not found inside response body",
			"resp_body", respBody,
		)
		return "", errors.New("text not found")
	}
	return gjson.Get(string(respBody), "text").String(), nil
}
