package headless

import (
	"context"
	"os"
	"strconv"

	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog/log"
)

func New(ctx context.Context) (context.Context, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
	)

	isHeadless, err := strconv.ParseBool(os.Getenv("HEADLESS_BROWSER_HEADLESS_MODE"))
	if err != nil {
		return nil, err
	}

	if !isHeadless {
		opts = append(opts,
			chromedp.Flag("headless", false),
			chromedp.Flag("hide-scrollbars", false),
			chromedp.Flag("mute-audio", false),
		)
	}

	allocCtx, allocCtxCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCtxCancel()

	taskCtx, taskCtxCancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer taskCtxCancel()

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		return nil, err
	}

	return taskCtx, nil
}
