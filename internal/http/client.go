package http

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"time"
)

type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

type Client struct {
	client *http.Client
	logger *slog.Logger
	config RetryConfig
}

func NewClient(client *http.Client, logger *slog.Logger, config RetryConfig) *Client {
	return &Client{
		client: client,
		logger: logger,
		config: config,
	}
}

func (c *Client) Do(ctx context.Context, req *http.Request) ([]byte, error) {
	if c.logger != nil {
		c.logger.Info("DO: sending request")
	}

	return doRequest(ctx, c.client, req, c.logger, c.config)
}

func doRequest(ctx context.Context, client *http.Client, req *http.Request, logger *slog.Logger, cfg RetryConfig) ([]byte, error) {
    var lastErr error
    if logger != nil {
        logger.Info("DOrequest: sending request", "url", req.URL.String(), "method", req.Method)
    }

    for i := 0; i < cfg.MaxRetries; i++ {
        if logger != nil {
            logger.Info("sending request", "attempt", i+1)
        }

        select {
        case <-ctx.Done():
            return nil, fmt.Errorf("context canceled: %w", ctx.Err())
        default:
            if logger != nil {
                logger.Debug("sending request", "attempt", i+1)
            }

            res, err := client.Do(req.WithContext(ctx))
            if err != nil {
                lastErr = fmt.Errorf("failed to send request: %w", err)
                delay := time.Duration(math.Min(
                    float64(cfg.BaseDelay)*math.Pow(2, float64(i)),
                    float64(cfg.MaxDelay),
                ))

                timer := time.NewTimer(delay)
                select {
                case <-ctx.Done():
                    timer.Stop()
                    return nil, fmt.Errorf("context canceled during retry: %w", ctx.Err())
                case <-timer.C:
                }
                continue
            }

            if res == nil {
                lastErr = fmt.Errorf("response is nil")
                if logger != nil {
                    logger.Error("response is nil", "attempt", i+1)
                }
                continue
            }
            defer res.Body.Close()

            body, err := io.ReadAll(res.Body)
            if err != nil {
                lastErr = fmt.Errorf("failed to read response body: %w", err)
                continue
            }

            // Логируем статус и тело ответа
            if logger != nil {
                logger.Info("received response", "status", res.StatusCode, "body", string(body))
            }

            if res.StatusCode != http.StatusOK {
                if logger != nil {
                    logger.Error("request failed", "attempt", i+1, "status", res.StatusCode, "error", lastErr)
                }
                lastErr = fmt.Errorf("unexpected status code %d: %s", res.StatusCode, string(body))
                continue
            }

            if logger != nil {
                logger.Debug("request completed successfully", "attempt", i+1)
            }
            return body, nil
        }
    }

    if lastErr != nil {
        return nil, fmt.Errorf("max retries exceeded, but no specific error was recorded")
    }
    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
