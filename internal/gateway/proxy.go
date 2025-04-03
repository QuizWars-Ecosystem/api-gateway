package gateway

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/siderolabs/grpc-proxy/proxy"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	apperrors "github.com/QuizWars-Ecosystem/go-common/pkg/error"
)

var _ proxy.Backend = (*Proxy)(nil)

type Proxy struct {
	grpcConns map[string]*grpc.ClientConn
	logger    *zap.Logger
}

func NewProxy(grcConns map[string]*grpc.ClientConn, logger *zap.Logger) *Proxy {
	connsMap := make(map[string]*grpc.ClientConn, len(grcConns))

	for service, conn := range grcConns {
		connsMap[fmt.Sprintf("/%s", service)] = conn
	}

	return &Proxy{
		grpcConns: connsMap,
		logger:    logger,
	}
}

func (p *Proxy) String() string {
	return "proxy"
}

func (p *Proxy) GetConnection(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
	name := strings.Split(fullMethodName, ".")[0]

	if conn, ok := p.grpcConns[name]; ok {
		return ctx, conn, nil
	}

	return nil, nil, apperrors.Internal(errors.New("connection not found"))
}

func (p *Proxy) AppendInfo(_ bool, resp []byte) ([]byte, error) {
	return resp, nil
}

func (p *Proxy) BuildError(_ bool, err error) ([]byte, error) {
	return []byte(err.Error()), nil
}

func (p *Proxy) Director(_ context.Context, _ string) (proxy.Mode, []proxy.Backend, error) {
	return proxy.One2One, []proxy.Backend{p}, nil
}
