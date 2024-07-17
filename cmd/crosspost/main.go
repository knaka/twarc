package main

import (
	"context"
	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
	"github.com/igolaizola/twai/pkg/twitter"
	"log"
	"os"
	"os/signal"
	"time"

	. "github.com/knaka/go-utils"
)

func main() {
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()
	ctx, cancel = signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	b := twitter.NewBrowser(&twitter.BrowserConfig{
		Wait:        1 * time.Second,
		CookieStore: twitter.NewCookieStore("/Users/knaka/tmp/x-cookies.txt"),
		Headless:    false,
		BinPath:     "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
	})
	V0(b.Start(ctx))
	chromedp.ListenTarget(ctx, func(v any) {
		if ev, ok := v.(*browser.EventDownloadProgress); ok {
			if ev.TotalBytes != 0 {
				log.Println("done")
			}
		}
	})
	posts := V(b.Posts(ctx, "knaka", 10, false))
	log.Println(posts)
}
