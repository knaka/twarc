package twarc

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"log"

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

type Legacy struct {
	IDStr                string   `json:"id_str"`
	FullText             string   `json:"full_text"`
	InReplyToStatusIDStr string   `json:"in_reply_to_status_id_str"`
	Entities             Entities `json:"entities"`
}

type Tweet struct {
	RestID string `json:"rest_id"`
	Legacy Legacy `json:"legacy"`
}

type Tweets []Tweet

func extractTweets(jsonStr string) (tweets []Tweet) {
	result := gjson.Get(jsonStr, `data.user.result.timeline_v2.timeline.instructions.0.entries.#.content.itemContent.tweet_results.result|@flatten`).String()
	V0(json.Unmarshal([]byte(result), &tweets))
	result = gjson.Get(jsonStr, `data.user.result.timeline_v2.timeline.instructions.0.entries.#.content.items.#.item.itemContent.tweet_results.result|@flatten`).String()
	var conversationTweets []Tweet
	V0(json.Unmarshal([]byte(result), &conversationTweets))
	tweets = append(tweets, conversationTweets...)
	log.Println(tweets)
	return
}
