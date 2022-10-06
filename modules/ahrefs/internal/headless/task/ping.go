package task

import (
	"math/rand"
	"time"

	"github.com/chromedp/chromedp"
)

func Ping() chromedp.Tasks {
	rand.Seed(time.Now().UnixNano())

	menuSelectors := []map[string]string{
		{
			"url":             "https://app.ahrefs.com/dashboard",
			"visibleSelector": `body#dashboard`,
		},
		{
			"url":             "https://app.ahrefs.com/site-explorer",
			"visibleSelector": `body#site-explorer`,
		},
		{
			"url":             "https://app.ahrefs.com/keywords-explorer",
			"visibleSelector": `body#keywords-explorer`,
		},
		{
			"url":             "https://app.ahrefs.com/site-audit",
			"visibleSelector": `body#site-audit`,
		},
		{
			"url":             "https://app.ahrefs.com/rank-tracker",
			"visibleSelector": `body#rank-tracker`,
		},
		{
			"url":             "https://app.ahrefs.com/content-explorer",
			"visibleSelector": `body#content-explorer`,
		},
		{
			"url":             "https://app.ahrefs.com/domain-comparison",
			"visibleSelector": `body#domain-comparison`,
		},
		{
			"url":             "https://app.ahrefs.com/batch-analysis",
			"visibleSelector": `body#batch-analysis`,
		},
		{
			"url":             "https://app.ahrefs.com/link-intersect",
			"visibleSelector": `body#link-intersect`,
		},
	}

	menuSelector := menuSelectors[rand.Intn(len(menuSelectors))]

	return chromedp.Tasks{
		chromedp.Navigate(menuSelector["url"]),
		chromedp.WaitVisible(menuSelector["visibleSelector"]),
	}
}
