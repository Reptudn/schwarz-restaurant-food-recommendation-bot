package scraper

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseHTML(html string) ([]MenuItem, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var items []MenuItem

	doc.Find("app-category").Each(func(_ int, cat *goquery.Selection) {
		category := strings.TrimSpace(cat.Find("h3.category-header").Text())

		cat.Find(".product-wrapper").Each(func(_ int, prod *goquery.Selection) {
			name := strings.TrimSpace(prod.Find("span.pre-wrap").Text())
			if name == "" {
				return
			}

			var prices []Price
			prod.Find(".price").Each(func(_ int, p *goquery.Selection) {
				parts := strings.Fields(strings.TrimSpace(p.Text()))
				if len(parts) >= 2 {
					prices = append(prices, Price{
						Label: parts[0],
						Value: strings.Join(parts[1:], " "),
					})
				}
			})

			var tags []string
			prod.Find("app-product-custom-tag img").Each(func(_ int, img *goquery.Selection) {
				if alt, exists := img.Attr("alt"); exists && alt != "" {
					tags = append(tags, alt)
				}
			})

			items = append(items, MenuItem{
				Category: category,
				Name:     name,
				Prices:   prices,
				Tags:     tags,
			})
		})
	})

	return items, nil
}

func parseDetailPage(html string) (*MenuItemDetail, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	detail := &MenuItemDetail{
		ProductName:  extractProductName(doc),
		Allergens:   extractAllergens(doc),
		SubAllergens: extractSubAllergens(doc),
		Additives:   extractAdditives(doc),
		Traces:     extractTraces(doc),
		Nutrition:  extractNutrition(doc),
		PDFLink:    extractPDFLink(doc),
	}

	return detail, nil
}

func extractProductName(doc *goquery.Document) string {
	// Try to get product name from title or live announcer
	name := strings.TrimSpace(doc.Find("#cdk-live-announcer-0").Text())
	if name != "" {
		// Format: "Pulled turkey burger with BBQ sauce and steakhouse fries | Restaurant Bad Wimpfen | Bad Wimpfen | Menu | Schwarz | Speiseplan"
		if idx := strings.Index(name, " |"); idx > 0 {
			return strings.TrimSpace(name[:idx])
		}
		return name
	}
	// Try meta title
	title := doc.Find("meta[property='og:title']").First().AttrOr("content", "")
	if title != "" {
		return strings.TrimSpace(title)
	}
	return ""
}

func extractAllergens(doc *goquery.Document) []Allergen {
	var allergens []Allergen

	doc.Find("app-product-details div.ng-star-inserted").Each(func(_ int, section *goquery.Selection) {
		header := strings.TrimSpace(section.Find("h3").Text())
		if strings.EqualFold(header, "allergens") || strings.EqualFold(header, "Allergene") {
			section.Find("mat-list-item").Each(func(_ int, item *goquery.Selection) {
				code := strings.TrimSpace(item.Find("[matlistitemavatar]").Text())
				name := strings.TrimSpace(item.Find(".mdc-list-item__primary-text span").Text())
				if code != "" && name != "" {
					allergens = append(allergens, Allergen{Code: code, Name: name})
				}
			})
		}
	})

	return allergens
}

func extractSubAllergens(doc *goquery.Document) []string {
	var subAllergens []string

	doc.Find("img.subAllergen").Each(func(_ int, img *goquery.Selection) {
		if alt, exists := img.Attr("alt"); exists && alt != "" {
			subAllergens = append(subAllergens, alt)
		}
	})

	return subAllergens
}

func extractAdditives(doc *goquery.Document) []Additive {
	var additives []Additive

	doc.Find("app-product-details div.ng-star-inserted").Each(func(_ int, section *goquery.Selection) {
		header := strings.TrimSpace(section.Find("h3").Text())
		if strings.EqualFold(header, "additives") || strings.EqualFold(header, "Zusatzstoffe") {
			section.Find("mat-list-item").Each(func(_ int, item *goquery.Selection) {
				code := strings.TrimSpace(item.Find("[matlistitemavatar]").Text())
				name := strings.TrimSpace(item.Find(".mdc-list-item__primary-text span").Text())
				if code != "" && name != "" {
					additives = append(additives, Additive{Code: code, Name: name})
				}
			})
		}
	})

	return additives
}

func extractTraces(doc *goquery.Document) []string {
	var traces []string

	doc.Find("img.subTraces").Each(func(_ int, img *goquery.Selection) {
		if alt, exists := img.Attr("alt"); exists && alt != "" {
			traces = append(traces, alt)
		}
	})

	return traces
}

func extractNutrition(doc *goquery.Document) *Nutrition {
	nutrition := &Nutrition{}

	// Find the nutrition table in .eqWrap containers
	doc.Find(".eqWrap .eq").Each(func(_ int, eq *goquery.Selection) {
		// Each eq container has headers in first td and values in second td
		eq.Find("table").Each(func(_ int, table *goquery.Selection) {
			table.Find("tr").Each(func(_ int, row *goquery.Selection) {
				header := strings.TrimSpace(row.Find("th").First().Text())
				value := strings.TrimSpace(row.Find("td").First().Text())

				switch strings.ToLower(header) {
				case "energy kj":
					nutrition.EnergyKj = value
				case "energy kcal":
					nutrition.EnergyKcal = value
				case "protein":
					nutrition.Protein = value
				case "fat":
					nutrition.Fat = value
				case "carbohydrate":
					nutrition.Carbohydrates = value
				}
			})
		})
	})

	// If the above didn't work, try alternative selector - direct td parsing
	if nutrition.EnergyKj == "" && nutrition.EnergyKcal == "" {
		doc.Find(".eqWrap table td").Each(func(_ int, td *goquery.Selection) {
			text := strings.TrimSpace(td.Text())
			prev := strings.TrimSpace(td.Prev().Text())

			switch strings.ToLower(prev) {
			case "energy kj":
				nutrition.EnergyKj = text
			case "energy kcal":
				nutrition.EnergyKcal = text
			case "protein":
				nutrition.Protein = text
			case "fat":
				nutrition.Fat = text
			case "carbohydrate":
				nutrition.Carbohydrates = text
			}
		})
	}

	// Return nil if no nutrition data found
	if nutrition.EnergyKj == "" && nutrition.EnergyKcal == "" &&
		nutrition.Protein == "" && nutrition.Fat == "" && nutrition.Carbohydrates == "" {
		return nil
	}

	return nutrition
}

func extractPDFLink(doc *goquery.Document) string {
	var pdfLink string

	doc.Find("app-print a.customListItem").Each(func(_ int, a *goquery.Selection) {
		if href, exists := a.Attr("href"); exists && href != "" {
			pdfLink = href
		}
	})

	return pdfLink
}