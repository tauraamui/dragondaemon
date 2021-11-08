package process

import (
	"context"

	"github.com/tauraamui/dragondaemon/pkg/log"
)

type Event int

const SHUTDOWN_EVT Event = 0x50

type Process interface {
	Setup() Process
	Start() <-chan struct{}
	Stop()
	Wait()
}

type Settings struct {
	WaitForShutdownMsg string
	Process            func(context.Context, chan struct{}) []chan struct{}
}

func New(settings Settings) Process {
	return &process{
		started:            make(chan struct{}),
		waitForShutdownMsg: settings.WaitForShutdownMsg,
		process:            settings.Process,
	}
}

type process struct {
	started            chan struct{}
	process            func(context.Context, chan struct{}) []chan struct{}
	waitForShutdownMsg string
	canceller          context.CancelFunc
	signals            []chan struct{}
}

func (p *process) logShutdown() {
	if len(p.waitForShutdownMsg) > 0 {
		log.Info(p.waitForShutdownMsg)
	}
}

func (p *process) Setup() Process {
	p.initStarted()
	return p
}

func (p *process) initStarted() {
	if p.started == nil {
		p.started = make(chan struct{})
	}
}

func (p *process) Start() <-chan struct{} {
	p.initStarted()
	go func(s chan struct{}) {
		ctx, canceller := context.WithCancel(context.Background())
		p.canceller = canceller
		p.signals = append(p.signals, p.process(ctx, s)...)
	}(p.started)
	return p.started
}

func (p *process) Stop() {
	p.logShutdown()
	if p.canceller != nil {
		p.canceller()
	}
}

func (p *process) Wait() {
	for _, sig := range p.signals {
		<-sig
	}
}
