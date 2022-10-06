package task

import (
	"context"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func Login(username, password string, cookies []*network.Cookie) chromedp.Tasks {
	var err error
	return chromedp.Tasks{
		chromedp.Navigate("https://app.ahrefs.com/user/login"),
		chromedp.WaitVisible(`//h1[contains(text(),"Sign in to Ahrefs")]`),
		chromedp.SendKeys("input[name=email]", username),
		chromedp.SendKeys("input[name=password]", password),
		chromedp.Click("button[type=submit]", chromedp.NodeVisible),

		chromedp.WaitVisible(`//h2[contains(text(),"Projects")]`),

		chromedp.ActionFunc(func(ctx context.Context) error {
			cookies, err = network.GetAllCookies().Do(ctx)
			return err
		}),
	}
}
