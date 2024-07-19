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
		name       string
		args       args
		wantTweets int
	}{
		{
			"Test 1",
			args{json},
			21,
		},
		{
			"Test 3",
			args{json3},
			19,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTweets, err := ExtractTweetsFromUserTweets(tt.args.json); err != nil || len(gotTweets) != tt.wantTweets {
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
			gotTweets, err := ExtractTweetsFromTweetDetail(tt.args.jsonStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractTweetsFromTweetDetail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotTweets) != tt.wantTweets {

				t.Errorf("ExtractTweetsFromTweetDetail() = %v, want %v", len(gotTweets), tt.wantTweets)
			}
		})
	}
}
