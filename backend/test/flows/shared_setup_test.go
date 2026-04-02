package flows_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/gomega"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/ya-breeze/diary.be/pkg/auth"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/generated/goclient"
	"github.com/ya-breeze/diary.be/pkg/server"
)

// --- Flat response types for test assertions (use strings for dates) ---

type TestAuthResponse struct {
	Token string
}

type TestItemsResponse struct {
	Date         string
	Title        string
	Body         string
	Tags         []string
	PreviousDate *string
	NextDate     *string
}

type TestItemsListResponse struct {
	Items      []TestItemsResponse
	TotalCount int
}

type TestSyncChangeResponse struct {
	Id            int32
	UserId        string
	Date          string
	OperationType string
	Timestamp     time.Time
	ItemSnapshot  *TestItemsResponse
	Metadata      []string
}

type TestSyncResponse struct {
	Changes []TestSyncChangeResponse
	HasMore bool
	NextId  *int32
}

type TestAssetsBatchFile struct {
	SavedName    string
	OriginalName string
}

type TestAssetsBatchResponse struct {
	Count int
	Files []TestAssetsBatchFile
}

// --- TestAPIClient ---

// TestAPIClient wraps the low-level goclient.Client with helper methods that
// accept/return simple Go types (string dates, plain slices) for easier test assertions.
type TestAPIClient struct {
	serverAddr string
	httpClient *http.Client
	authHeader string
}

func newTestAPIClient(serverAddr string) *TestAPIClient {
	return &TestAPIClient{
		serverAddr: serverAddr,
		httpClient: &http.Client{},
	}
}

func (c *TestAPIClient) SetToken(token string) {
	c.authHeader = "Bearer " + token
}

func (c *TestAPIClient) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.serverAddr+path, bodyReader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.authHeader != "" {
		req.Header.Set("Authorization", c.authHeader)
	}
	return req, nil
}

func (c *TestAPIClient) do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

// Authorize logs in with email/password and returns the JWT token.
func (c *TestAPIClient) Authorize(ctx context.Context, email, password string) (*TestAuthResponse, *http.Response, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/v1/authorize",
		map[string]string{"email": email, "password": password})
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	var result goclient.AuthorizeJSONRequestBody // same shape as response
	if resp.StatusCode == http.StatusOK {
		var authResp struct{ Token string `json:"token"` }
		if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
			return nil, resp, err
		}
		_ = result
		return &TestAuthResponse{Token: authResp.Token}, resp, nil
	}
	return nil, resp, fmt.Errorf("authorize failed: HTTP %d", resp.StatusCode)
}

// PutItems saves a diary item. Returns the saved item and raw response.
func (c *TestAPIClient) PutItems(ctx context.Context, date, title, body string, tags []string) (*TestItemsResponse, *http.Response, error) {
	parsedDate, _ := time.Parse("2006-01-02", date)
	payload := goclient.ItemsRequest{
		Date:  openapi_types.Date{Time: parsedDate},
		Title: title,
		Body:  &body,
		Tags:  &tags,
	}
	req, err := c.newRequest(ctx, http.MethodPut, "/v1/items", payload)
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp, fmt.Errorf("PutItems failed: HTTP %d", resp.StatusCode)
	}
	var raw goclient.ItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, resp, err
	}
	return toTestItemsResponse(raw), resp, nil
}

// GetItems fetches diary items. All filter params are optional (empty string = unset).
func (c *TestAPIClient) GetItems(ctx context.Context, date, search, tags string) (*TestItemsListResponse, *http.Response, error) {
	params := url.Values{}
	if date != "" {
		params.Set("date", date)
	}
	if search != "" {
		params.Set("search", search)
	}
	if tags != "" {
		params.Set("tags", tags)
	}
	u := c.serverAddr + "/v1/items?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}
	if c.authHeader != "" {
		req.Header.Set("Authorization", c.authHeader)
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp, fmt.Errorf("GetItems failed: HTTP %d", resp.StatusCode)
	}
	var raw goclient.ItemsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, resp, err
	}
	result := &TestItemsListResponse{
		TotalCount: raw.TotalCount,
		Items:      make([]TestItemsResponse, len(raw.Items)),
	}
	for i, item := range raw.Items {
		result.Items[i] = *toTestItemsResponse(item)
	}
	return result, resp, nil
}

// GetAsset fetches an asset by path. Returns the raw response (caller closes body).
func (c *TestAPIClient) GetAsset(ctx context.Context, path string) (*http.Response, error) {
	params := url.Values{}
	params.Set("path", path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.serverAddr+"/v1/assets?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	if c.authHeader != "" {
		req.Header.Set("Authorization", c.authHeader)
	}
	return c.do(req)
}

// GetChanges fetches sync changes. since=0 means from the beginning; limit=0 means no limit.
func (c *TestAPIClient) GetChanges(ctx context.Context, since, limit int32) (*TestSyncResponse, *http.Response, error) {
	u := fmt.Sprintf("%s/v1/sync/changes?since=%d", c.serverAddr, since)
	if limit > 0 {
		u += fmt.Sprintf("&limit=%d", limit)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}
	if c.authHeader != "" {
		req.Header.Set("Authorization", c.authHeader)
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp, fmt.Errorf("GetChanges failed: HTTP %d", resp.StatusCode)
	}
	var raw goclient.SyncResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, resp, err
	}
	result := &TestSyncResponse{
		HasMore: raw.HasMore,
		NextId:  raw.NextId,
		Changes: make([]TestSyncChangeResponse, len(raw.Changes)),
	}
	for i, ch := range raw.Changes {
		result.Changes[i] = toTestSyncChange(ch)
	}
	return result, resp, nil
}

// UploadAssetsBatch uploads multiple files as a multipart form to /v1/assets/batch.
func (c *TestAPIClient) UploadAssetsBatch(ctx context.Context, files []*os.File) (*TestAssetsBatchResponse, *http.Response, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for _, f := range files {
		fw, err := w.CreateFormFile("assets", filepath.Base(f.Name()))
		if err != nil {
			return nil, nil, err
		}
		if _, err := io.Copy(fw, f); err != nil {
			return nil, nil, err
		}
	}
	w.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serverAddr+"/v1/assets/batch", &buf)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	if c.authHeader != "" {
		req.Header.Set("Authorization", c.authHeader)
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp, fmt.Errorf("UploadAssetsBatch failed: HTTP %d", resp.StatusCode)
	}
	var raw struct {
		Count int `json:"count"`
		Files []struct {
			SavedName    string `json:"savedName"`
			OriginalName string `json:"originalName"`
		} `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, resp, err
	}
	result := &TestAssetsBatchResponse{Count: raw.Count, Files: make([]TestAssetsBatchFile, len(raw.Files))}
	for i, f := range raw.Files {
		result.Files[i] = TestAssetsBatchFile{SavedName: f.SavedName, OriginalName: f.OriginalName}
	}
	return result, resp, nil
}

// --- conversion helpers ---

func toTestItemsResponse(r goclient.ItemsResponse) *TestItemsResponse {
	t := &TestItemsResponse{
		Date:  r.Date.Time.Format("2006-01-02"),
		Title: r.Title,
	}
	if r.Body != nil {
		t.Body = *r.Body
	}
	if r.Tags != nil {
		t.Tags = *r.Tags
	}
	if r.PreviousDate != nil {
		s := r.PreviousDate.Time.Format("2006-01-02")
		t.PreviousDate = &s
	}
	if r.NextDate != nil {
		s := r.NextDate.Time.Format("2006-01-02")
		t.NextDate = &s
	}
	return t
}

func toTestSyncChange(ch goclient.SyncChangeResponse) TestSyncChangeResponse {
	r := TestSyncChangeResponse{
		Id:            ch.Id,
		UserId:        ch.UserId,
		Date:          ch.Date.Time.Format("2006-01-02"),
		OperationType: string(ch.OperationType),
		Timestamp:     ch.Timestamp,
	}
	if ch.Metadata != nil {
		r.Metadata = *ch.Metadata
	}
	if ch.ItemSnapshot != nil {
		r.ItemSnapshot = toTestItemsResponse(*ch.ItemSnapshot)
	}
	return r
}

// --- SharedTestSetup ---

//nolint:containedctx
type SharedTestSetup struct {
	Logger     *slog.Logger
	Cfg        *config.Config
	Storage    database.Storage
	ServerAddr string
	APIClient  *TestAPIClient
	Ctx        context.Context
	Cancel     context.CancelFunc
	TestEmail  string
	TestPass   string
	TempDir    string
}

func newCancellableContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

func SetupTestEnvironment() *SharedTestSetup {
	setup := &SharedTestSetup{}

	setup.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	var err error
	setup.TempDir, err = os.MkdirTemp("", "flow_test_assets")
	Expect(err).NotTo(HaveOccurred())

	setup.Cfg = &config.Config{
		Port:             0,
		DataPath:         setup.TempDir,
		Issuer:           "test-issuer",
		JWTSecret:        "test-secret-key-for-jwt-tokens",
		SessionSecret:    "test-session-secret-key-minimum-32-characters-long",
		DisableRateLimit: true,
		CookieSecure:     false,
	}

	setup.Storage = database.NewStorage(setup.Logger, setup.Cfg)
	Expect(setup.Storage.Open()).To(Succeed())

	setup.TestEmail = "test@test.com"
	setup.TestPass = "testpassword123"

	hashedPassBytes, err := auth.HashPassword([]byte(setup.TestPass))
	Expect(err).ToNot(HaveOccurred())
	hashedPass := base64.StdEncoding.EncodeToString(hashedPassBytes)

	_, err = setup.Storage.CreateUser(setup.TestEmail, hashedPass)
	Expect(err).ToNot(HaveOccurred())

	setup.Ctx, setup.Cancel = newCancellableContext()

	addr, _, err := server.Serve(setup.Ctx, setup.Logger, setup.Storage, setup.Cfg)
	Expect(err).ToNot(HaveOccurred())

	tcpAddr, ok := addr.(*net.TCPAddr)
	Expect(ok).To(BeTrue())
	setup.ServerAddr = fmt.Sprintf("http://localhost:%d", tcpAddr.Port)
	setup.Logger.Info("Test server started", "address", setup.ServerAddr)

	// Wait for server to be ready
	Eventually(func() bool {
		//nolint:noctx
		resp, err := http.Post(setup.ServerAddr+"/v1/authorize", "application/json", nil)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusBadRequest
	}, "5s", "100ms").Should(BeTrue())

	setup.APIClient = newTestAPIClient(setup.ServerAddr)

	return setup
}

func (setup *SharedTestSetup) TeardownTestEnvironment() {
	if setup.Cancel != nil {
		setup.Cancel()
	}
	if setup.Storage != nil {
		setup.Storage.Close()
	}
	if setup.TempDir != "" {
		os.RemoveAll(setup.TempDir)
	}
}

// LoginAndGetToken authenticates and configures the API client with the token.
func (setup *SharedTestSetup) LoginAndGetToken() string {
	authResp, httpResp, err := setup.APIClient.Authorize(
		context.Background(), setup.TestEmail, setup.TestPass,
	)
	Expect(err).ToNot(HaveOccurred())
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
	Expect(authResp.Token).ToNot(BeEmpty())

	setup.APIClient.SetToken(authResp.Token)
	return authResp.Token
}
