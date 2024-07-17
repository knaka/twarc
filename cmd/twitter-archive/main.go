package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/knaka/twarc"
	"log"
	"os"
	"strings"
	"time"

	. "github.com/knaka/go-utils"
)

func Main(timeout time.Duration, headless, verbose bool) (err error) {
	defer Catch(&err)

	var opts []chromedp.ContextOption
	if verbose {
		opts = append(opts, chromedp.WithDebugf(log.Printf))
	}
	ctx, cancel := chromedp.NewRemoteAllocator(context.Background(), "ws://127.0.0.1:9222")
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx, opts...)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	ch, errCh := make(chan string, 1), make(chan error, 1)
	urls := make(map[network.RequestID]string)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if verbose {
			log.Printf("%T: %+v\n", ev, ev)
		}
		switch event := ev.(type) {
		case *network.EventRequestWillBeSent:
			url := event.Request.URL
			if strings.Contains(url, "/UserTweets?") {
				urls[event.RequestID] = url
			}
		//case *network.EventResponseReceived:
		case *network.EventLoadingFinished:
			//case *network.EventDataReceived:
			//log.Println(event.Data)
			if _, ok := urls[event.RequestID]; ok {
				fc := chromedp.FromContext(ctx)
				ctx := cdp.WithExecutor(ctx, fc.Target)
				go func() {
					byts, err := network.GetResponseBody(event.RequestID).Do(ctx)
					if err != nil {
						panic(err)
					}
					data := string(byts)
					tweets := V(twarc.ExtractTweets(data))
					for _, tweet := range tweets {
						log.Println(tweet.CreatedAt, tweet.FullText)
					}
				}()
			}
		//case *network.EventLoadingFinished:
		//	event.
		case *browser.EventDownloadProgress:
			log.Println("d2666c4", event.State)
		}
	})

	go func() {
		if err := chromedp.Run(ctx, chromedp.Navigate("https://x.com/knaka")); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	case res := <-ch:
		V0(fmt.Fprintln(os.Stdout, res))
	}

	return nil
}

func main() {
	verbose := flag.Bool("v", false, "verbose")
	headless := flag.Bool("h", false, "headless")
	timeout := flag.Duration("timeout", 10*time.Minute, "timeout")
	if err := Main(*timeout, *headless, *verbose); err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}
}
