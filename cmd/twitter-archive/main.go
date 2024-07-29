package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/knaka/twarc"
	"html"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	. "github.com/knaka/go-utils"
)

func main() {
	verbose := flag.Bool("v", false, "verbose")
	timeout := flag.Duration("t", 10*time.Minute, "timeout")
	outpath := flag.String("o", "", "output path")
	format := flag.String("f", "docdir", "format")
	page := flag.String("u", "", "username")
	port := flag.Int("p", 0, "port")
	query := flag.String("q", "", "query string")
	flag.Parse()
	//if *page == "" {
	//	user, err := osuser.Current()
	//	if err != nil {
	//		log.Printf("%+v", err)
	//		os.Exit(1)
	//	}
	//	page = &user.Username
	//}
	startOpts := []twarc.StartOption{}
	startOpts = append(startOpts, twarc.WithVerbose(*verbose))
	startOpts = append(startOpts, twarc.WithTimeout(*timeout))
	startOpts = append(startOpts, twarc.WithPort(*port))
	if *page != "" {
		startOpts = append(startOpts, twarc.WithPage(*page))
	}
	if *query != "" {
		startOpts = append(startOpts, twarc.WithQuery(*query))
	}
	tweets, err := twarc.Start(startOpts...)
	if err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}
	switch *format {
	case "docdir":
		re := regexp.MustCompile(`\bhttps://t.co/[a-zA-Z0-9]+\b`)
		var yearPrev int
		var monthPrev time.Month
		now := time.Now()
		var filePostfix string
		var outputPath string
		homeDir := V(os.UserHomeDir())
		var output *os.File
		for _, tweet := range tweets {
			tweetId := tweet.ID
			userId := tweet.UserID
			permanentUrl := fmt.Sprintf("https://twitter.com/%d/status/%d", userId, tweetId)
			t := tweet.CreatedAt
			yearCurrent := t.Year()
			monthCurrent := t.Month()
			if yearPrev != yearCurrent || monthPrev != monthCurrent {
				if now.Year() == yearCurrent && now.Month() == monthCurrent {
					filePostfix = "-tmp"
				} else {
					filePostfix = ""
				}
				outputPath = filepath.Join(
					homeDir,
					"doc",
					fmt.Sprintf("%04d", yearCurrent),
					fmt.Sprintf("%02d00tweets%s.tsv", monthCurrent, filePostfix),
				)
				info, err := os.Stat(outputPath)
				if os.IsNotExist(err) {
					//_, _ = fmt.Fprintf(os.Stdout, "Creates and switches to file %s\n", outputPath)
					_ = output.Close()
					//output, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY, 0644)
					output, err = os.Create(outputPath)
					if err != nil {
						log.Panicf("panic 218a510 (%v)", err)
					}
				} else if err != nil || info.IsDir() {
					log.Panicf("panic a5dc12a (%v)", err)
				} else {
					if filePostfix != "" {
						//_, _ = fmt.Fprintf(os.Stdout, "Removes and switches to %s\n", outputPath)
						_ = os.Remove(outputPath)
						output, err = os.Create(outputPath)
						if err != nil {
							log.Panicf("panic 218a510 (%v)", err)
						}
					} else {
						break
					}
				}
			}
			text := tweet.FullText
			text = strings.ReplaceAll(text, "\n", " ")
			//text = strings.ReplaceAll(text, "\r", " ")
			text = html.UnescapeString(text)
			//fmt.Fprintf(os.Stderr, "ae9be4b: %s\n", text)
			i := 0
			urls := tweet.Entities.URLs
			text = re.ReplaceAllStringFunc(text, func(s string) string {
				ret := s
				if i < len(urls) {
					ret = urls[i].ExpandedURL
					i = i + 1
				}
				return ret
			})
			_, _ = fmt.Fprintf(output, "%s\t%s\t%s\n", permanentUrl, t.Local().Format(time.RFC3339Nano), text)
			yearPrev = yearCurrent
			monthPrev = monthCurrent
		}
	case "csv":
		var out io.Writer
		if *outpath == "" {
			out = os.Stdout
		} else {
			file := V(os.Create(*outpath))
			defer func() { V0(file.Close()) }()
			out = file
		}
		writer := csv.NewWriter(out)
		defer writer.Flush()
		V0(writer.Write([]string{"tweet_id", "text", "language", "type", "bookmark_count", "favorite_count", "retweet_count", "reply_count", "view_count", "created_at", "client", "hashtags", "urls", "media_type", "media_urls"}))
		for _, tweet := range tweets {
			entities := tweet.Entities
			urls := ""
			delim := ""
			for _, url := range entities.URLs {
				urls += fmt.Sprintf("%s%s", delim, url.ExpandedURL)
				delim = ","
			}
			mediaType := ""
			mediaUrls := ""
			delim = ""
			for _, media := range entities.Media {
				mediaType = "photo"
				mediaUrls += fmt.Sprintf("%s%s", delim, media.MediaURLHTTPS)
			}
			hashtags := ""
			delim = ""
			for _, hashtag := range entities.Hashtags {
				hashtags += fmt.Sprintf("%s#%s", delim, hashtag.Text)
				delim = ","
			}
			V0(writer.Write([]string{
				fmt.Sprintf("%d", tweet.ID),
				tweet.FullText,
				tweet.Lang,
				"Tweet",
				fmt.Sprintf("%d", tweet.BookmarkCount),
				fmt.Sprintf("%d", tweet.FavoriteCount),
				fmt.Sprintf("%d", tweet.RetweetCount),
				fmt.Sprintf("%d", tweet.ReplyCount),
				fmt.Sprintf("%d", tweet.ViewCount),
				tweet.CreatedAt.Local().Format("2006-01-02 15:04:05"),
				tweet.Source,
				hashtags,
				urls,
				mediaType,
				mediaUrls,
			}))
		}
	}
}
