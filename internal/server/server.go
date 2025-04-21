package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/QuizWars-Ecosystem/api-gateway/internal/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"

	"github.com/DavidMovas/gopherbox/pkg/closer"
	questionsv1 "github.com/QuizWars-Ecosystem/api-gateway/gen/external/questions/v1"
	usersv1 "github.com/QuizWars-Ecosystem/api-gateway/gen/external/users/v1"
	"github.com/QuizWars-Ecosystem/api-gateway/internal/config"
	"github.com/QuizWars-Ecosystem/api-gateway/internal/gateway"
	"github.com/QuizWars-Ecosystem/go-common/pkg/abstractions"
	"github.com/QuizWars-Ecosystem/go-common/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var _ abstractions.Server = (*Server)(nil)

type Server struct {
	gateway *gateway.Gateway
	logger  *log.Logger
	cfg     *config.Config
	closer  *closer.Closer
}

func NewServer(_ context.Context, cfg *config.Config) (*Server, error) {
	cl := closer.NewCloser()

	logger := log.NewLogger(cfg.Local, cfg.LogLevel)
	cl.PushIO(logger)

	srvOpts := []*gateway.ServiceOption{
		{
			Address: "users-service",
			RegisterFunc: []func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error{
				usersv1.RegisterUsersAuthServiceHandler,
				usersv1.RegisterUsersAdminServiceHandler,
				usersv1.RegisterUsersSocialServiceHandler,
				usersv1.RegisterUsersProfileServiceHandler,
			},
		},
		{
			Address: "questions-service",
			RegisterFunc: []func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error{
				questionsv1.RegisterQuestionsServiceHandler,
				questionsv1.RegisterQuestionsAdminServiceHandler,
				questionsv1.RegisterQuestionsClientServiceHandler,
			},
		},
	}

	gt, err := gateway.NewGateway(cfg.ConsulURL, srvOpts, logger)
	if err != nil {
		logger.Zap().Error("error initializing gateway", zap.Error(err))
		return nil, err
	}

	err = mux.RegisterRuntimeMux(gt.Runtime())
	if err != nil {
		logger.Zap().Error("error registering runtime mux", zap.Error(err))
		return nil, err
	}

	if cfg.Local {
		reflection.Register(gt.Proxy())
	}

	return &Server{
		gateway: gt,
		logger:  logger,
		cfg:     cfg,
		closer:  cl,
	}, nil
}

func (s *Server) Start() error {
	httpPort := s.cfg.HTTPPort
	grpcPort := s.cfg.GRPCPort
	logger := s.logger.Zap()

	if err := s.gateway.Start(); err != nil {
		logger.Error("error starting gateway", zap.Error(err))
		return err
	}

	group := errgroup.Group{}

	group.Go(func() error {
		logger.Info("starting http runtime server", zap.String("port", httpPort))

		if ls, err := net.Listen("tcp", fmt.Sprintf(":%s", httpPort)); err == nil {
			mux := s.gateway.Runtime()
			tcpSrv := &http.Server{Handler: mux}

			s.closer.PushIO(ls)
			s.closer.PushIO(tcpSrv)

			if err = tcpSrv.Serve(ls); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("error serving http mux server", zap.Error(err))
				return err
			}

			logger.Info("http mux server stopped")

			return nil
		} else {
			logger.Error("error starting http runtime server", zap.Error(err))
			return err
		}
	})

	group.Go(func() error {
		logger.Info("starting grpc proxy server", zap.String("port", grpcPort))

		if ls, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort)); err == nil {
			proxy := s.gateway.Proxy()

			s.closer.PushIO(ls)
			s.closer.PushNE(proxy.GracefulStop)

			if err = proxy.Serve(ls); err != nil {
				logger.Error("error serving grpc proxy server", zap.Error(err))
				return err
			}

			logger.Info("grpc proxy server stopped")

			return nil
		} else {
			logger.Error("error starting grpc proxy server", zap.Error(err))
			return err
		}
	})

	s.closer.Push(s.gateway.Stop)

	return group.Wait()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Zap().Info("Shutting down server...")
	return s.closer.Close(ctx)
}
