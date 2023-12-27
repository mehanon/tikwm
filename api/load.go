package api

import (
	"fmt"
	"github.com/cavaliergopher/grab/v3"
	"net/http"
	"os"
	"path"
	"time"
)

var (
	DefaultClient = &grab.Client{
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
			Timeout: time.Minute * 5,
		},
		// UserAgent from https://explore.whatismybrowser.com/useragents/parse/505617920-tiktok-android-webkit
		UserAgent: "Mozilla/5.0 (Linux; Android 13; 2109119DG Build/TKQ1.220829.002; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/119.0.6045.193 Mobile Safari/537.36 trill_320403 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/trill app_version/32.4.3 ByteLocale/en ByteFullLocale/en Region/MY AppId/1180 Spark/1.4.6.3-bugfix AppVersion/32.4.3 BytedanceWebview/d8a21c6",
	}
	TimeoutDownload = time.Second
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
		} else if post.Wmplay != "" {
			urls = []string{post.Wmplay}
		} else {
			urls = []string{post.Play}
		}
	}
	return urls
}

func (post Post) DownloadVideo(directory ...string) (file string, err error) {
	posts, err := post.Download(directory...)
	if len(posts) == 0 {
		return "", err
	}
	return posts[0], err
}

func (post Post) Download(directory ...string) (files []string, err error) {
	urls := post.ContentUrls()
	dir := ""
	if len(directory) != 0 {
		dir = directory[0]
	}

	fileType := ""
	if post.IsAlbum() {
		fileType = ".jpg"
	} else {
		fileType = ".mp4"
	}

	files = []string{}
	for i, url := range urls {
		fileFormat := fmt.Sprintf("%s_%s_%s", post.Author.UniqueId, time.Unix(post.CreateTime, 0).Format(time.DateOnly), post.Id)
		if len(urls) > 1 {
			fileFormat += fmt.Sprintf("_%d", i+1)
		}
		fileFormat += fileType

		tmp, err := os.Create(path.Join(dir, fileFormat))
		if err != nil {
			return files, err
		}
		files = append(files, tmp.Name())
		if err := tmp.Close(); err != nil {
			return files, err
		}

		req, err := grab.NewRequest(tmp.Name(), url)
		if err != nil {
			return files, err
		}
		resp := DefaultClient.Do(req)
		<-resp.Done
		if resp.Err() != nil {
			return files, err
		}
		// ensuring no silent 429
		if i+1 == len(urls) {
			time.Sleep(TimeoutDownload)
		}
	}

	return
}
