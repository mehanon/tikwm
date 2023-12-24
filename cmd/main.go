package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mehanon/tikwm/api"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	time.Now().Add(-time.Hour * 24 * 30)

	cmd := flag.String("cmd", "video", "what to do (video | profile | info)")
	until := flag.String("until", "1970-01-01 00:00:00", "dont download videos earlier than")
	sd := flag.Bool("sd", false, "don't request HD sources of videos (less requests => notably faster)")
	folder := flag.String("folder", "./", "folder to save files")
	json_ := flag.Bool("json", false, "print info as json, dont download")
	flag.Parse()

	urls := flag.Args()
	if len(urls) == 0 {
		println("no arguments were passed, use -help to get help")
		os.Exit(0)
	}

	for _, url := range urls {
		ensureDir(*folder)

		switch {
		case *cmd == "video":
			vid, err := api.GetPost(url, !*sd)
			if err != nil {
				log.Fatalf("%s: %s", url, err.Error())
			}

			if *json_ {
				buffer, err := json.MarshalIndent(vid, "", "\t")
				if err != nil {
					log.Fatalf("%s: %s", url, err.Error())
				}
				print(string(buffer))
				continue
			}

			filename, err := vid.Download(*folder)
			if err != nil {
				log.Fatalf("%s: %s", url, err.Error())
			}
			log.Printf("%s: %s", url, filename)

		case *cmd == "profile":
			until, err := time.Parse(time.DateTime, *until)
			if err != nil {
				log.Fatalf("%s: %s", url, err.Error())
			}

			vidChan, err := api.GetUserFeedUntilVerbose(url, !*sd, func(vid *api.Post) bool {
				return time.Unix(vid.CreateTime, 0).Before(until)
			}, func(err error) {
				if err != nil {
					log.Fatalf("%s: %s", url, err.Error())
				}
			})

			jsonRet := []string{}
			for vid := range vidChan {
				if *json_ {
					buffer, err := json.MarshalIndent(vid, "", "\t")
					if err != nil {
						log.Fatalf("%s: %s", url, err.Error())
					}
					jsonRet = append(jsonRet, string(buffer))
					continue
				}

				filename, err := vid.Download(*folder)
				if err != nil {
					log.Fatalf("%s: %s", url, err.Error())
				}
				log.Printf("%s: %s", url, filename)
			}

			if *json_ {
				fmt.Printf("[%s]", strings.Join(jsonRet, ",\n"))
			}

		case *cmd == "info":
			vid, err := api.GetUserDetail(url)
			if err != nil {
				log.Fatalf("%s: %s", url, err.Error())
			}

			buffer, err := json.MarshalIndent(vid, "", "\t")
			if err != nil {
				log.Fatalf("%s: %s", url, err.Error())
			}
			print(string(buffer))
		}
	}
}

func ensureDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			log.Fatalln(err)
		}
	}
}
