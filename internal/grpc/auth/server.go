package auth

import (
	"context"
	tradeService "github.com/mjolner36/trade_item_service/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context, email string, password string, appID int) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type serverAPI struct {
	tradeService.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	tradeService.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(
	ctx context.Context,
	req *tradeService.LoginRequest,
) (*tradeService.LoginResponse, error) {
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		//TODO:
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &tradeService.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *tradeService.RegisterRequest,
) (*tradeService.RegisterResponse, error) {
	//TODO:валидация на все handler

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		//TODO:
	}

	return &tradeService.RegisterResponse{UserId: userID}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *tradeService.IsAdminRequest,
) (*tradeService.IsAdminResponse, error) {
	panic("implement me")
}
