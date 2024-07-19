package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/knaka/twarc"
	"io"
	"log"
	"os"
	"time"

	. "github.com/knaka/go-utils"
)

func main() {
	verbose := flag.Bool("v", false, "verbose")
	timeout := flag.Duration("timeout", 10*time.Minute, "timeout")
	outpath := flag.String("o", "", "outpath")
	format := flag.String("f", "csv", "format")
	page := flag.String("page", "", "page")
	query := flag.String("q", "", "query")
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
	if *format == "csv" {
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
