package flows_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	. "github.com/onsi/gomega"

	"github.com/ya-breeze/diary.be/pkg/auth"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/generated/goclient"
	"github.com/ya-breeze/diary.be/pkg/server"
)

// SharedTestSetup contains all the shared test infrastructure
//
//nolint:containedctx
type SharedTestSetup struct {
	Logger     *slog.Logger
	Cfg        *config.Config
	Storage    database.Storage
	ServerAddr string
	APIClient  *goclient.APIClient
	Ctx        context.Context
	Cancel     context.CancelFunc
	TestEmail  string
	TestPass   string
	TempDir    string
}

// Helper to create cancellable context outside function literal
func newCancellableContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// SetupTestEnvironment creates and configures the shared test environment
func SetupTestEnvironment() *SharedTestSetup {
	setup := &SharedTestSetup{}

	setup.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create temporary directory for assets
	var err error
	setup.TempDir, err = os.MkdirTemp("", "flow_test_assets")
	Expect(err).NotTo(HaveOccurred())

	setup.Cfg = &config.Config{
		Port:             0, // Use random available port
		DataPath:         setup.TempDir,
		Issuer:           "test-issuer",
		JWTSecret:        "test-secret-key-for-jwt-tokens",
		SessionSecret:    "test-session-secret-key-minimum-32-characters-long",
		DisableRateLimit: true,  // Disable rate limiting for tests
		CookieSecure:     false, // Allow HTTP for tests
	}

	setup.Storage = database.NewStorage(setup.Logger, setup.Cfg)
	Expect(setup.Storage.Open()).To(Succeed())

	// Create test user
	setup.TestEmail = "test@test.com"
	setup.TestPass = "testpassword123"

	hashedPassBytes, err := auth.HashPassword([]byte(setup.TestPass))
	Expect(err).ToNot(HaveOccurred())
	hashedPass := base64.StdEncoding.EncodeToString(hashedPassBytes)

	_, err = setup.Storage.CreateUser(setup.TestEmail, hashedPass)
	Expect(err).ToNot(HaveOccurred())

	// Create context outside BeforeEach to avoid fatcontext linting issue
	setup.Ctx, setup.Cancel = newCancellableContext()

	// Start test server
	addr, _, err := server.Serve(setup.Ctx, setup.Logger, setup.Storage, setup.Cfg)
	Expect(err).ToNot(HaveOccurred())

	tcpAddr, ok := addr.(*net.TCPAddr)
	Expect(ok).To(BeTrue(), "Failed to cast address to *net.TCPAddr")
	setup.ServerAddr = fmt.Sprintf("http://localhost:%d", tcpAddr.Port)
	setup.Logger.Info("Test server started", "address", setup.ServerAddr)

	// Wait for server to be ready by polling the authorize endpoint
	Eventually(func() bool {
		//nolint
		resp, err := http.Post(setup.ServerAddr+"/v1/authorize", "application/json", nil)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		// We expect 400 (bad request) because we're not sending valid JSON,
		// but this means the server is up and responding
		return resp.StatusCode == http.StatusBadRequest
	}, "5s", "100ms").Should(BeTrue())

	// Create API client
	clientConfig := goclient.NewConfiguration()
	clientConfig.Servers = goclient.ServerConfigurations{
		{
			URL:         setup.ServerAddr,
			Description: "Test server",
		},
	}

	setup.APIClient = goclient.NewAPIClient(clientConfig)

	return setup
}

// TeardownTestEnvironment cleans up the test environment
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

// LoginAndGetToken performs login and returns the JWT token
func (setup *SharedTestSetup) LoginAndGetToken() string {
	authData := goclient.AuthData{
		Email:    setup.TestEmail,
		Password: setup.TestPass,
	}

	authResponse, httpResponse, err := setup.APIClient.AuthAPI.Authorize(context.Background()).AuthData(authData).Execute()
	Expect(err).ToNot(HaveOccurred())
	defer httpResponse.Body.Close()
	Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
	Expect(authResponse.Token).ToNot(BeEmpty())

	// Configure client with JWT token for subsequent requests
	clientConfig := setup.APIClient.GetConfig()
	clientConfig.AddDefaultHeader("Authorization", "Bearer "+authResponse.Token)

	return authResponse.Token
}
