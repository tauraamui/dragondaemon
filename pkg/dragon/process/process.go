package process

import (
	"context"

	"github.com/tauraamui/dragondaemon/pkg/log"
)

type Event int

const SHUTDOWN_EVT Event = 0x50

type Process interface {
	Setup() Process
	Start() <-chan interface{}
	Stop()
	Wait()
}

type Settings struct {
	WaitForShutdownMsg string
	Process            func(context.Context, chan interface{}) []chan interface{}
}

func New(settings Settings) Process {
	return &process{
		started:            make(chan interface{}),
		waitForShutdownMsg: settings.WaitForShutdownMsg,
		process:            settings.Process,
	}
}

type process struct {
	started            chan interface{}
	process            func(context.Context, chan interface{}) []chan interface{}
	waitForShutdownMsg string
	canceller          context.CancelFunc
	signals            []chan interface{}
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
		p.started = make(chan interface{})
	}
}

func (p *process) Start() <-chan interface{} {
	p.initStarted()
	go func(s chan interface{}) {
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
