package ports

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/ibgl/microservice-users/internal/app/mocks"
	"github.com/ibgl/microservice-users/internal/app/user"
	"github.com/stretchr/testify/mock"
)

func Test_signIn(t *testing.T) {
	mockUUID := uuid.New()
	tt := []struct {
		name            string
		method          string
		serviceRequest  *user.SignInRequest
		serviceResponse *user.LoginResponse
		serviceError    error
		body            string
		want            string
		statusCode      int
	}{
		{
			name:   "With a valid username and password",
			method: http.MethodPost,
			serviceRequest: &user.SignInRequest{
				Email:    "email",
				Password: "password",
			},
			serviceResponse: &user.LoginResponse{
				Access: user.Token{
					Value:  "access",
					UserId: mockUUID,
				},
				Refresh: user.Token{
					Value:  "refresh",
					UserId: mockUUID,
				},
			},
			serviceError: nil,
			body:         `{"email":"email","password":"password"}`,
			want:         `{"access":"access","refresh":"refresh"}`,
			statusCode:   http.StatusOK,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, "/api/v1/signIn", strings.NewReader(tc.body))
			responseRecorder := httptest.NewRecorder()

			authMock := new(mocks.AuthServiceMock)
			authMock.On("SignIn", mock.Anything, tc.serviceRequest).Return(tc.serviceResponse, tc.serviceError)

			app := mocks.NewAppMock(nil).SetAuthService(authMock)
			server := NewHttpServer(app)

			handler := server.signIn
			handler(responseRecorder, request)

			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			if strings.TrimSpace(responseRecorder.Body.String()) != tc.want {
				t.Errorf("Want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}
