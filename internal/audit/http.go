package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// HTTPObserver доставляет события аудита HTTP POST-запросами в формате JSON.
type HTTPObserver struct {
	url    string
	client *http.Client
}

// NewHTTPObserver создаёт приёмник для полного HTTP или HTTPS URL.
// Таймаут запроса составляет две секунды; перенаправления не выполняются.
func NewHTTPObserver(address string) (*HTTPObserver, error) {
	parsed, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("parse audit URL: %w", err)
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return nil, fmt.Errorf("audit URL must be an absolute HTTP or HTTPS URL")
	}

	return &HTTPObserver{
		url: address,
		client: &http.Client{
			Timeout: 2 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

// Notify отправляет событие и проверяет успешный статус 2xx.
// Ошибки HTTP-запроса, чтения ответа и статусы вне 2xx возвращаются вызывающему коду.
// Контекст позволяет отменить запрос или ограничить его длительность.
func (o *HTTPObserver) Notify(ctx context.Context, event Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode audit event: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create audit request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("send audit event: %w", err)
	}
	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return fmt.Errorf("read audit response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("audit receiver returned status %d", resp.StatusCode)
	}
	return nil
}
