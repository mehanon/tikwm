package tt

import (
	"encoding/json"
	"fmt"
	"github.com/cavaliergopher/grab/v3"
	"net/http"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"
)

type DownloadOpt struct {
	Directory      string
	DownloadWith   func(url string, filename string) error
	ValidateWith   func(filename string) (bool, error)
	FilenameFormat func(post *Post, i int) string
	Timeout        time.Duration
	TimeoutOnError time.Duration
	NoSync         bool
	Retries        int
}

func ValidateWithFfprobe(ffprobe ...string) func(filename string) (bool, error) {
	ffprobe_ := "ffprobe"
	if len(ffprobe) != 0 {
		ffprobe_ = ffprobe[0]
	}

	return func(filename string) (bool, error) {
		out, err := exec.Command(ffprobe_, "-loglevel", "error", "-of", "json", "-show_entries", "stream_tags:format_tags", filename).Output()
		if err != nil {
			return false, fmt.Errorf("err: %s,\n%s", err.Error(), string(out))
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal(out, &metadata); err != nil || len(metadata) == 0 {
			return false, err
		}

		return true, nil
	}
}

func (opt *DownloadOpt) downloadRetrying(post *Post, url string, filename string, try int, lastErr error) error {
	if try > opt.Retries {
		return lastErr
	}

	ret := func(err error) error {
		time.Sleep(opt.TimeoutOnError)
		retry, retryErr := GetPost(post.ID())
		if retryErr != nil {
			return opt.downloadRetrying(retry, url, filename, try+1, retryErr)
		}
		return opt.downloadRetrying(retry, url, filename, try+1, err)
	}

	if err := opt.DownloadWith(url, filename); err != nil {
		return ret(err)
	}

	if valid, err := opt.ValidateWith(filename); err != nil {
		return ret(err)
	} else if !valid {
		time.Sleep(opt.TimeoutOnError)
		return ret(err)
	}

	return nil
}

func (opt *DownloadOpt) Defaults() *DownloadOpt {
	ret := opt
	if ret == nil {
		ret = &DownloadOpt{}
	}
	if ret.DownloadWith == nil {
		ret.DownloadWith = func(url string, filename string) error {
			req, err := grab.NewRequest(filename, url)
			if err != nil {
				return err
			}

			if resp := DefaultDownloadClient.Do(req); resp.Err() != nil {
				return err
			}
			return nil
		}
	}
	if ret.ValidateWith == nil {
		ret.ValidateWith = func(filename string) (bool, error) {
			return true, nil
		}
	}
	if ret.FilenameFormat == nil {
		ret.FilenameFormat = formatFilename
	}
	if ret.Timeout == 0 {
		ret.Timeout = DefaultTimeout
	}
	if ret.TimeoutOnError == 0 {
		ret.TimeoutOnError = DefaultTimeoutOnError
	}
	if ret.Retries < 0 {
		ret.Retries = 0
	} else if ret.Retries == 0 {
		ret.Retries = 3
	}
	return ret
}

var (
	DefaultDownloadClient = &grab.Client{
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
			Timeout: time.Minute * 5,
		},
		// UserAgent from https://explore.whatismybrowser.com/useragents/parse/505617920-tiktok-android-webkit,
		// UserAgent: "Mozilla/5.0 (Linux; Android 13; 2109119DG Build/TKQ1.220829.002; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/119.0.6045.193 Mobile Safari/537.36 trill_320403 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/trill app_version/32.4.3 ByteLocale/en ByteFullLocale/en Region/MY AppId/1180 Spark/1.4.6.3-bugfix AppVersion/32.4.3 BytedanceWebview/d8a21c6",
		// This UserAgent should be less trackable
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Safari/605.1.1",
	}
	DefaultTimeout        = time.Millisecond * 100
	DefaultTimeoutOnError = time.Second * 10
	FilenameFormat        = formatFilename
	downloadSync          = &sync.Mutex{}
)

func (post Post) IsAlbum() bool {
	return len(post.Images) != 0
}

func (post Post) IsVideo() bool {
	return !post.IsAlbum()
}

func (post Post) ContentUrls() []string {
	urls := post.Images
	if post.IsVideo() {
		if post.Hdplay != "" {
			urls = []string{post.Hdplay}
		} else if post.Play != "" {
			urls = []string{post.Play}
		} else {
			urls = []string{post.Wmplay}
		}
	}
	return urls
}

func (post Post) DownloadVideo(opts ...DownloadOpt) (file string, err error) {
	posts, err := post.Download(opts...)
	if len(posts) == 0 {
		return "", err
	}
	return posts[0], err
}

func (post Post) Download(opts ...DownloadOpt) (files []string, err error) {
	opt := &DownloadOpt{}
	if len(opts) != 0 {
		opt = &opts[0]
	}
	opt = opt.Defaults()

	if !opt.NoSync {
		downloadSync.Lock()
		defer downloadSync.Unlock()
	}

	urls := post.ContentUrls()

	files = []string{}
	for i, url := range urls {
		tmp, err := os.Create(path.Join(opt.Directory, FilenameFormat(&post, i)))
		if err != nil {
			return files, err
		}
		files = append(files, tmp.Name())
		if err := tmp.Close(); err != nil {
			return files, err
		}

		if i > 0 {
			time.Sleep(opt.Timeout)
		}
		if err := opt.downloadRetrying(&post, url, tmp.Name(), 0, nil); err != nil {
			return nil, err
		}
	}

	return
}

func formatFilename(post *Post, i int) string {
	filename := fmt.Sprintf("%s_%s_%s", post.Author.UniqueId, time.Unix(post.CreateTime, 0).Format(time.DateOnly), post.ID())
	if post.IsVideo() {
		return filename + ".mp4"
	}
	return fmt.Sprintf("%s_%d.jpg", filename, i+1)
}
