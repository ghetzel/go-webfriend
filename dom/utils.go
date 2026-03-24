package dom

import (
	"github.com/playwright-community/playwright-go"
)

func FromPlaywright(locator playwright.Locator) (elements []*Element) {
	locator.WaitFor()

	return elementTreeFrom(locator, 0)
}

func elementTreeFrom(locator playwright.Locator, depth int) (elements []*Element) {
	if children, err := locator.All(); err == nil {
		for _, child := range children {
			var child = child.First()
			var element = new(Element)

			element.ID, _ = child.GetAttribute(`id`)

			// var x, _ = child.Evaluate(`this.attributes`, nil)
			// element.Name = typeutil.String(x)

			var grandchildren = child.Locator(`*:nth-child(1n)`)

			if cnt, err := grandchildren.Count(); err == nil && cnt > 0 {
				if gc := elementTreeFrom(grandchildren, depth+1); len(gc) > 0 {
					element.Children = gc
				}
			}

			element.Text, _ = child.InnerText()
			element.HTML, _ = child.InnerHTML()

			if element.Text != `` || len(element.Children) > 0 {
				elements = append(elements, element)
			}
		}
	}

	return
}
