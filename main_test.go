package twarc

import (
	"os"
	"testing"

	. "github.com/knaka/go-utils"
)

func Test_extractTweets(t *testing.T) {
	json := string(V(os.ReadFile("/tmp/test.json")))
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
			10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTweets := extractTweets(tt.args.json); len(gotTweets) != tt.wantTweets {
				t.Errorf("extractTweets() = %v, want %v", len(gotTweets), tt.wantTweets)
			}
		})
	}
}
