package api

import "log"

type Predicate func(vid *Post) bool

func GetUserFeedAll(uniqueID string, hd bool) ([]Post, error) {
	videoChan, err := GetUserFeedUntilVerbose(uniqueID, hd, nil, nil)
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
func GetUserFeedUntilVerbose(uniqueID string, hd bool, pred func(vid *Post) bool, onError func(err error)) (chan Post, error) {
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

	videos, err := userFeedUntilInternal(uniqueID, "0", pred)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(videos)/2; i++ {
		videos[i], videos[len(videos)-i-1] = videos[len(videos)-i-1], videos[i]
	}

	videoChan := make(chan Post, 100)
	go func() {
		defer func() {
			if r := recover(); r != nil {
			}
		}()

		defer close(videoChan)
		for _, vid := range videos {
			if !hd {
				videoChan <- vid
				continue
			}

			vidHD, err := GetPost(vid.VideoId, true)
			if err != nil {
				onError(err)
				videoChan <- vid
			} else {
				videoChan <- *vidHD
			}
		}
	}()

	return videoChan, err
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
