package queue

import (
	"marte-desc/internal/ytdlp"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// fakeYtDlp writes a shell script that mimics yt-dlp output; body is its shell body.
func fakeYtDlp(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "yt-dlp")
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func newQueue(bin string, max int, done func(Job)) (*Queue, *sync.Map) {
	last := &sync.Map{}
	q := New(Deps{
		Env:       func() (ytdlp.Env, error) { return ytdlp.Env{Bin: bin}, nil },
		OnUpdate:  func(j Job) { last.Store(j.ID, j) },
		OnDone:    done,
		MaxActive: func() int { return max },
	})
	return q, last
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timeout waiting for condition")
}

func status(q *Queue, id string) string {
	for _, j := range q.List() {
		if j.ID == id {
			return j.Status
		}
	}
	return ""
}

func TestJobCompletes(t *testing.T) {
	bin := fakeYtDlp(t, `echo "MARTE|downloading|50|100|NA|10|1"
echo "MARTEFILE|/out/a.mp4"`)
	var doneJob Job
	q, _ := newQueue(bin, 1, func(j Job) { doneJob = j })
	ids := q.Enqueue([]Request{{URL: "u", Dir: "/out", Mode: "video", VideoFormat: "mp4"}})
	waitFor(t, func() bool { return status(q, ids[0]) == StatusDone })
	if doneJob.Path != "/out/a.mp4" || doneJob.Percent != 100 {
		t.Fatalf("bad done job: %+v", doneJob)
	}
}

func TestFailureReportsLastStderrLine(t *testing.T) {
	bin := fakeYtDlp(t, `echo "ERROR: Video unavailable" >&2; exit 1`)
	q, _ := newQueue(bin, 1, nil)
	ids := q.Enqueue([]Request{{URL: "u"}})
	waitFor(t, func() bool { return status(q, ids[0]) == StatusError })
	if got := q.List()[0].Error; got != "Video unavailable" {
		t.Fatalf("error = %q", got)
	}
}

func TestConcurrencyLimitAndCancelQueued(t *testing.T) {
	bin := fakeYtDlp(t, `sleep 30`)
	q, _ := newQueue(bin, 1, nil)
	ids := q.Enqueue([]Request{{URL: "a"}, {URL: "b"}})
	waitFor(t, func() bool { return status(q, ids[0]) == StatusDownloading })
	if s := status(q, ids[1]); s != StatusQueued {
		t.Fatalf("second job should wait, got %s", s)
	}
	q.Cancel(ids[1])
	if s := status(q, ids[1]); s != StatusCanceled {
		t.Fatalf("queued cancel: %s", s)
	}
	q.Cancel(ids[0]) // running: kills process
	waitFor(t, func() bool { return status(q, ids[0]) == StatusCanceled })
}

func TestRetryAfterError(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "ran")
	bin := fakeYtDlp(t, `if [ -f "`+marker+`" ]; then echo "MARTEFILE|/ok.mp4"; else touch "`+marker+`"; echo "ERROR: boom" >&2; exit 1; fi`)
	q, _ := newQueue(bin, 1, nil)
	ids := q.Enqueue([]Request{{URL: "u"}})
	waitFor(t, func() bool { return status(q, ids[0]) == StatusError })
	q.Retry(ids[0])
	waitFor(t, func() bool { return status(q, ids[0]) == StatusDone })
}
