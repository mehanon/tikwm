package api

import (
	"log"
)

type Predicate func(vid *Post) bool

func GetUserFeedAll(uniqueID string, hd bool) ([]Post, error) {
	videoChan, _, err := GetUserFeedUntilVerbose(uniqueID, hd, nil, nil)
	if err != nil {
		return nil, err
	}
	ret := []Post{}
	for vid := range videoChan {
		ret = append(ret, vid)
	}
	return ret, nil
}

// GetUserFeedUntilVerbose -- pred = nil => download all, onError = nil => log and skip errors
func GetUserFeedUntilVerbose(uniqueID string, hd bool, pred func(vid *Post) bool, onError func(err error)) (chan Post, int, error) {
	if pred == nil {
		pred = func(vid *Post) bool {
			return false
		}
	}
	if onError == nil {
		onError = func(err error) {
			if err != nil {
				log.Println(err)
			}
		}
	}

	posts, err := userFeedUntilInternal(uniqueID, "0", pred)
	if err != nil {
		return nil, 0, err
	}
	for i := 0; i < len(posts)/2; i++ {
		posts[i], posts[len(posts)-i-1] = posts[len(posts)-i-1], posts[i]
	}

	postChan := make(chan Post, 100)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// It's ok to panic in onError to interrupt the loop,
				// nothing else would panic (95% sure).
			}
		}()

		defer close(postChan)
		for _, vid := range posts {
			if !hd {
				postChan <- vid
				continue
			}

			vidHD, err := GetPost(vid.VideoId, true)
			if err != nil {
				onError(err)
				postChan <- vid
			} else {
				postChan <- *vidHD
			}
		}
	}()

	return postChan, len(posts), err
}

func userFeedUntilInternal(uniqueID string, cursor string, pred Predicate) ([]Post, error) {
	feed, err := GetUserFeed(uniqueID, MaxUserFeedCount, cursor)
	if err != nil {
		return nil, err
	}

	ret := []Post{}
	for _, vid := range feed.Videos {
		if pred(&vid) {
			return ret, nil
		}
		ret = append(ret, vid)
	}

	if !feed.HasMore {
		return ret, nil
	}

	deeperRet, err := userFeedUntilInternal(uniqueID, feed.Cursor, pred)
	if err != nil {
		return ret, err
	}

	return append(ret, deeperRet...), nil
}
