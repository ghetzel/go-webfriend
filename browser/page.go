package browser

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

type Page struct {
	playwright.Page
	browser *Browser
}

func NewPage(browser *Browser) (*Page, error) {
	if browser.browser == nil {
		return nil, fmt.Errorf("missing playwright browser")
	}

	var page = &Page{
		browser: browser,
	}

	if pg, err := page.browser.browser.NewPage(); err == nil {
		page.Page = pg
		return page, nil
	} else {
		return nil, err
	}
}
