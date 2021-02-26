package brew

import (
    "io"
    "sync"

    "github.com/cheggaaa/pb/v3"
    getter "github.com/hashicorp/go-getter"
)

// defaultProgressBar is the default instance of a cheggaaa
// progress bar.
var defaultProgressBar getter.ProgressTracker = &ProgressBar{}

// ProgressBar wraps a github.com/cheggaaa/pb.Pool
// in order to display download progress for one or multiple
// downloads.
//
// If two different instance of ProgressBar try to
// display a progress only one will be displayed.
// It is therefore recommended to use DefaultProgressBar
type ProgressBar struct {
    // lock everything below
    lock sync.Mutex

    bar *pb.ProgressBar
}

// TrackProgress instantiates a new progress bar that will
// display the progress of stream until closed.
// total can be 0.
func (cpb *ProgressBar) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
    cpb.lock.Lock()
    defer cpb.lock.Unlock()

    if cpb.bar == nil {
        cpb.bar = pb.New64(totalSize)
        cpb.bar.SetTemplate(pb.Simple)
        cpb.bar.Start()
    }
    cpb.bar.SetCurrent(currentSize)
    reader := cpb.bar.NewProxyReader(stream)

    return &readCloser{
        Reader: reader,
        close: func() error {
            cpb.lock.Lock()
            defer cpb.lock.Unlock()

            cpb.bar.Finish()
            return nil
        },
    }
}

type readCloser struct {
    io.Reader
    close func() error
}

func (c *readCloser) Close() error { return c.close() }
