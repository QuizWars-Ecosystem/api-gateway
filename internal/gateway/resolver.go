package gateway

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

const customScheme = "dynamic"

var (
	_ resolver.Builder  = (*Builder)(nil)
	_ resolver.Resolver = (*Resolver)(nil)
)

func NewBuilder(output <-chan []*api.ServiceEntry, logger *zap.Logger) *Builder {
	return &Builder{
		output: output,
		logger: logger,
	}
}

type Builder struct {
	output <-chan []*api.ServiceEntry
	logger *zap.Logger
}

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := NewResolver(target, cc, opts, b.output, b.logger)

	go r.watch()

	return r, nil
}

func (b *Builder) Scheme() string {
	return customScheme
}

type Resolver struct {
	target    resolver.Target
	cc        resolver.ClientConn
	opts      resolver.BuildOptions
	addresses []resolver.Address
	input     <-chan []*api.ServiceEntry
	logger    *zap.Logger
}

func NewResolver(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions, input <-chan []*api.ServiceEntry, logger *zap.Logger) *Resolver {
	return &Resolver{
		target: target,
		cc:     cc,
		opts:   opts,
		input:  input,
		logger: logger,
	}
}

func (r *Resolver) ResolveNow(_ resolver.ResolveNowOptions) {}

func (r *Resolver) Close() {}

func (r *Resolver) watch() {
	for serviceEntries := range r.input {
		r.update(serviceEntries)
	}
}

func (r *Resolver) update(entries []*api.ServiceEntry) {
	addrs := make([]resolver.Address, 0, len(entries))

	r.logger.Info("updating resolver address state", zap.String("target", r.target.String()), zap.Int("amount", len(entries)))

	for _, entry := range entries {
		addrs = append(addrs, resolver.Address{
			ServerName: entry.Service.Service,
			Addr:       fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port),
		})
	}

	r.addresses = addrs

	if err := r.cc.UpdateState(resolver.State{Addresses: addrs}); err != nil {
		r.logger.Error("error updating resolver state", zap.String("target", r.target.String()), zap.Error(err))
	}
}
