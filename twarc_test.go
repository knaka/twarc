package twarc

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/knaka/go-utils"
)

func Test_extractTweets(t *testing.T) {
	json := string(V(os.ReadFile(filepath.Join("testdata", "user-tweets.json"))))
	//json2 := string(V(os.ReadFile(filepath.Join("testdata", "user-tweets2.json"))))
	json3 := string(V(os.ReadFile(filepath.Join("testdata", "user-tweets3.json"))))
	type args struct {
		json string
	}
	tests := []struct {
		name        string
		args        args
		wantTweets  int
		wantConvIds int
	}{
		{
			"Test 1",
			args{json},
			21,
			1,
		},
		{
			"Test 3",
			args{json3},
			19,
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTweets, convIds, err := extractTweetsFromUserTweets(tt.args.json); err != nil || len(gotTweets) != tt.wantTweets || len(convIds) != tt.wantConvIds {
				t.Errorf("ExtractTweets() = %v, want %v", len(gotTweets), tt.wantTweets)
			}
		})
	}
}

func TestExtractTweetsFromTweetDetail(t *testing.T) {
	json4 := string(V(os.ReadFile(filepath.Join("testdata", "tweet-detail.json"))))
	type args struct {
		jsonStr string
	}
	tests := []struct {
		name       string
		args       args
		wantTweets int
		wantErr    bool
	}{
		{
			"Test 4",
			args{json4},
			4,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTweets, err := extractTweetsFromTweetDetail(tt.args.jsonStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTweetsFromTweetDetail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotTweets) != tt.wantTweets {

				t.Errorf("extractTweetsFromTweetDetail() = %v, want %v", len(gotTweets), tt.wantTweets)
			}
		})
	}
}

func Test_extractTweetsFromSearchTimeline(t *testing.T) {
	jsonStr := string(V(os.ReadFile(filepath.Join("testdata", "search-timeline.json"))))
	type args struct {
		jsonStr string
	}
	tests := []struct {
		name             string
		args             args
		wantLegacyTweets int
		wantErr          bool
	}{
		{
			"Test 1",
			args{jsonStr},
			10,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLegacyTweets, err := extractTweetsFromSearchTimeline(tt.args.jsonStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTweetsFromSearchTimeline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotLegacyTweets) != tt.wantLegacyTweets {
				t.Errorf("extractTweetsFromSearchTimeline() = %v, want %v", len(gotLegacyTweets), tt.wantLegacyTweets)
			}
		})
	}
}
