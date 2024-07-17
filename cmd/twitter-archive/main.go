package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
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
	//ctx, cancel := chromedp.NewExecAllocator(context.Background(),
	//	chromedp.Flag("headless", headless),
	//	chromedp.ExecPath("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"),
	//	chromedp.UserDataDir(""),
	//	chromedp.Flag("new-window", true),
	//)
	ctx, cancel := chromedp.NewRemoteAllocator(context.Background(), "ws://127.0.0.1:9222")
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx, opts...)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	ch, errch := make(chan string, 1), make(chan error, 1)
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
		case *network.EventDataReceived:
			if url, ok := urls[event.RequestID]; ok {
				log.Println("efa761a", url)
			}
		case *browser.EventDownloadProgress:
			log.Println("d2666c4", event.State)
		}
	})

	go func() {
		if err := chromedp.Run(ctx, chromedp.Navigate("https://x.com/knaka")); err != nil {
			errch <- err
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errch:
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
