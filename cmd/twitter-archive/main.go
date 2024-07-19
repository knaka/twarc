package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/tidwall/gjson"
	"log"
	neturl "net/url"
	"os"
	osuser "os/user"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"sync"
	"time"

	. "github.com/knaka/go-utils"
)

type URL struct {
	URL         string `json:"url"`
	ExpandedURL string `json:"expanded_url"`
}

type Media struct {
	URL           string `json:"url"`
	ExpandedURL   string `json:"expanded_url"`
	MediaURLHTTPS string `json:"media_url_https"`
}

type Entities struct {
	URLs  []URL   `json:"urls"`
	Media []Media `json:"media"`
}

type Tweet struct {
	IDStr                string `json:"id_str"`
	ID                   int64
	FullText             string `json:"full_text"`
	InReplyToStatusIDStr string `json:"in_reply_to_status_id_str"`
	InReplyToStatusID    int64
	Entities             Entities `json:"entities"`
	UserIDStr            string   `json:"user_id_str"`
	UserID               int64
	CreatedAtRuby        string `json:"created_at"`
	CreatedAt            time.Time
}

func updateTweets(tweets []*Tweet) {
	for i, tweet := range tweets {
		tweets[i].ID = V(strconv.ParseInt(tweet.IDStr, 10, 64))
		tweets[i].UserID = V(strconv.ParseInt(tweet.UserIDStr, 10, 64))
		tweets[i].CreatedAt = V(time.Parse(time.RubyDate, tweet.CreatedAtRuby))
		if tweet.InReplyToStatusIDStr != "" {
			tweets[i].InReplyToStatusID = V(strconv.ParseInt(tweet.InReplyToStatusIDStr, 10, 64))
		}
	}
}

func ExtractTweetsFromTweetDetail(jsonStr string) (tweets []*Tweet, err error) {
	result := gjson.Get(jsonStr, `data.threaded_conversation_with_injections_v2.instructions.#.entries.#.content.itemContent.tweet_results.result.legacy|@flatten`).String()
	V0(json.Unmarshal([]byte(result), &tweets))
	result = gjson.Get(jsonStr, `data.threaded_conversation_with_injections_v2.instructions.#.entries.#.content.items.#.item.itemContent.tweet_results.result.legacy|@flatten|@flatten`).String()
	if result != "" {
		var conversationTweets []*Tweet
		V0(json.Unmarshal([]byte(result), &conversationTweets))
		tweets = append(tweets, conversationTweets...)
	}
	updateTweets(tweets)
	return
}

func ExtractTweetsFromUserTweets(jsonStr string) (tweets []*Tweet, conversationIds []string, err error) {
	defer Catch(&err)
	result := gjson.Get(jsonStr, `data.user.result.timeline_v2.timeline.instructions.#.entries.#.content.itemContent.tweet_results.result.legacy|@flatten`).String()
	V0(json.Unmarshal([]byte(result), &tweets))
	result = gjson.Get(jsonStr, `data.user.result.timeline_v2.timeline.instructions.#.entries.#.content.items.#.item.itemContent.tweet_results.result.legacy|@flatten|@flatten`).String()
	if result != "" {
		var conversationTweets []*Tweet
		V0(json.Unmarshal([]byte(result), &conversationTweets))
		tweets = append(tweets, conversationTweets...)
	}
	updateTweets(tweets)
	return
}

func getTypeName(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}

type StartFuncParams struct {
	Verbose  bool
	Port     int
	Duration time.Duration
}

type StartOption func(*StartFuncParams)

func WithVerbose(verbose bool) StartOption {
	return func(opts *StartFuncParams) {
		opts.Verbose = verbose
	}
}

func WithTimeout(timeout time.Duration) StartOption {
	return func(opts *StartFuncParams) {
		opts.Duration = timeout
	}
}

//func WithDataHandler(matchRegex string, dataFunc func(string, string)) StartOption {
//	return func(opts *StartFuncParams) {
//		opts.DataHandlers = append(opts.DataHandlers, DataHandler{
//			ReMatch:    regexp.MustCompile(matchRegex),
//			DataFunc:   dataFunc,
//			TargetUrls: make(TargetUrls),
//		})
//	}
//}

func WithPort(port int) StartOption {
	return func(opts *StartFuncParams) {
		opts.Port = port
	}
}

func Open(ctx context.Context, targetUrl string, matchRegex string, cb func(url string, body string)) {
	cdpCtx := chromedp.FromContext(ctx)
	ctxExec := cdp.WithExecutor(ctx, cdpCtx.Target)
	reMatch := regexp.MustCompile(matchRegex)
	targetUrls := make(map[network.RequestID]string)
	chromedp.ListenTarget(ctx, func(ev any) {
		switch event := ev.(type) {
		case *network.EventRequestWillBeSent:
			if reMatch.MatchString(event.Request.URL) {
				log.Println("Matched", event.Request.URL)
				targetUrls[event.RequestID] = event.Request.URL
			}
		case *network.EventLoadingFinished:
			if url, ok := targetUrls[event.RequestID]; ok {
				log.Println("Fetched", url)
				go func() {
					cb(url, string(V(network.GetResponseBody(event.RequestID).Do(ctxExec))))
				}()
			}
		}
	})
	V0(chromedp.Run(ctx, chromedp.Navigate(targetUrl)))
}

var mu = sync.Mutex{}
var tweets []*Tweet

func Start(
	targetUrl string,
	startOpts ...StartOption,
) (err error) {
	defer Catch(&err)

	funcParams := StartFuncParams{
		Verbose:  false,
		Port:     9222,
		Duration: 10 * time.Minute,
	}
	for _, startOpt := range startOpts {
		startOpt(&funcParams)
	}
	var cdpCtxOpts []chromedp.ContextOption
	if funcParams.Verbose {
		cdpCtxOpts = append(cdpCtxOpts, chromedp.WithDebugf(log.Printf))
	}
	debugUrl := neturl.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", "127.0.0.1", funcParams.Port),
	}
	ctx, cancel := chromedp.NewRemoteAllocator(context.Background(), debugUrl.String())
	defer cancel()
	if funcParams.Duration > 0 {
		ctx, cancel = context.WithTimeout(ctx, funcParams.Duration)
		defer cancel()
	}

	waitGroup := sync.WaitGroup{}

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()
	// Not `ListenTarget` but `ListenBrowser` can detect a `page` target detached.
	chromedp.ListenBrowser(ctx, func(ev any) {
		switch ev.(type) {
		case *target.EventDetachedFromTarget:
			waitGroup.Done()
		}
	})

	// `Run` to acquire the browser.
	V0(chromedp.Run(ctx))
	// `Run` has opened a new `page` target.
	waitGroup.Add(1)

	Open(ctx, targetUrl, `/UserTweets\?`, func(url string, data string) {
		tweetsNew, convIds := V2(ExtractTweetsFromUserTweets(data))
		mu.Lock()
		defer mu.Unlock()
		tweets = append(tweets, tweetsNew...)
		// If some conversations have missing tweets.
		for _, convId := range convIds {
			// Should handle cancel?
			ctxChild, _ := chromedp.NewContext(ctx)
			Open(ctxChild, fmt.Sprintf("https://x.com/i/status/%d", convId), `/TweetDetail\?`, func(url string, data string) {
				tweetsNew := V(ExtractTweetsFromTweetDetail(data))
				mu.Lock()
				defer mu.Unlock()
				tweets = append(tweets, tweetsNew...)
			})
			waitGroup.Add(1)
		}
	})

	waitGroup.Wait()

	return nil
}

func main() {
	verbose := flag.Bool("v", false, "verbose")
	timeout := flag.Duration("timeout", 10*time.Minute, "timeout")
	page := flag.String("page", "", "page")
	if *page == "" {
		user, err := osuser.Current()
		if err != nil {
			log.Printf("%+v", err)
			os.Exit(1)
		}
		page = &user.Username
	}
	url := neturl.URL{
		Scheme: "https",
		Host:   "x.com",
		Path:   fmt.Sprintf("/%s", *page),
	}
	if err := Start(
		url.String(),
		WithVerbose(*verbose),
		WithTimeout(*timeout),
	); err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}
	time.Sleep(time.Second * 1)
	sort.Slice(tweets, func(i, j int) bool {
		return tweets[i].CreatedAt.After(tweets[j].CreatedAt)
	})
	tweets = slices.CompactFunc(tweets, func(i, j *Tweet) bool {
		return i.ID == j.ID
	})
	for _, tweet := range tweets {
		fmt.Println(tweet.CreatedAt, tweet.FullText)
	}
}
