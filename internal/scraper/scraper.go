package scraper

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	baseURL                    = "https://speiseplan.schwarz/menu"
	waitSelector              = "app-category"
	detailSelector          = "app-product-details"
	nutritionContentSelector = ".eqWrap"
	defaultTimeout        = 120 * time.Second
)

type Service struct {
	chromePath string
}

type Option func(*Service)

func WithChromePath(path string) Option {
	return func(s *Service) {
		s.chromePath = path
	}
}

func NewService(opts ...Option) *Service {
	s := &Service{}
	for _, o := range opts {
		o(s)
	}
	return s
}

// FetchMenu fetches the menu for a location/restaurant/date and enriches each
// item with detail data by navigating to each product page.
// date format: "2026-05-07" — leave empty for the site's default day.
func (s *Service) FetchMenu(ctx context.Context, location, restaurant, date string) (*DayMenu, error) {
	loc := url.PathEscape(location)
	r := url.PathEscape(restaurant)
	d := url.PathEscape(date)

	listURL := fmt.Sprintf("%s/%s/%s", baseURL, loc, r)
	if d != "" {
		listURL = fmt.Sprintf("%s/date/%s", listURL, d)
	}

	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	if s.chromePath != "" {
		allocOpts = append(allocOpts, chromedp.ExecPath(s.chromePath))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, allocOpts...)
	defer cancelAlloc()

	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	timeoutCtx, cancelTimeout := context.WithTimeout(taskCtx, defaultTimeout)
	defer cancelTimeout()

	// --- load the list page ---
	var listHTML string
	if err := chromedp.Run(timeoutCtx,
		chromedp.Navigate(listURL),
		chromedp.WaitVisible(waitSelector, chromedp.ByQuery),
		chromedp.OuterHTML("html", &listHTML),
	); err != nil {
		return nil, fmt.Errorf("scraper: navigate list %s: %w", listURL, err)
	}

	items, err := parseHTML(listHTML)
	if err != nil {
		return nil, fmt.Errorf("scraper: parse list: %w", err)
	}

	// --- click each product, navigate to its detail page, go back ---
	for i := range items {
		// Select the nth button across all .product-wrapper elements.
		// nth-of-type doesn't work cross-category, so we use JS to index them.
		clickJS := fmt.Sprintf(
			`document.querySelectorAll('.product-wrapper button[appclickableareatarget]')[%d].click()`,
			i,
		)

		var detailHTML string
		err := chromedp.Run(timeoutCtx,
			// make sure we're on the list page
			chromedp.Navigate(listURL),
			chromedp.WaitVisible(waitSelector, chromedp.ByQuery),
			// click the nth product button via JS
			chromedp.Evaluate(clickJS, nil),
			// the Angular router navigates to the detail page
			chromedp.WaitVisible(detailSelector, chromedp.ByQuery),
			// wait for content to load
			chromedp.Sleep(500*time.Millisecond),
			// click the nutrition tab - try multiple selectors
			chromedp.Click("[id*='mat-tab-label-1']", chromedp.ByQuery),
			chromedp.Sleep(1*time.Second),
			chromedp.WaitVisible(nutritionContentSelector, chromedp.ByQuery),
			// now capture HTML - includes both tabs' content
			chromedp.OuterHTML("html", &detailHTML),
		)
		if err != nil {
			// detail fetch failed — not fatal, continue with next item
			continue
		}

		detail, err := parseDetailPage(detailHTML)
		if err != nil {
			continue
		}
		items[i].Detail = detail
	}

	return &DayMenu{
		Date:  d,
		Items: items,
	}, nil
}