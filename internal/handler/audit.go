package handler

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/Vitaly898/metricsCollector/internal/audit"
)

// AuditPublisher публикует события успешно обработанных запросов записи метрик.
// Обработчики допускают nil вместо издателя, чтобы отключить аудит.
type AuditPublisher interface {
	// Publish доставляет событие приёмникам аудита и возвращает ошибки доставки.
	Publish(context.Context, audit.Event) error
}

func publishAudit(publisher AuditPublisher, r *http.Request, metrics []string) {
	if publisher == nil {
		return
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	event := audit.Event{
		Timestamp: time.Now().Unix(),
		Metrics:   metrics,
		IPAddress: ip,
	}
	if err := publisher.Publish(r.Context(), event); err != nil {
		logger.Errorw("Failed to publish audit event", "error", err)
	}
}
