// Package client provides a Twilio API client for internal use.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Client is a Twilio API client.
type Client struct {
	accountSID string
	authToken  string
	baseURL    string
	httpClient *http.Client
}

// Config configures the Twilio client.
type Config struct {
	AccountSID string
	AuthToken  string
	BaseURL    string
	HTTPClient *http.Client
}

// New creates a new Twilio client.
func New(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	accountSID := cfg.AccountSID
	if accountSID == "" {
		accountSID = os.Getenv("TWILIO_ACCOUNT_SID")
	}
	if accountSID == "" {
		return nil, fmt.Errorf("TWILIO_ACCOUNT_SID is required")
	}

	authToken := cfg.AuthToken
	if authToken == "" {
		authToken = os.Getenv("TWILIO_AUTH_TOKEN")
	}
	if authToken == "" {
		return nil, fmt.Errorf("TWILIO_AUTH_TOKEN is required")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.twilio.com/2010-04-01"
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &Client{
		accountSID: accountSID,
		authToken:  authToken,
		baseURL:    baseURL,
		httpClient: httpClient,
	}, nil
}

// AccountSID returns the account SID.
func (c *Client) AccountSID() string {
	return c.accountSID
}

// Call represents a Twilio call resource.
type Call struct {
	SID         string `json:"sid"`
	AccountSID  string `json:"account_sid"`
	To          string `json:"to"`
	From        string `json:"from"`
	Status      string `json:"status"`
	Direction   string `json:"direction"`
	Duration    string `json:"duration"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Price       string `json:"price"`
	PriceUnit   string `json:"price_unit"`
	AnsweredBy  string `json:"answered_by"`
	CallerName  string `json:"caller_name"`
	URI         string `json:"uri"`
	DateCreated string `json:"date_created"`
	DateUpdated string `json:"date_updated"`
}

// MakeCallParams are parameters for making a call.
type MakeCallParams struct {
	To                  string
	From                string
	URL                 string            // TwiML URL
	Twiml               string            // Inline TwiML
	StatusCallback      string            // Webhook for status updates
	StatusCallbackEvent []string          // Events to receive
	MachineDetection    string            // "Enable" or "DetectMessageEnd"
	Timeout             int               // Ring timeout in seconds
	Record              bool              // Record the call
	RecordingChannels   string            // "mono" or "dual"
	CustomParameters    map[string]string // Custom parameters
}

// MakeCall initiates an outbound call.
func (c *Client) MakeCall(ctx context.Context, params *MakeCallParams) (*Call, error) {
	endpoint := fmt.Sprintf("%s/Accounts/%s/Calls.json", c.baseURL, c.accountSID)

	data := url.Values{}
	data.Set("To", params.To)
	data.Set("From", params.From)

	if params.URL != "" {
		data.Set("Url", params.URL)
	}
	if params.Twiml != "" {
		data.Set("Twiml", params.Twiml)
	}
	if params.StatusCallback != "" {
		data.Set("StatusCallback", params.StatusCallback)
	}
	for _, event := range params.StatusCallbackEvent {
		data.Add("StatusCallbackEvent", event)
	}
	if params.MachineDetection != "" {
		data.Set("MachineDetection", params.MachineDetection)
	}
	if params.Timeout > 0 {
		data.Set("Timeout", fmt.Sprintf("%d", params.Timeout))
	}
	if params.Record {
		data.Set("Record", "true")
	}
	if params.RecordingChannels != "" {
		data.Set("RecordingChannels", params.RecordingChannels)
	}
	for k, v := range params.CustomParameters {
		data.Set(k, v)
	}

	var call Call
	if err := c.post(ctx, endpoint, data, &call); err != nil {
		return nil, err
	}
	return &call, nil
}

// GetCall retrieves a call by SID.
func (c *Client) GetCall(ctx context.Context, callSID string) (*Call, error) {
	endpoint := fmt.Sprintf("%s/Accounts/%s/Calls/%s.json", c.baseURL, c.accountSID, callSID)

	var call Call
	if err := c.get(ctx, endpoint, &call); err != nil {
		return nil, err
	}
	return &call, nil
}

// UpdateCallParams are parameters for updating a call.
type UpdateCallParams struct {
	URL    string // New TwiML URL
	Twiml  string // Inline TwiML
	Status string // "completed" to hang up, "canceled" to cancel
}

// UpdateCall modifies an in-progress call.
func (c *Client) UpdateCall(ctx context.Context, callSID string, params *UpdateCallParams) (*Call, error) {
	endpoint := fmt.Sprintf("%s/Accounts/%s/Calls/%s.json", c.baseURL, c.accountSID, callSID)

	data := url.Values{}
	if params.URL != "" {
		data.Set("Url", params.URL)
	}
	if params.Twiml != "" {
		data.Set("Twiml", params.Twiml)
	}
	if params.Status != "" {
		data.Set("Status", params.Status)
	}

	var call Call
	if err := c.post(ctx, endpoint, data, &call); err != nil {
		return nil, err
	}
	return &call, nil
}

// HangupCall ends a call.
func (c *Client) HangupCall(ctx context.Context, callSID string) (*Call, error) {
	return c.UpdateCall(ctx, callSID, &UpdateCallParams{Status: "completed"})
}

// PhoneNumber represents a Twilio phone number.
type PhoneNumber struct {
	SID          string `json:"sid"`
	PhoneNumber  string `json:"phone_number"`
	FriendlyName string `json:"friendly_name"`
	Capabilities struct {
		Voice bool `json:"voice"`
		SMS   bool `json:"sms"`
		MMS   bool `json:"mms"`
	} `json:"capabilities"`
}

// PhoneNumberList is a list of phone numbers.
type PhoneNumberList struct {
	PhoneNumbers []PhoneNumber `json:"incoming_phone_numbers"`
}

// ListPhoneNumbers returns all phone numbers on the account.
func (c *Client) ListPhoneNumbers(ctx context.Context) ([]PhoneNumber, error) {
	endpoint := fmt.Sprintf("%s/Accounts/%s/IncomingPhoneNumbers.json", c.baseURL, c.accountSID)

	var list PhoneNumberList
	if err := c.get(ctx, endpoint, &list); err != nil {
		return nil, err
	}
	return list.PhoneNumbers, nil
}

// Error represents a Twilio API error.
type Error struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	MoreInfo string `json:"more_info"`
	Status   int    `json:"status"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("twilio error %d: %s", e.Code, e.Message)
}

// get performs a GET request.
func (c *Client) get(ctx context.Context, url string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	return c.do(req, result)
}

// post performs a POST request with form data.
func (c *Client) post(ctx context.Context, url string, data url.Values, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.do(req, result)
}

// do executes a request with authentication.
func (c *Client) do(req *http.Request, result any) error {
	req.SetBasicAuth(c.accountSID, c.authToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var apiErr Error
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return fmt.Errorf("twilio error: %s", string(body))
		}
		return &apiErr
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}
