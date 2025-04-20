package gateway

import (
	"github.com/QuizWars-Ecosystem/go-common/pkg/log"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

const planType = "service"

type Plan struct {
	client  *api.Client
	logger  *log.Logger
	service string
	plan    *watch.Plan
	input   chan<- []*api.ServiceEntry
	errCh   chan<- error
}

func NewPlan(client *api.Client, logger *log.Logger, serviceName string, input chan<- []*api.ServiceEntry) *Plan {
	p := &Plan{}

	pl, _ := watch.Parse(map[string]interface{}{
		"type":        planType,
		"service":     serviceName,
		"passingonly": true,
	})

	p.client = client
	p.logger = logger
	p.service = serviceName
	p.plan = pl
	p.input = input

	pl.Handler = p.handle

	return p
}

func (p *Plan) handle(_ uint64, data interface{}) {
	if !p.plan.IsStopped() {
		entries := data.([]*api.ServiceEntry)
		if len(entries) > 0 {
			p.input <- entries
		}
	}
}

func (p *Plan) Run(errCh chan<- error) {
	go func() {
		if err := p.plan.RunWithClientAndHclog(p.client, p.logger.HCLogger()); err != nil {
			errCh <- err
		}
	}()

	p.errCh = errCh
}

func (p *Plan) Stop() {
	p.plan.Stop()

	if p.plan.IsStopped() {
		close(p.input)
		close(p.errCh)
	}
}
