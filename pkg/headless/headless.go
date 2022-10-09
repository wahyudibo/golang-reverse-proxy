package headless

import (
	"context"
	"os"
	"strconv"

	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog/log"
)

type HeadlessBrowser struct {
	Context     context.Context
	CancelFuncs []context.CancelFunc
}

func New(ctx context.Context) (*HeadlessBrowser, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Flag("user-agent", os.Getenv("PROXY_USER_AGENT")),
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
	taskCtx, taskCtxCancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		return nil, err
	}

	return &HeadlessBrowser{
		Context: taskCtx,
		CancelFuncs: []context.CancelFunc{
			taskCtxCancel,
			allocCtxCancel,
		},
	}, nil
}

func (h *HeadlessBrowser) Close() {
	for _, cancel := range h.CancelFuncs {
		cancel()
	}
}
