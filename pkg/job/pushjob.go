package job

import (
	"context"
	"time"

	"github.com/infrawatch/ci-server-go/pkg/ghclient"
	"github.com/infrawatch/ci-server-go/pkg/logging"
)

// PushJob contains logic for dealing with github push events
type PushJob struct {
	event             *ghclient.Push
	client            *ghclient.Client
	scriptOutput      []byte
	afterScriptOutput []byte

	cancel context.CancelFunc

	BasePath string
	Log      *logging.Logger
}

// SetLogger impelements type job.Job
func (p *PushJob) SetLogger(l *logging.Logger) {
	p.Log = l
}

// Cancel used to manually cancel job if context does no cancel first
func (p *PushJob) Cancel() {
	p.cancel()
}

// GetRefName get reference name from event that triggered job
func (p *PushJob) GetRefName() string {
	return p.event.RefName
}

// Run ...
func (p *PushJob) Run(ctx context.Context) {
	commit := p.event.Ref.GetHead()
	cj := newCoreJob(p.client, p.event.Repo, *commit, p.Log)
	cj.BasePath = "/tmp/"

	cj.getResources()

	ctxTimeoutScript, cancelScript := context.WithTimeout(ctx, time.Second*300)
	p.cancel = cancelScript
	defer cancelScript()
	cj.runScript(ctxTimeoutScript)

	// It is highly NOT recommended to create top level contexts in lower functions
	// 'After script' is responsible for cleaning up resources, so it must run even when a cancel signal
	// has been sent by the main server goroutine. This still garauntees an exit after timeout
	// so it isn't too terrible
	ctxTimeoutAfterScrip, cancelAfterScript := context.WithTimeout(context.Background(), time.Second)
	defer cancelAfterScript()
	cj.runAfterScript(ctxTimeoutAfterScrip)
	cj.postResults()
}