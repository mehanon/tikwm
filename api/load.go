package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
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
		if post.Wmplay == "" {
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
		fileFormat := fmt.Sprintf("%s_%s_%d_*%s", post.Author.UniqueId, time.Unix(post.CreateTime, 0).Format(time.DateOnly), i, fileType)

		tmp, err := os.CreateTemp(dir, fileFormat)
		if err != nil {
			return urls, err
		}
		defer tmp.Close()

		resp, err := http.Get(url)
		if err != nil {
			return urls, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return urls, fmt.Errorf("bad status: %s", resp.Status)
		}
		if _, err = io.Copy(tmp, resp.Body); err != nil {
			return urls, err
		}

		files = append(files, tmp.Name())
	}

	return
}
