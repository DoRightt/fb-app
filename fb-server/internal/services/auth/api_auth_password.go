package auth

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	internalErr "fightbettr.com/fb-server/pkg/errors"
	"fightbettr.com/fb-server/pkg/httplib"
	"fightbettr.com/fb-server/pkg/model"
	"fightbettr.com/fb-server/pkg/utils"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// ResetPassword handles the process of resetting a user's password.
// It expects a JSON request containing the user's email address.
// If the email is valid and associated with an existing user, a reset token is generated,
// and an email containing the reset link is sent to the user.
// The reset token is also stored in the database for verification during the password reset process.
// A successful response is returned if the email exists, and the reset process is initiated.
// In case of errors, appropriate error responses are sent with details in the response body.
func (s *service) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	decoder := json.NewDecoder(r.Body)
	var req model.ResetPasswordRequest
	if err := decoder.Decode(&req); err != nil {
		httplib.ErrorResponseJSON(w, http.StatusBadRequest, internalErr.AuthDecode, err)
		return
	}

	noEmail := len(req.Email) < 1 || req.Email == " "
	if noEmail {
		httplib.ErrorResponseJSON(w, http.StatusBadRequest, internalErr.AuthFormEmailEmpty,
			fmt.Errorf("%s", "Empty 'email'"))
		return
	}

	req.Email = strings.ToLower(req.Email)

	credentials, err := s.Repo.FindUserCredentials(ctx, model.UserCredentialsRequest{
		Email: req.Email,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			httplib.ErrorResponseJSON(w, http.StatusNotFound, http.StatusNotFound, err)
			return
		} else {
			s.Logger.Errorf("Failed to find user credentials: %s", err)
			httplib.ErrorResponseJSON(w, http.StatusInternalServerError, internalErr.UserCredentials, err)
			return
		}
	}

	user, err := s.Repo.FindUser(ctx, &model.UserRequest{
		UserId: credentials.UserId,
	})
	if err != nil {
		s.Logger.Errorf("Failed to find user: %s", err)
		httplib.ErrorResponseJSON(w, http.StatusInternalServerError, internalErr.Profile, err)
		return
	}

	tx, err := s.Repo.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		s.Logger.Errorf("Failed to create registration transaction: %s", err)
		httplib.ErrorResponseJSON(w, http.StatusBadRequest, internalErr.Tx, err)
		return
	}

	rn := rand.New(rand.NewSource(time.Now().UnixNano()))
	salt := rn.Int()

	token := utils.GenerateHashFromString(fmt.Sprintf("%s:%s:%d", req.Email, time.Now(), +salt))
	tokenExpire := time.Now().Unix() + 60*60*48
	credentials.TokenType = model.TokenResetPassword
	credentials.Token = token
	credentials.TokenExpire = tokenExpire

	if err := s.Repo.ResetPassword(ctx, &credentials); err != nil {
		s.Logger.Errorf("Failed to reset user credentials: %s", err)
		httplib.ErrorResponseJSON(w, http.StatusInternalServerError, internalErr.TxCommit, err)
		return
	}

	if txErr := tx.Commit(ctx); txErr != nil {
		s.Logger.Errorf("Failed to commit registration transaction: %s", txErr)
		httplib.ErrorResponseJSON(w, http.StatusBadRequest, 11, txErr)
		return
	}

	go s.HandleEmailEvent(ctx, &model.EmailData{
		Subject: model.EmailResetPassword,
		Recipient: model.EmailAddrSpec{
			Email: credentials.Email,
			Name:  user.Name,
		},
		Token: credentials.Token,
	})

	httplib.ResponseJSON(w, httplib.SuccessfulResult())
}

// RecoverPassword handles the process of recovering a user's password based on a provided reset token.
// It expects a JSON request containing the reset token, new password, and confirmation password.
// If the token is valid, the password is updated, and the token is marked as used.
// The response includes a successful result if the password recovery process is completed.
// In case of errors, appropriate error responses are sent with details in the response body.
func (s *service) RecoverPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	decoder := json.NewDecoder(r.Body)
	var req model.RecoverPasswordRequest
	if err := decoder.Decode(&req); err != nil {
		httplib.ErrorResponseJSON(w, http.StatusBadRequest, internalErr.AuthDecode, err)
		return
	}

	if len(req.Token) < 2 || req.Token == " " {
		httplib.ErrorResponseJSON(w, http.StatusBadRequest, internalErr.Token,
			fmt.Errorf("empty 'token'"))
		return
	}

	noPassword := len(req.Password) < 6
	noConfirm := len(req.ConfirmPassword) < 6

	if noPassword && noConfirm {
		httplib.ErrorResponseJSON(w, http.StatusBadRequest, internalErr.AuthFormPasswordInvalid,
			fmt.Errorf("empty body 'password'"))
		return
	}

	if noConfirm {
		httplib.ErrorResponseJSON(w, http.StatusBadRequest, internalErr.Auth,
			fmt.Errorf("empty body 'confirm_password'"))
		return
	}

	if req.Password != req.ConfirmPassword {
		httplib.ErrorResponseJSON(w, http.StatusBadRequest, internalErr.AuthFormPasswordsMismatch,
			fmt.Errorf("password are not equal"))
		return
	}

	credentials, err := s.Repo.FindUserCredentials(ctx, model.UserCredentialsRequest{
		Token: req.Token,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			httplib.ErrorResponseJSON(w, http.StatusNotFound, http.StatusNotFound, err)
			return
		} else {
			s.Logger.Errorf("Failed to find user credentials: %s", err)
			httplib.ErrorResponseJSON(w, http.StatusInternalServerError, internalErr.UserCredentials, err)
			return
		}
	}

	tx, err := s.Repo.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		s.Logger.Errorf("Failed to create registration transaction: %s", err)
		httplib.ErrorResponseJSON(w, http.StatusBadRequest, internalErr.Tx, err)
		return
	}

	salt := utils.GetRandomString(saltLength)
	password := utils.GenerateSaltedHash(req.Password, salt)

	credentials.Password = password
	credentials.Salt = salt

	if err := s.Repo.ConfirmCredentialsToken(ctx, tx, model.UserCredentialsRequest{
		UserId: credentials.UserId,
	}); err != nil {
		s.Logger.Errorf("Failed to reset user credentials: %s", err)
		httplib.ErrorResponseJSON(w, http.StatusInternalServerError, internalErr.UserCredentials, err)
		return
	}

	if err := s.Repo.UpdatePassword(ctx, tx, credentials); err != nil {
		s.Logger.Errorf("Failed to update user password: %s", err)
		httplib.ErrorResponseJSON(w, http.StatusInternalServerError, internalErr.UserCredentialsReset, err)
		return
	}

	if txErr := tx.Commit(ctx); txErr != nil {
		s.Logger.Errorf("Failed to commit registration transaction: %s", txErr)
		httplib.ErrorResponseJSON(w, http.StatusBadRequest, internalErr.TxCommit, txErr)
		return
	}

	httplib.ResponseJSON(w, httplib.SuccessfulResult())
}
