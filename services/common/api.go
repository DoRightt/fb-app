package common

import (
	"context"
	"net/http"
	authRepo "projects/fb-server/repo/auth"
	commonRepo "projects/fb-server/repo/common"
	"projects/fb-server/services"
)

type service struct {
	*services.ApiHandler
	Repo     *commonRepo.CommonRepo `json:"-" yaml:"-"`
	AuthRepo *authRepo.AuthRepo     `json:"-" yaml:"-"`
}

func New(h *services.ApiHandler) services.ApiService {
	return &service{
		ApiHandler: h,
		Repo:       commonRepo.New(h.Repo),
		AuthRepo:   authRepo.New(h.Repo),
	}
}

func (s *service) Shutdown(ctx context.Context, sig string) {}

func (s *service) Init(ctx context.Context) error {
	return nil
}

func (s *service) ApplyRoutes() {
	s.Router.HandleFunc("/fighters", s.SearchFighters).Methods(http.MethodGet)
}