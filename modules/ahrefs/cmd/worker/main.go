package main

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func main() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Flag("headless", false),
		chromedp.Flag("hide-scrollbars", false),
		chromedp.Flag("mute-audio", false),
	)

	allocCtx, allocCtxCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCtxCancel()

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(allocCtx, time.Second)
	defer timeoutCtxCancel()

	// also set up a custom logger
	taskCtx, taskCtxCancel := chromedp.NewContext(timeoutCtx, chromedp.WithLogf(log.Printf))
	defer taskCtxCancel()

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		log.Fatalf("ERROR: %+v\n", err)
	}

	err := chromedp.Run(taskCtx,
		chromedp.Navigate(`https://app.ahrefs.com/user/login`),
		chromedp.WaitVisible(`//h1[contains(text(),"Sign in to Ahrefs")]`),
		chromedp.SendKeys("input[name=email]", "bleclerc797@gmail.com"),
		chromedp.SendKeys("input[name=password]", "Elws@#9935"),
		chromedp.Click("button[type=submit]", chromedp.NodeVisible),

		chromedp.WaitVisible(`//h2[contains(text(),"Projects")]`),

		chromedp.ActionFunc(func(ctx context.Context) error {
			cookies, err := network.GetAllCookies().Do(ctx)
			if err != nil {
				return err
			}

			for i, cookie := range cookies {
				log.Printf("chrome cookie %d: %+v", i, cookie)
			}

			return nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
}
