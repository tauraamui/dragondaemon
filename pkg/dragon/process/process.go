package process

import (
	"context"

	"github.com/tauraamui/dragondaemon/pkg/log"
)

type Event int

const SHUTDOWN_EVT Event = 0x50

type Process interface {
	Setup() Process
	Start()
	Stop()
	Wait()
}

type Settings struct {
	WaitForShutdownMsg string
	Process            func(context.Context) []chan interface{}
}

func New(settings Settings) Process {
	return &process{
		waitForShutdownMsg: settings.WaitForShutdownMsg,
		process:            settings.Process,
	}
}

type process struct {
	process            func(context.Context) []chan interface{}
	waitForShutdownMsg string
	canceller          context.CancelFunc
	signals            []chan interface{}
}

func (p *process) logShutdown() {
	if len(p.waitForShutdownMsg) > 0 {
		log.Info(p.waitForShutdownMsg)
	}
}

func (p *process) Setup() Process { return p }

func (p *process) Start() {
	ctx, canceller := context.WithCancel(context.Background())
	p.canceller = canceller
	p.signals = append(p.signals, p.process(ctx)...)
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
