package twarc

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/knaka/go-utils"
)

func Test_extractTweets(t *testing.T) {
	json := string(V(os.ReadFile(filepath.Join("testdata", "user-tweets.json"))))
	//json := string(V(os.ReadFile(filepath.Join("testdata", "user-tweets2.json"))))
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTweets, err := ExtractTweets(tt.args.json); err != nil || len(gotTweets) != tt.wantTweets {
				t.Errorf("ExtractTweets() = %v, want %v", len(gotTweets), tt.wantTweets)
			}
		})
	}
}
