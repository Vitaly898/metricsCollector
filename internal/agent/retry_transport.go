package agent

import (
	"net/http"
	"time"
)

var defaultRetryIntervals = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

type retryRoundTripper struct {
	next      http.RoundTripper
	intervals []time.Duration
}

func NewRetryRoundTripper(next http.RoundTripper, intervals []time.Duration) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	return &retryRoundTripper{next: next, intervals: intervals}
}

func (rt *retryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := 0; i <= len(rt.intervals); i++ {
		attempt := req
		if i > 0 {
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(rt.intervals[i-1]):
			}

			if req.GetBody != nil {
				body, berr := req.GetBody()
				if berr != nil {
					return nil, berr
				}
				attempt = req.Clone(req.Context())
				attempt.Body = body
			}
		}

		resp, err = rt.next.RoundTrip(attempt)
		if err == nil && resp.StatusCode < http.StatusInternalServerError {
			return resp, nil
		}
		if err == nil {
			resp.Body.Close()
		}
	}

	return resp, err
}
