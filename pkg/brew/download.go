package brew

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "log"
    "net/url"
    "os"
    "os/signal"
    "path/filepath"
    "sync"

    getter "github.com/hashicorp/go-getter"
)

func downloadFile(src url.URL, dest string, showProgress bool) error {
    ctx, cancel := context.WithCancel(context.Background())

    pwd, err := os.Getwd()
    if err != nil {
        return err
    }

    opts := []getter.ClientOption{}
    if showProgress {
        opts = append(opts, getter.WithProgress(defaultProgressBar))
    }

    // Build the client
    client := &getter.Client{
        Ctx:     ctx,
        Src:     src.String(),
        Dst:     dest,
        Pwd:     pwd,
        Mode:    getter.ClientModeFile,
        Options: opts,
    }

    wg := sync.WaitGroup{}
    wg.Add(1)
    errChan := make(chan error, 2)
    go func() {
        defer wg.Done()
        defer cancel()
        if err := client.Get(); err != nil {
            errChan <- err
        }
    }()

    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt)

    select {
    case <-c:
        signal.Reset(os.Interrupt)
        cancel()
        wg.Wait()
    case <-ctx.Done():
        wg.Wait()
    case err := <-errChan:
        wg.Wait()
        return err
    }

    return err
}

func FetchResourceWithChecksum(rawsrc string, checksum string, showProgress bool) string {
    src, _ := url.Parse(rawsrc)
    dest := filepath.Join(CurrentPlatform().HomebrewCache(), filepath.Base(src.Path))
    q, err := url.ParseQuery(src.RawQuery)
    if err != nil {
        panic(err)
    }
    q.Add("archive", "false")
    q.Add("checksum", "sha256:" + checksum)
    src.RawQuery = q.Encode()
    fmt.Printf("fetch\t%s\n", rawsrc)
    downloadFile(*src, dest, showProgress)

    fmt.Printf("verify\t%s -> %s\n", rawsrc, checksum)
    hasher := sha256.New()
    f, err := os.Open(dest)
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    if _, err := io.Copy(hasher, f); err != nil {
        log.Fatal(err)
    }

    actual := hex.EncodeToString(hasher.Sum(nil))
    if actual != checksum {
        log.Fatalf("Unexpected checksum: Expected %s but got %s", checksum, actual)
    }
    return dest
}

func FetchResource(rawsrc string, showProgress bool) string {
    src, _ := url.Parse(rawsrc)
    dest := filepath.Join(CurrentPlatform().HomebrewCache(), filepath.Base(src.Path))
    q, err := url.ParseQuery(src.RawQuery)
    if err != nil {
        panic(err)
    }
    q.Add("archive", "false")
    src.RawQuery = q.Encode()
    fmt.Printf("fetch\t%s\n", rawsrc)
    downloadFile(*src, dest, showProgress)
    return dest
}
