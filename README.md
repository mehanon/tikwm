# Tikwm API

https://tikwm.com is the best middleman for getting TikTok video info, afaik.

Request syncing with a timeout is built-in, no worries. Other words don't really matter, here's the common code:

## Library Example

```go
package main

import (
  "github.com/mehanon/tikwm/api"
  "log"
  "time"
)

func main() {
  // api.GetVideo(url string, HD bool)
  videoHD, _ := api.GetVideo("https://www.tiktok.com/@locallygrownwig/video/6901498776523951365", true)
  videoHD, _ = api.GetVideo("6901498776523951365", true)              // by ID 
  videoHD, _ = api.GetVideo("https://vm.tiktok.com/ZM66UoB9m/", true) // by shorten link 
  localname, _ = videoHD.Download()

  // Get user posts for the last 30 days
  until := time.Now().Add(-time.Hour * 24 * 30)
  // func GetUserFeedUntilVerbose(uniqueID string, hd bool, pred func(vid *Post) bool, onError func(err error)) (chan Post, error) {
  vidChan, _ := api.GetUserFeedUntilVerbose(url, true,
    func(vid *api.Post) bool {
      return time.Unix(vid.CreateTime, 0).Before(until)
    },
    func(err error) {
      if err != nil {
        log.Fatalf("%s: %s", url, err.Error())
      }
    })

  for vid := range vidChan {
    localname, _ := vid.Download()
    log.Println(localname)
  }
}

```

## Executable Example

* `./tikwm "https://www.tiktok.com/@locallygrownwig/video/6901498776523951365"` -- download this video in HD to current
  folder
* `./tikwm -cmd profile -until "2023-01-01 00:00:00" losertron` -- download all @losertron content from 2023 to now
* `./tikwm -cmd info losertron` -- get user info about @losertron profile

```
$ ./tikmeh
Usage of ./target/tikmeh_linux_amd64_2023-12-23:
  -cmd string
        what to do (video | profile | info) (default "video")
  -folder string
        folder to save files (default "./downloads")
  -json
        print info as json, dont download
  -sd
        don't request HD sources of videos
  -until string
        dont download videos earlier than (default "1970-01-01 00:00:00")

```
