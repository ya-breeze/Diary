package flows_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auth Error Flows", func() {
	var setup *SharedTestSetup

	BeforeEach(func() {
		setup = SetupTestEnvironment()
	})

	AfterEach(func() {
		setup.TeardownTestEnvironment()
	})

	Describe("Authorization endpoint", func() {
		Context("when credentials are wrong", func() {
			It("returns 401 for incorrect password", func() {
				_, httpResp, err := setup.APIClient.Authorize(
					context.Background(), setup.TestEmail, "wrong-password",
				)
				Expect(err).To(HaveOccurred())
				Expect(httpResp).NotTo(BeNil())
				Expect(httpResp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("returns 401 for unknown email", func() {
				_, httpResp, err := setup.APIClient.Authorize(
					context.Background(), "nobody@example.com", setup.TestPass,
				)
				Expect(err).To(HaveOccurred())
				Expect(httpResp).NotTo(BeNil())
				Expect(httpResp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("Protected endpoints", func() {
		Context("when no Authorization header is sent", func() {
			It("returns 401 for GET /v1/items", func() {
				// Do not call LoginAndGetToken — client has no token
				_, httpResp, err := setup.APIClient.GetItems(context.Background(), "", "", "")
				Expect(err).To(HaveOccurred())
				Expect(httpResp).NotTo(BeNil())
				Expect(httpResp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when a malformed token is sent", func() {
			It("returns 401 for GET /v1/items", func() {
				setup.APIClient.SetToken("not-a-valid-jwt")
				_, httpResp, err := setup.APIClient.GetItems(context.Background(), "", "", "")
				Expect(err).To(HaveOccurred())
				Expect(httpResp).NotTo(BeNil())
				Expect(httpResp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when a structurally valid but wrong-secret token is sent", func() {
			It("returns 401 for GET /v1/items", func() {
				// JWT signed with a different secret
				setup.APIClient.SetToken(
					"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
						"eyJzdWIiOiIxMjM0NTY3ODkwIn0." +
						"SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
				)
				_, httpResp, err := setup.APIClient.GetItems(context.Background(), "", "", "")
				Expect(err).To(HaveOccurred())
				Expect(httpResp).NotTo(BeNil())
				Expect(httpResp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})
	})
})
