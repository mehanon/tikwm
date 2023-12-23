package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	URL              string        = "https://tikwm.com/api"
	Timeout          time.Duration = time.Second
	MaxUserFeedCount int           = 34
	requestSync      *sync.Mutex   = &sync.Mutex{}
)

func Raw(method string, query map[string]string) ([]byte, error) {
	requestSync.Lock()
	defer unlock()

	url := fmt.Sprintf("%s/%s", URL, method)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for key, val := range query {
		q.Add(key, val)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func RawParsed[T any](method string, query map[string]string) (*T, error) {
	data, err := Raw(method, query)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Code          int     `json:"code"`
		Msg           string  `json:"msg"`
		ProcessedTime float64 `json:"processed_time"`
		Data          *T      `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		queryStr := "???"
		if buf, err := json.Marshal(query); err == nil {
			queryStr = string(buf)
		}
		return nil, fmt.Errorf("tikwm error: %s (%d) [query: %s]", resp.Msg, resp.Code, queryStr)
	}

	return resp.Data, nil
}

func GetPost(url string, hd bool) (*Post, error) {
	query := map[string]string{"url": url}
	if hd {
		query["hd"] = "1"
	}
	return RawParsed[Post]("", query)
}

func GetUserFeed(uniqueID string, count int, cursor string) (*UserFeed, error) {
	query := map[string]string{"unique_id": uniqueID, "count": strconv.Itoa(count), "cursor": cursor}
	return RawParsed[UserFeed]("user/posts", query)
}

func GetUserDetail(uniqueID string) (*UserDetail, error) {
	query := map[string]string{"unique_id": uniqueID}
	return RawParsed[UserDetail]("user/info", query)
}

func unlock() {
	time.Sleep(Timeout)
	requestSync.Unlock()
}
