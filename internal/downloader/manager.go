package downloader

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var rangeSupportCache sync.Map
var probeClient = &http.Client{Timeout: 5 * time.Second}

var ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
	"AppleWebKit/537.36 (KHTML, like Gecko) " +
	"Chrome/120.0.0.0 Safari/537.36"

func Download(ctx context.Context, rawurl, outPath string, verbose bool, md5sum, sha256sum string, progressCh chan<- tea.Msg, id int) error {
	// Check if server supports range requests
	supportsRange, _ := checkRangeSupport(ctx, rawurl)

	if supportsRange {
		// Use concurrent downloader
		d := NewConcurrentDownloader()
		d.SetProgressChan(progressCh)
		d.SetID(id)
		return d.Download(ctx, rawurl, outPath, verbose, md5sum, sha256sum)
	}

	// Use single-threaded downloader
	d := NewSingleDownloader()
	d.SetProgressChan(progressCh)
	d.SetID(id)
	return d.Download(ctx, rawurl, outPath, verbose)
}

func TUIDownload(ctx context.Context, rawurl, outPath string, verbose bool, md5sum, sha256sum string, progressCh chan<- tea.Msg, id int, state *ProgressState) error {
	// Check if server supports range requests
	supportsRange, _ := checkRangeSupport(ctx, rawurl)

	if supportsRange {
		// Use concurrent downloader
		d := NewConcurrentDownloader()
		d.SetProgressChan(progressCh)
		d.SetID(id)
		if state != nil {
			d.SetProgressState(state)
		}
		return d.Download(ctx, rawurl, outPath, verbose, md5sum, sha256sum)
	}

	// Use single-threaded downloader
	d := NewSingleDownloader()
	d.SetProgressChan(progressCh)
	d.SetID(id)
	if state != nil {
		d.SetProgressState(state)
	}
	return d.Download(ctx, rawurl, outPath, verbose)
}

func checkRangeSupport(ctx context.Context, rawurl string) (bool, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return false, err
	}
	hostKey := u.Host
	if v, ok := rangeSupportCache.Load(hostKey); ok {
		return v.(bool), nil
	}

	headCtx, headCancel := context.WithTimeout(ctx, 3*time.Second)
	defer headCancel()

	headReq, err := http.NewRequestWithContext(headCtx, http.MethodHead, rawurl, nil)
	if err == nil {
		headReq.Header.Set("User-Agent", ua)
		if resp, err := probeClient.Do(headReq); err == nil {
			resp.Body.Close()
			if strings.Contains(strings.ToLower(resp.Header.Get("Accept-Ranges")), "bytes") {
				rangeSupportCache.Store(hostKey, true)
				return true, nil
			}
			if resp.Header.Get("Content-Range") != "" {
				rangeSupportCache.Store(hostKey, true)
				return true, nil
			}
		}
	}

	getCtx, getCancel := context.WithTimeout(ctx, 5*time.Second)
	defer getCancel()

	getReq, err := http.NewRequestWithContext(getCtx, http.MethodGet, rawurl, nil)
	if err != nil {
		return false, err
	}
	getReq.Header.Set("Range", "bytes=0-0")
	getReq.Header.Set("User-Agent", ua)

	resp, err := probeClient.Do(getReq)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusPartialContent {
		rangeSupportCache.Store(hostKey, true)
		return true, nil
	}
	if strings.Contains(strings.ToLower(resp.Header.Get("Accept-Ranges")), "bytes") || resp.Header.Get("Content-Range") != "" {
		rangeSupportCache.Store(hostKey, true)
		return true, nil
	}

	rangeSupportCache.Store(hostKey, false)
	return false, nil
}
