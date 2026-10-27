package queue

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"marte-media/internal/ytdlp"
)

const (
	StatusQueued      = "queued"
	StatusDownloading = "downloading"
	StatusProcessing  = "processing"
	StatusDone        = "done"
	StatusError       = "error"
	StatusCanceled    = "canceled"
)

// Request describes one download. Each playlist entry is its own Request.
type Request struct {
	URL          string `json:"url"`
	Title        string `json:"title"`
	Channel      string `json:"channel"`
	Thumbnail    string `json:"thumbnail"`
	Mode         string `json:"mode"`
	Height       int    `json:"height"`
	VideoFormat  string `json:"videoFormat"`
	AudioFormat  string `json:"audioFormat"`
	AudioBitrate int    `json:"audioBitrate"`
	Dir          string `json:"dir"`
}

type Job struct {
	ID         string  `json:"id"`
	Request    Request `json:"request"`
	Status     string  `json:"status"`
	Percent    float64 `json:"percent"`
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	Speed      float64 `json:"speed"`
	ETA        int     `json:"eta"`
	Path       string  `json:"path"`
	Error      string  `json:"error"`
}

// Deps are the things the queue needs from the outside world.
type Deps struct {
	Env       func() (ytdlp.Env, error)
	OnUpdate  func(Job) // called on every state change (already throttled for progress)
	OnDone    func(Job) // called once when a job finishes successfully
	MaxActive func() int
}

type Queue struct {
	deps Deps

	mu       sync.Mutex
	jobs     []*Job
	cancels  map[string]context.CancelFunc
	nextID   int
	active   int
	lastEmit map[string]time.Time
}

func New(deps Deps) *Queue {
	return &Queue{
		deps:     deps,
		cancels:  map[string]context.CancelFunc{},
		lastEmit: map[string]time.Time{},
	}
}

// Enqueue adds jobs and starts as many as the concurrency limit allows.
func (q *Queue) Enqueue(reqs []Request) []string {
	q.mu.Lock()
	ids := make([]string, 0, len(reqs))
	var added []Job
	for _, r := range reqs {
		q.nextID++
		j := &Job{ID: fmt.Sprintf("j%d-%d", time.Now().UnixNano(), q.nextID), Request: r, Status: StatusQueued, ETA: -1}
		q.jobs = append(q.jobs, j)
		ids = append(ids, j.ID)
		added = append(added, *j)
	}
	q.mu.Unlock()
	for _, j := range added {
		q.emit(j)
	}
	q.pump()
	return ids
}

func (q *Queue) List() []Job {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]Job, len(q.jobs))
	for i, j := range q.jobs {
		out[i] = *j
	}
	return out
}

// Cancel stops a running job or removes a queued one from the run order.
func (q *Queue) Cancel(id string) {
	q.mu.Lock()
	j := q.find(id)
	if j == nil {
		q.mu.Unlock()
		return
	}
	if cancel, ok := q.cancels[id]; ok {
		q.mu.Unlock()
		cancel() // run() records the canceled state
		return
	}
	if j.Status != StatusQueued {
		q.mu.Unlock()
		return
	}
	j.Status = StatusCanceled
	snap := *j
	q.mu.Unlock()
	q.emit(snap)
}

// Retry re-queues a failed or canceled job.
func (q *Queue) Retry(id string) {
	q.mu.Lock()
	j := q.find(id)
	if j == nil || (j.Status != StatusError && j.Status != StatusCanceled) {
		q.mu.Unlock()
		return
	}
	*j = Job{ID: j.ID, Request: j.Request, Status: StatusQueued, ETA: -1}
	snap := *j
	q.mu.Unlock()
	q.emit(snap)
	q.pump()
}

// ClearFinished drops done, failed and canceled jobs from the list.
func (q *Queue) ClearFinished() {
	q.mu.Lock()
	kept := q.jobs[:0]
	for _, j := range q.jobs {
		switch j.Status {
		case StatusDone, StatusError, StatusCanceled:
			delete(q.lastEmit, j.ID)
		default:
			kept = append(kept, j)
		}
	}
	q.jobs = kept
	q.mu.Unlock()
}

// pump starts queued jobs while there is capacity.
func (q *Queue) pump() {
	for {
		q.mu.Lock()
		limit := 1
		if q.deps.MaxActive != nil {
			limit = max(1, q.deps.MaxActive())
		}
		if q.active >= limit {
			q.mu.Unlock()
			return
		}
		var next *Job
		for _, j := range q.jobs {
			if j.Status == StatusQueued {
				next = j
				break
			}
		}
		if next == nil {
			q.mu.Unlock()
			return
		}
		next.Status = StatusDownloading
		q.active++
		ctx, cancel := context.WithCancel(context.Background())
		q.cancels[next.ID] = cancel
		snap := *next
		q.mu.Unlock()
		q.emit(snap)
		go q.run(ctx, next.ID, snap.Request)
	}
}

func (q *Queue) run(ctx context.Context, id string, r Request) {
	env, err := q.deps.Env()
	var path string
	if err == nil {
		path, err = ytdlp.Download(ctx, env, ytdlp.Options{
			URL: r.URL, Dir: r.Dir, Mode: r.Mode, Height: r.Height,
			VideoFormat: r.VideoFormat, AudioFormat: r.AudioFormat, AudioBitrate: r.AudioBitrate,
		}, func(p ytdlp.Progress) { q.progress(id, p) })
	}

	q.mu.Lock()
	j := q.find(id)
	switch {
	case errors.Is(err, context.Canceled):
		j.Status = StatusCanceled
	case err != nil:
		j.Status, j.Error = StatusError, err.Error()
	default:
		j.Status, j.Percent, j.Path, j.ETA = StatusDone, 100, path, 0
	}
	snap := *j
	q.cancels[id]()
	delete(q.cancels, id)
	delete(q.lastEmit, id)
	q.active--
	q.mu.Unlock()

	q.emit(snap)
	if snap.Status == StatusDone && q.deps.OnDone != nil {
		q.deps.OnDone(snap)
	}
	q.pump()
}

// progress updates a job and emits at most every 200ms (phase changes always emit).
func (q *Queue) progress(id string, p ytdlp.Progress) {
	q.mu.Lock()
	j := q.find(id)
	if j == nil {
		q.mu.Unlock()
		return
	}
	prev := j.Status
	if p.Phase == "processing" {
		j.Status = StatusProcessing
	}
	j.Percent, j.Downloaded, j.Total, j.Speed, j.ETA = p.Percent, p.Downloaded, p.Total, p.Speed, p.ETA
	if p.Phase == "processing" {
		j.Percent = 100
	}
	now := time.Now()
	if j.Status == prev && now.Sub(q.lastEmit[id]) < 200*time.Millisecond {
		q.mu.Unlock()
		return
	}
	q.lastEmit[id] = now
	snap := *j
	q.mu.Unlock()
	q.emit(snap)
}

func (q *Queue) emit(j Job) {
	if q.deps.OnUpdate != nil {
		q.deps.OnUpdate(j)
	}
}

func (q *Queue) find(id string) *Job {
	for _, j := range q.jobs {
		if j.ID == id {
			return j
		}
	}
	return nil
}
