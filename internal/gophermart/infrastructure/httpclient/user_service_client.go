package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type UserServiceClient struct {
	baseURL string
	client  *http.Client
}

func NewUserServiceClient(baseURL string) *UserServiceClient {
	return &UserServiceClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type Claims struct {
	UserID int64  `json:"user_id"`
	Login  string `json:"login"`
}

type ValidateResponse struct {
	UserID int64  `json:"user_id"`
	Login  string `json:"login"`
}

func (c *UserServiceClient) ValidateToken(ctx context.Context, token string) (*Claims, error) {
	url := fmt.Sprintf("%s/api/auth/validate", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("invalid token")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var validateResp ValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&validateResp); err != nil {
		return nil, err
	}

	return &Claims{
		UserID: validateResp.UserID,
		Login:  validateResp.Login,
	}, nil
}

