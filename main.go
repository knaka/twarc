package twarc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/tidwall/gjson"
	"log"
	neturl "net/url"
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

type Hashtag struct {
	Text string `json:"text"`
}

type Entities struct {
	URLs     []URL     `json:"urls"`
	Media    []Media   `json:"media"`
	Hashtags []Hashtag `json:"hashtags"`
}

// Tweet represents a tweet.
type Tweet struct {
	ID                int64
	FullText          string `json:"full_text"`
	Lang              string `json:"lang"`
	BookmarkCount     int    `json:"bookmark_count"`
	FavoriteCount     int    `json:"favorite_count"`
	RetweetCount      int    `json:"retweet_count"`
	ReplyCount        int    `json:"reply_count"`
	ViewCount         int    `json:"view_count"`
	Source            string `json:"source"`
	InReplyToStatusID int64
	Entities          Entities `json:"entities"`
	UserID            int64
	CreatedAt         time.Time
}

// legacyTweet deserializes a tweet in response JSON.
type legacyTweet struct {
	Tweet
	IDStr                string `json:"id_str"`
	InReplyToStatusIDStr string `json:"in_reply_to_status_id_str"`
	UserIDStr            string `json:"user_id_str"`
	CreatedAtRuby        string `json:"created_at"`
}

func extractTweetsFromSearchTimeline(jsonStr string) (legacyTweets []*legacyTweet, err error) {
	legacyTweetsJson := gjson.Get(jsonStr, "data.search_by_raw_query.search_timeline.timeline.instructions.#.entries.#.content.itemContent.tweet_results.result.legacy|@flatten")
	legacyTweetsJson.ForEach(func(_, legacyTweetJson gjson.Result) (proceeds bool) {
		var legacyTweetNew legacyTweet
		V0(json.Unmarshal([]byte(legacyTweetJson.String()), &legacyTweetNew))
		legacyTweets = append(legacyTweets, &legacyTweetNew)
		return true
	})
	return

}

func extractTweetsFromTweetDetail(jsonStr string) (legacyTweets []*legacyTweet, err error) {
	contentsJson := gjson.Get(jsonStr, "data.threaded_conversation_with_injections_v2.instructions.#.entries.#.content|@flatten")
	contentsJson.ForEach(func(_, contentJson gjson.Result) (proceeds bool) {
		if contentJson.Get("itemContent").Exists() {
			legacyTweetJson := contentJson.Get("itemContent.tweet_results.result.legacy")
			var legacyTweetNew legacyTweet
			V0(json.Unmarshal([]byte(legacyTweetJson.String()), &legacyTweetNew))
			legacyTweets = append(legacyTweets, &legacyTweetNew)
		}
		if contentJson.Get("items").Exists() {
			legacyTweetsJson := contentJson.Get("items.#.item.itemContent.tweet_results.result.legacy")
			var legacyTweetsNew []*legacyTweet
			V0(json.Unmarshal([]byte(legacyTweetsJson.String()), &legacyTweetsNew))
			legacyTweets = append(legacyTweets, legacyTweetsNew...)
		}
		return true
	})
	return
}

func extractTweetsFromUserTweets(jsonStr string) (legacyTweets []*legacyTweet, conversationIds []int64, err error) {
	defer Catch(&err)
	contentsJson := gjson.Get(jsonStr, "data.user.result.timeline_v2.timeline.instructions.#.entries.#.content|@flatten")
	contentsJson.ForEach(func(_, contentJson gjson.Result) (proceeds bool) {
		if contentJson.Get("itemContent").Exists() {
			legacyTweetJson := contentJson.Get("itemContent.tweet_results.result.legacy")
			var legacyTweetNew legacyTweet
			V0(json.Unmarshal([]byte(legacyTweetJson.String()), &legacyTweetNew))
			legacyTweets = append(legacyTweets, &legacyTweetNew)
		}
		if contentJson.Get("items").Exists() {
			legacyTweetsJson := contentJson.Get("items.#.item.itemContent.tweet_results.result.legacy")
			allTweetIdsJson := contentJson.Get("metadata.conversationMetadata.allTweetIds")
			var legacyTweetsNew []*legacyTweet
			V0(json.Unmarshal([]byte(legacyTweetsJson.String()), &legacyTweetsNew))
			legacyTweets = append(legacyTweets, legacyTweetsNew...)
			if len(allTweetIdsJson.Array()) > len(legacyTweetsJson.Array()) {
				conversationIds = append(conversationIds, allTweetIdsJson.Array()[0].Int())
			}
		}
		return true
	})
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
	Page     string
	Query    string
}

type StartOption func(*StartFuncParams)

func WithPage(page string) StartOption {
	return func(opts *StartFuncParams) {
		opts.Page = page
	}
}

func WithQuery(query string) StartOption {
	return func(opts *StartFuncParams) {
		opts.Query = query
	}
}

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

func WithPort(port int) StartOption {
	return func(opts *StartFuncParams) {
		opts.Port = port
	}
}

func Open(ctx context.Context, targetUrl string, matchRegex string, cb func(url string, body string)) {
	reMatch := regexp.MustCompile(matchRegex)
	targetUrls := make(map[network.RequestID]string)
	chromedp.ListenTarget(ctx, func(ev any) {
		//typeName := getTypeName(ev)
		//log.Println("Event type:", typeName)
		switch event := ev.(type) {
		case *network.EventRequestWillBeSent:
			if reMatch.MatchString(event.Request.URL) {
				log.Println("Matched", event.Request.URL[:100], "...")
				targetUrls[event.RequestID] = event.Request.URL
			}
		case *network.EventLoadingFinished:
			if url, ok := targetUrls[event.RequestID]; ok {
				log.Println("Fetched", url[:100], "...")
				go func() {
					cdpCtx := chromedp.FromContext(ctx)
					ctxExec := cdp.WithExecutor(ctx, cdpCtx.Target)
					body := string(V(network.GetResponseBody(event.RequestID).Do(ctxExec)))
					cb(url, body)
				}()
			}
		}
	})
	V0(chromedp.Run(ctx, chromedp.Navigate(targetUrl)))
}

func Start(
	startOpts ...StartOption) (tweets []*Tweet, err error) {
	defer Catch(&err)

	var legacyTweets []*legacyTweet
	mu := sync.Mutex{}

	funcParams := StartFuncParams{
		Verbose:  false,
		Port:     9222,
		Duration: 10 * time.Minute,
	}
	for _, startOpt := range startOpts {
		startOpt(&funcParams)
	}
	if funcParams.Page == "" && funcParams.Query == "" {
		return nil, fmt.Errorf("page or query must be specified")
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
	if funcParams.Page != "" {
		url := neturl.URL{
			Scheme: "https",
			Host:   "x.com",
			Path:   fmt.Sprintf("/%s", funcParams.Page),
		}
		targetUrl := url.String()
		Open(ctx, targetUrl, `/UserTweets\?`, func(url string, data string) {
			legacyTweetsNew, convIds := V2(extractTweetsFromUserTweets(data))
			mu.Lock()
			defer mu.Unlock()
			legacyTweets = append(legacyTweets, legacyTweetsNew...)
			if funcParams.Verbose {
				go func() {
					for _, t := range legacyTweetsNew {
						log.Println(t.CreatedAtRuby, t.FullText)
					}
				}()
			}
			// If some conversations have missing legacyTweets.
			for _, convId := range convIds {
				// Should handle cancel?
				ctxChild, cancel := chromedp.NewContext(ctx)
				Open(ctxChild, fmt.Sprintf("https://x.com/i/status/%d", convId), `/TweetDetail\?`, func(url string, data string) {
					legacyTweetsNew := V(extractTweetsFromTweetDetail(data))
					mu.Lock()
					defer mu.Unlock()
					legacyTweets = append(legacyTweets, legacyTweetsNew...)
					if funcParams.Verbose {
						go func() {
							for _, t := range legacyTweetsNew {
								log.Println(t.CreatedAtRuby, t.FullText)
							}
						}()
					}
					cancel()
				})
				waitGroup.Add(1)
			}
		})
	} else {
		url := neturl.URL{
			Scheme:   "https",
			Host:     "x.com",
			Path:     "search",
			RawQuery: fmt.Sprintf("q=%s", neturl.QueryEscape(funcParams.Query)),
		}
		targetUrl := url.String()
		Open(ctx, targetUrl, `/SearchTimeline\?`, func(url string, data string) {
			legacyTweetsNew := V(extractTweetsFromSearchTimeline(data))
			mu.Lock()
			defer mu.Unlock()
			legacyTweets = append(legacyTweets, legacyTweetsNew...)
			if funcParams.Verbose {
				go func() {
					for _, t := range legacyTweetsNew {
						log.Println(t.CreatedAtRuby, t.FullText)
					}
				}()
			}
		})
	}
	waitGroup.Wait()

	// Update missing fields.
	for _, t := range legacyTweets {
		t.ID = V(strconv.ParseInt(t.IDStr, 10, 64))
		t.UserID = V(strconv.ParseInt(t.UserIDStr, 10, 64))
		t.CreatedAt = V(time.Parse(time.RubyDate, t.CreatedAtRuby))
		if t.InReplyToStatusIDStr != "" {
			t.InReplyToStatusID = V(strconv.ParseInt(t.InReplyToStatusIDStr, 10, 64))
		}
	}
	// Sort.
	sort.Slice(legacyTweets, func(i, j int) bool {
		return legacyTweets[i].CreatedAt.After(legacyTweets[j].CreatedAt)
	})
	// Uniq.
	legacyTweets = slices.CompactFunc(legacyTweets, func(i, j *legacyTweet) bool {
		return i.ID == j.ID
	})
	// Extract Tweet-s.
	for _, t := range legacyTweets {
		tweets = append(tweets, &t.Tweet)
	}
	return
}
