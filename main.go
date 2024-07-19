package twarc

import (
	"encoding/json"
	. "github.com/knaka/go-utils"
	"github.com/tidwall/gjson"
	"strconv"
	"time"
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

func updateTweets(tweets []Tweet) {
	for i, tweet := range tweets {
		tweets[i].ID = V(strconv.ParseInt(tweet.IDStr, 10, 64))
		tweets[i].UserID = V(strconv.ParseInt(tweet.UserIDStr, 10, 64))
		tweets[i].CreatedAt = V(time.Parse(time.RubyDate, tweet.CreatedAtRuby))
		if tweet.InReplyToStatusIDStr != "" {
			tweets[i].InReplyToStatusID = V(strconv.ParseInt(tweet.InReplyToStatusIDStr, 10, 64))
		}
	}
}

func ExtractTweetsFromTweetDetail(jsonStr string) (tweets []Tweet, err error) {
	result := gjson.Get(jsonStr, `data.threaded_conversation_with_injections_v2.instructions.#.entries.#.content.itemContent.tweet_results.result.legacy|@flatten`).String()
	V0(json.Unmarshal([]byte(result), &tweets))
	result = gjson.Get(jsonStr, `data.threaded_conversation_with_injections_v2.instructions.#.entries.#.content.items.#.item.itemContent.tweet_results.result.legacy|@flatten|@flatten`).String()
	if result != "" {
		var conversationTweets []Tweet
		V0(json.Unmarshal([]byte(result), &conversationTweets))
		tweets = append(tweets, conversationTweets...)
	}
	updateTweets(tweets)
	return
}

func ExtractTweetsFromUserTweets(jsonStr string) (tweets []Tweet, err error) {
	defer Catch(&err)
	result := gjson.Get(jsonStr, `data.user.result.timeline_v2.timeline.instructions.#.entries.#.content.itemContent.tweet_results.result.legacy|@flatten`).String()
	V0(json.Unmarshal([]byte(result), &tweets))
	result = gjson.Get(jsonStr, `data.user.result.timeline_v2.timeline.instructions.#.entries.#.content.items.#.item.itemContent.tweet_results.result.legacy|@flatten|@flatten`).String()
	if result != "" {
		var conversationTweets []Tweet
		V0(json.Unmarshal([]byte(result), &conversationTweets))
		tweets = append(tweets, conversationTweets...)
	}
	updateTweets(tweets)
	return tweets, nil
}
