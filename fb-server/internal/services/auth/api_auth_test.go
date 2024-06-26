package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	mock_repo "fightbettr.com/fb-server/internal/repo/auth/mocks"
	mock_tx "fightbettr.com/fb-server/internal/repo/mocs"
	"fightbettr.com/fb-server/internal/services"
	mock_logger "fightbettr.com/fb-server/pkg/logger/mocks"
	"fightbettr.com/fb-server/pkg/model"
	"fightbettr.com/fb-server/pkg/utils"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRegister(t *testing.T) {
	tests := []struct {
		name           string
		mockBehavior   func(ctx context.Context, mockRepo *mock_repo.MockFbAuthRepo, mockTx *mock_tx.MockTestTx, mockLogger *mock_logger.MockFbLogger)
		req            *http.Request
		expectedStatus int
	}{
		{
			name: "Success",
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {
				user := model.User{
					Name:      "Test",
					CreatedAt: time.Now().Unix(),
				}

				mrepo.EXPECT().BeginTx(gomock.Any(), pgx.TxOptions{
					IsoLevel: pgx.Serializable,
				}).Return(mtx, nil)
				mrepo.EXPECT().TxCreateUser(gomock.Any(), mtx, user)
				mrepo.EXPECT().TxNewAuthCredentials(gomock.Any(), mtx, gomock.Any())
				mtx.EXPECT().Commit(gomock.Any())
			},
			req: (func() *http.Request {
				token, err := getFakeToken()
				require.NoError(t, err)

				registerReq := model.RegisterRequest{
					Email:    "test@gmail.com",
					Password: "12345qwerty",
					Name:     "Test",
					TermsOk:  true,
					Token:    "12edaws1i2hrj1h2vgv1fvj3v5j23j5",
				}

				return createFakeRequestWithBody(token, registerReq)
			})(),
			expectedStatus: http.StatusOK,
		},
		{
			name: "Bad request because of empty body",
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {

			},
			req:            httptest.NewRequest("POST", "/example", nil),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "TermsOk false",
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {

			},
			req: (func() *http.Request {
				token, err := getFakeToken()
				require.NoError(t, err)

				registerReq := model.RegisterRequest{
					Email:    "test@gmail.com",
					Password: "12345qwerty",
					Name:     "Test",
					TermsOk:  false,
					Token:    "12edaws1i2hrj1h2vgv1fvj3v5j23j5",
				}

				return createFakeRequestWithBody(token, registerReq)
			})(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "BeginTx error",
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {
				mrepo.EXPECT().BeginTx(gomock.Any(), pgx.TxOptions{
					IsoLevel: pgx.Serializable,
				}).Return(nil, errors.New("error"))

				mlogger.EXPECT().Errorf("Unable to begin transaction: %s", errors.New("error"))
			},
			req: (func() *http.Request {
				token, err := getFakeToken()
				require.NoError(t, err)

				registerReq := model.RegisterRequest{
					Email:    "test@gmail.com",
					Password: "12345qwerty",
					Name:     "Test",
					TermsOk:  true,
					Token:    "12edaws1i2hrj1h2vgv1fvj3v5j23j5",
				}

				return createFakeRequestWithBody(token, registerReq)
			})(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Tx Commit error",
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {
				user := model.User{
					Name:      "Test",
					CreatedAt: time.Now().Unix(),
				}

				mrepo.EXPECT().BeginTx(gomock.Any(), pgx.TxOptions{
					IsoLevel: pgx.Serializable,
				}).Return(mtx, nil)
				mrepo.EXPECT().TxCreateUser(gomock.Any(), mtx, user)
				mrepo.EXPECT().TxNewAuthCredentials(gomock.Any(), mtx, gomock.Any())

				mtx.EXPECT().Commit(gomock.Any()).Return(errors.New("error"))

				mlogger.EXPECT().Errorf("Unable to commit transaction: %s", errors.New("error"))
			},
			req: (func() *http.Request {
				token, err := getFakeToken()
				require.NoError(t, err)

				registerReq := model.RegisterRequest{
					Email:    "test@gmail.com",
					Password: "12345qwerty",
					Name:     "Test",
					TermsOk:  true,
					Token:    "12edaws1i2hrj1h2vgv1fvj3v5j23j5",
				}

				return createFakeRequestWithBody(token, registerReq)
			})(),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			mockRepo := mock_repo.NewMockFbAuthRepo(ctrl)
			mockLogger := mock_logger.NewMockFbLogger(ctrl)
			mockTx := mock_tx.NewMockTestTx(ctrl)

			handler := &services.ApiHandler{
				Logger: mockLogger,
			}

			service := &service{
				Repo:       mockRepo,
				ApiHandler: handler,
			}

			w := httptest.NewRecorder()

			tc.mockBehavior(ctx, mockRepo, mockTx, mockLogger)

			service.Register(w, tc.req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestConfirmRegistration(t *testing.T) {
	tests := []struct {
		name           string
		mockBehavior   func(ctx context.Context, mockRepo *mock_repo.MockFbAuthRepo, mockTx *mock_tx.MockTestTx, mockLogger *mock_logger.MockFbLogger)
		req            *http.Request
		expectedStatus int
	}{
		{
			name: "Success",
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {
				credsReq := model.UserCredentialsRequest{
					Token: "fake_token",
				}

				credsReqUpdated := model.UserCredentialsRequest{
					UserId:    1,
					Token:     "fake_token",
					TokenType: model.TokenConfirmation,
				}

				credsResp := model.UserCredentials{
					UserId:      1,
					Email:       "test@gmail.com",
					Password:    "12345qwerty",
					Salt:        "fake_salt",
					Token:       "fake_token",
					TokenType:   model.TokenConfirmation,
					TokenExpire: time.Now().Unix() + 60*60*48,
					Active:      true,
				}

				mrepo.EXPECT().FindUserCredentials(gomock.Any(), credsReq).Return(credsResp, nil)
				mrepo.EXPECT().ConfirmCredentialsToken(gomock.Any(), nil, credsReqUpdated).Return(nil)
			},
			req: (func() *http.Request {
				r := httptest.NewRequest("POST", "/example", nil)
				query := r.URL.Query()
				query.Add("token", "fake_token")
				r.URL.RawQuery = query.Encode()

				return r
			})(),
			expectedStatus: http.StatusOK,
		},
		{
			name: "Empty token in request",
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {

			},
			req: (func() *http.Request {
				r := httptest.NewRequest("POST", "/example", nil)
				query := r.URL.Query()
				query.Add("token", "")
				r.URL.RawQuery = query.Encode()

				return r
			})(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "FindUserCredentials no rows error",
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {
				credsReq := model.UserCredentialsRequest{
					Token: "fake_token",
				}

				mrepo.EXPECT().FindUserCredentials(gomock.Any(), credsReq).Return(model.UserCredentials{}, pgx.ErrNoRows)
			},
			req: (func() *http.Request {
				r := httptest.NewRequest("POST", "/example", nil)
				query := r.URL.Query()
				query.Add("token", "fake_token")
				r.URL.RawQuery = query.Encode()

				return r
			})(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "FindUserCredentials internal error",
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {
				credsReq := model.UserCredentialsRequest{
					Token: "fake_token",
				}

				mrepo.EXPECT().FindUserCredentials(gomock.Any(), credsReq).Return(model.UserCredentials{}, errors.New("error"))
				mlogger.EXPECT().Errorf("Failed to get user credentials: %s", errors.New("error"))
			},
			req: (func() *http.Request {
				r := httptest.NewRequest("POST", "/example", nil)
				query := r.URL.Query()
				query.Add("token", "fake_token")
				r.URL.RawQuery = query.Encode()

				return r
			})(),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Expired Token",
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {
				credsReq := model.UserCredentialsRequest{
					Token: "fake_token",
				}

				credsResp := model.UserCredentials{
					UserId:      1,
					Email:       "test@gmail.com",
					Password:    "12345qwerty",
					Salt:        "fake_salt",
					Token:       "fake_token",
					TokenType:   model.TokenConfirmation,
					TokenExpire: time.Now().Unix() - 60*60*48,
					Active:      true,
				}

				mrepo.EXPECT().FindUserCredentials(gomock.Any(), credsReq).Return(credsResp, nil)
			},
			req: (func() *http.Request {
				r := httptest.NewRequest("POST", "/example", nil)
				query := r.URL.Query()
				query.Add("token", "fake_token")
				r.URL.RawQuery = query.Encode()

				return r
			})(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ConfirmCredentialsToken error",
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {
				credsReq := model.UserCredentialsRequest{
					Token: "fake_token",
				}

				credsReqUpdated := model.UserCredentialsRequest{
					UserId:    1,
					Token:     "fake_token",
					TokenType: model.TokenConfirmation,
				}

				credsResp := model.UserCredentials{
					UserId:      1,
					Email:       "test@gmail.com",
					Password:    "12345qwerty",
					Salt:        "fake_salt",
					Token:       "fake_token",
					TokenType:   model.TokenConfirmation,
					TokenExpire: time.Now().Unix() + 60*60*48,
					Active:      true,
				}

				mrepo.EXPECT().FindUserCredentials(gomock.Any(), credsReq).Return(credsResp, nil)
				mrepo.EXPECT().ConfirmCredentialsToken(gomock.Any(), nil, credsReqUpdated).Return(errors.New("error"))

				mlogger.EXPECT().Errorf("Failed to update user credentials: %s", errors.New("error"))
			},
			req: (func() *http.Request {
				r := httptest.NewRequest("POST", "/example", nil)
				query := r.URL.Query()
				query.Add("token", "fake_token")
				r.URL.RawQuery = query.Encode()

				return r
			})(),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			mockRepo := mock_repo.NewMockFbAuthRepo(ctrl)
			mockLogger := mock_logger.NewMockFbLogger(ctrl)
			mockTx := mock_tx.NewMockTestTx(ctrl)

			handler := &services.ApiHandler{
				Logger: mockLogger,
			}

			service := &service{
				Repo:       mockRepo,
				ApiHandler: handler,
			}

			w := httptest.NewRecorder()

			tc.mockBehavior(ctx, mockRepo, mockTx, mockLogger)

			service.ConfirmRegistration(w, tc.req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name           string
		mockBehavior   func(ctx context.Context, mockRepo *mock_repo.MockFbAuthRepo, mockTx *mock_tx.MockTestTx, mockLogger *mock_logger.MockFbLogger)
		req            *http.Request
		expectedStatus int
	}{
		{
			name: "Success",
			req: (func() *http.Request {
				token, err := getFakeToken()
				require.NoError(t, err)

				registerReq := model.AuthenticateRequest{
					Email:    "test@gmail.com",
					Password: "12345qwerty",
				}

				return createFakeRequestWithBody(token, registerReq)
			})(),
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {
				password := "12345qwerty"
				salt := "123qwer123"
				fakePassword := utils.GenerateSaltedHash(password, salt)

				userCredsReq := model.UserCredentialsRequest{
					Email: "test@gmail.com",
				}
				userCreds := model.UserCredentials{
					UserId:   1,
					Active:   true,
					Salt:     "123qwer123",
					Password: fakePassword,
				}
				userReq := &model.UserRequest{
					UserId: 1,
				}
				user := &model.User{UserId: 1}

				loadJwtCerts()

				mrepo.EXPECT().FindUserCredentials(gomock.Any(), userCredsReq).Return(userCreds, nil)
				mrepo.EXPECT().FindUser(gomock.Any(), userReq).Return(user, nil)

				mlogger.EXPECT().Debugf("Issuing JWT token for User [%d:%s:%s]", userCreds.UserId, userCreds.Email, gomock.Any())
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Bad request because of empty body",
			req:  httptest.NewRequest("POST", "/example", nil),
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {

			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Email is empty",
			req: (func() *http.Request {
				token, err := getFakeToken()
				require.NoError(t, err)

				registerReq := model.AuthenticateRequest{
					Email:    "",
					Password: "12345qwerty",
				}

				return createFakeRequestWithBody(token, registerReq)
			})(),
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {

			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Password is empty",
			req: (func() *http.Request {
				token, err := getFakeToken()
				require.NoError(t, err)

				registerReq := model.AuthenticateRequest{
					Email:    "test@gmail.com",
					Password: "",
				}

				return createFakeRequestWithBody(token, registerReq)
			})(),
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {

			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "FindUserCredentials error",
			req: (func() *http.Request {
				token, err := getFakeToken()
				require.NoError(t, err)

				registerReq := model.AuthenticateRequest{
					Email:    "test@gmail.com",
					Password: "12345qwerty",
				}

				return createFakeRequestWithBody(token, registerReq)
			})(),
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {
				expectedError := errors.New("Error")
				userCredsReq := model.UserCredentialsRequest{
					Email: "test@gmail.com",
				}
				mrepo.EXPECT().FindUserCredentials(gomock.Any(), userCredsReq).Return(model.UserCredentials{}, expectedError)
				mlogger.EXPECT().Errorf("Failed to get user credentials: %s", expectedError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "FindUserCredentials NoRows error",
			req: (func() *http.Request {
				token, err := getFakeToken()
				require.NoError(t, err)

				registerReq := model.AuthenticateRequest{
					Email:    "test@gmail.com",
					Password: "12345qwerty",
				}

				return createFakeRequestWithBody(token, registerReq)
			})(),
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {
				expectedError := pgx.ErrNoRows
				userCredsReq := model.UserCredentialsRequest{
					Email: "test@gmail.com",
				}
				mrepo.EXPECT().FindUserCredentials(gomock.Any(), userCredsReq).Return(model.UserCredentials{}, expectedError)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Creds is not active",
			req: (func() *http.Request {
				token, err := getFakeToken()
				require.NoError(t, err)

				registerReq := model.AuthenticateRequest{
					Email:    "test@gmail.com",
					Password: "12345qwerty",
				}

				return createFakeRequestWithBody(token, registerReq)
			})(),
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {
				userCredsReq := model.UserCredentialsRequest{
					Email: "test@gmail.com",
				}
				mrepo.EXPECT().FindUserCredentials(gomock.Any(), userCredsReq).Return(model.UserCredentials{UserId: 1, Active: false}, nil)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Password doesnt match",
			req: (func() *http.Request {
				token, err := getFakeToken()
				require.NoError(t, err)

				registerReq := model.AuthenticateRequest{
					Email:    "test@gmail.com",
					Password: "12345qwerty",
				}

				return createFakeRequestWithBody(token, registerReq)
			})(),
			mockBehavior: func(ctx context.Context, mrepo *mock_repo.MockFbAuthRepo, mtx *mock_tx.MockTestTx, mlogger *mock_logger.MockFbLogger) {
				userCredsReq := model.UserCredentialsRequest{
					Email: "test@gmail.com",
				}
				mrepo.EXPECT().FindUserCredentials(gomock.Any(), userCredsReq).Return(model.UserCredentials{UserId: 1, Active: true, Salt: "123ww1"}, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			viper.Set("auth.jwt.cert", "../../../hack/dev/certs/server-cert.pem")
			viper.Set("auth.jwt.key", "../../../hack/dev/certs/server-key.pem")
			defer viper.Reset()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			mockRepo := mock_repo.NewMockFbAuthRepo(ctrl)
			mockLogger := mock_logger.NewMockFbLogger(ctrl)
			mockTx := mock_tx.NewMockTestTx(ctrl)

			handler := &services.ApiHandler{
				Logger: mockLogger,
			}

			service := &service{
				Repo:       mockRepo,
				ApiHandler: handler,
			}

			w := httptest.NewRecorder()

			tc.mockBehavior(ctx, mockRepo, mockTx, mockLogger)

			service.Login(w, tc.req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestLogout(t *testing.T) {
	viper.SetDefault("auth.cookie_name", "test_cookie_name")
	defer viper.Reset()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repo.NewMockFbAuthRepo(ctrl)
	mockLogger := mock_logger.NewMockFbLogger(ctrl)
	handler := &services.ApiHandler{
		Logger: mockLogger,
	}

	service := &service{
		Repo:       mockRepo,
		ApiHandler: handler,
	}

	req := httptest.NewRequest("GET", "/logout", nil)
	w := httptest.NewRecorder()

	service.Logout(w, req)

	cookies := w.Result().Cookies()
	assert.Equal(t, 1, len(cookies))

	expectedCookie := http.Cookie{
		Name:    "test_cookie_name",
		Value:   "",
		Expires: time.Now().Add(1 * time.Second),
		Path:    "/",
	}

	assert.Equal(t, expectedCookie.Name, cookies[0].Name)
	assert.Equal(t, expectedCookie.Value, cookies[0].Value)
	assert.Equal(t, expectedCookie.Expires.Unix(), cookies[0].Expires.Unix())
	assert.Equal(t, expectedCookie.Path, cookies[0].Path)
	assert.Equal(t, http.StatusOK, w.Code)
}

func createFakeRequestWithToken(token jwt.Token) *http.Request {
	req := httptest.NewRequest("GET", "/example", nil)

	ctx := context.WithValue(req.Context(), model.ContextJWTPointer, token)
	req = req.WithContext(ctx)

	return req
}

func createFakeRequestWithBody(token jwt.Token, body any) *http.Request {
	b, err := json.Marshal(body)
	if err != nil {
		log.Fatal("Error while marshal test")
	}

	bodyString := strings.NewReader(string(b))

	req := httptest.NewRequest("POST", "/example", bodyString)

	ctx := context.WithValue(req.Context(), model.ContextJWTPointer, token)
	req = req.WithContext(ctx)

	return req
}

func getFakeToken() (jwt.Token, error) {
	tokenId, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("Unable to generate token id: %s", err)
	}

	token, err := jwt.NewBuilder().
		JwtID(tokenId.String()).
		Issuer("fb-fightbettr").
		Audience([]string{"localhost"}).
		IssuedAt(time.Now()).
		Subject("test").
		Expiration(time.Now().Add(5 * time.Second)).
		Build()

	if err != nil {
		return nil, err
	}

	return token, nil
}
