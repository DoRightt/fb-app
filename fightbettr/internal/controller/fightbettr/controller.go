package fightbettr

import (
	"context"

	authmodel "fightbettr.com/auth/pkg/model"
	fightersmodel "fightbettr.com/fighters/pkg/model"
)

type fightersGateway interface {
	SearchFighters(ctx context.Context, status fightersmodel.FighterStatus) ([]*fightersmodel.Fighter, error)
}

type authGateway interface {
	Register(ctx context.Context, req *authmodel.RegisterRequest) (*authmodel.UserCredentials, error)
	ConfirmRegistration(ctx context.Context, token string) (bool, error)
	Login(ctx context.Context, req *authmodel.AuthenticateRequest) (*authmodel.AuthenticateResult, error)
	ResetPassword(ctx context.Context, req *authmodel.ResetPasswordRequest) (bool, error)
	PasswordRecover(ctx context.Context, req *authmodel.RecoverPasswordRequest) (bool, error)
	GetCurrentUser(ctx context.Context) (*authmodel.User, error)
}

// Controller defines a gateway service controller.
type Controller struct {
	authGateway     authGateway
	fightersGateway fightersGateway
}

// New creates new Controller instance
func New(authGateway authGateway, fightersGateway fightersGateway) *Controller {
	return &Controller{
		authGateway,
		fightersGateway,
	}
}

// SearchFighters searches for fighters with the given status using the fightersGateway.
func (c *Controller) SearchFighters(ctx context.Context, status string) ([]*fightersmodel.Fighter, error) {
	fighters, err := c.fightersGateway.SearchFighters(ctx, fightersmodel.FighterStatus(status))
	if err != nil {
		return nil, err
	}

	return fighters, nil
}

// Register handles the registration of a new user. It takes a context and a 
// RegisterRequest, and returns the registered UserCredentials or an error.
func (c *Controller) Register(ctx context.Context, req *authmodel.RegisterRequest) (*authmodel.UserCredentials, error) {
	credentials, err := c.authGateway.Register(ctx, req)
	if err != nil {
		return &authmodel.UserCredentials{}, err
	}

	return credentials, nil
}

// ConfirmRegistration confirms a user's registration with the provided token.
func (c *Controller) ConfirmRegistration(ctx context.Context, token string) (bool, error) {
	ok, err := c.authGateway.ConfirmRegistration(ctx, token)
	if err != nil {
		return false, err
	}

	return ok, nil
}

// // Login authenticates a user with the provided credentials.
func (c *Controller) Login(ctx context.Context, req *authmodel.AuthenticateRequest) (*authmodel.AuthenticateResult, error) {
	token, err := c.authGateway.Login(ctx, req)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// ResetPassword resets a user's password with the provided request details.
func (c *Controller) ResetPassword(ctx context.Context, req *authmodel.ResetPasswordRequest) (bool, error) {
	ok, err := c.authGateway.ResetPassword(ctx, req)
	if err != nil {
		return false, err
	}

	return ok, nil
}

// PasswordRecover initiates the password recovery process for a user with the provided request details.
func (c *Controller) PasswordRecover(ctx context.Context, req *authmodel.RecoverPasswordRequest) (bool, error) {
	ok, err := c.authGateway.PasswordRecover(ctx, req)
	if err != nil {
		return false, err
	}

	return ok, nil
}

// GetCurrentUser retrieves the currently authenticated user.
func (c *Controller) GetCurrentUser(ctx context.Context) (*authmodel.User, error) {
	user, err := c.authGateway.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}
