package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Reptudn/schwarz-restaurant-food-recommendation-bot/internal/scraper"
)

func main() {
	ctx := context.Background()

	svc := scraper.NewService(
		scraper.WithChromePath("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"),
	)

	menu, err := svc.FetchMenu(ctx, "Bad Wimpfen", "Restaurant Bad Wimpfen", "2026-05-07")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("🍽️  Speiseplan — %s\n", menu.Date)
	fmt.Println(strings.Repeat("═", 60))

	currentCategory := ""
	for _, item := range menu.Items {
		if item.Category != currentCategory {
			currentCategory = item.Category
			fmt.Printf("\n  📂 %s\n", strings.ToUpper(currentCategory))
			fmt.Println("  " + strings.Repeat("─", 56))
		}

		fmt.Printf("\n  🥘 %s\n", item.Name)

		for _, p := range item.Prices {
			fmt.Printf("     💶 %-10s %s\n", p.Label, p.Value)
		}

		if len(item.Tags) > 0 {
			fmt.Printf("     🏷️  %s\n", strings.Join(item.Tags, " · "))
		}

		if item.Detail != nil {
			// Product name (if different from menu name)
			if item.Detail.ProductName != "" && item.Detail.ProductName != item.Name {
				fmt.Printf("     📛 Name: %s\n", item.Detail.ProductName)
			}

			// Nutrition
			if item.Detail.Nutrition != nil {
				fmt.Printf("     📊 Nutrition (per 100g):\n")
				if item.Detail.Nutrition.EnergyKj != "" {
					fmt.Printf("        • Energy (kJ): %s\n", item.Detail.Nutrition.EnergyKj)
				}
				if item.Detail.Nutrition.EnergyKcal != "" {
					fmt.Printf("        • Energy (kcal): %s\n", item.Detail.Nutrition.EnergyKcal)
				}
				if item.Detail.Nutrition.Protein != "" {
					fmt.Printf("        • Protein: %s\n", item.Detail.Nutrition.Protein)
				}
				if item.Detail.Nutrition.Fat != "" {
					fmt.Printf("        • Fat: %s\n", item.Detail.Nutrition.Fat)
				}
				if item.Detail.Nutrition.Carbohydrates != "" {
					fmt.Printf("        • Carbohydrates: %s\n", item.Detail.Nutrition.Carbohydrates)
				}
			} else {
				fmt.Printf("     ⚠️  No nutrition information available.\n")
			}

			// Allergens
			if len(item.Detail.Allergens) > 0 {
				codes := make([]string, len(item.Detail.Allergens))
				for i, a := range item.Detail.Allergens {
					codes[i] = fmt.Sprintf("%s=%s", a.Code, a.Name)
				}
				fmt.Printf("     ⚠️  Allergens: %s\n", strings.Join(codes, ", "))
			}

			// Sub-allergens
			if len(item.Detail.SubAllergens) > 0 {
				fmt.Printf("     ⚠️  Sub-Allergens: %s\n", strings.Join(item.Detail.SubAllergens, ", "))
			}

			// Additives
			if len(item.Detail.Additives) > 0 {
				adds := make([]string, len(item.Detail.Additives))
				for i, a := range item.Detail.Additives {
					adds[i] = fmt.Sprintf("%s=%s", a.Code, a.Name)
				}
				fmt.Printf("     🧪 Additives: %s\n", strings.Join(adds, ", "))
			}

			// Traces
			if len(item.Detail.Traces) > 0 {
				fmt.Printf("     ⚠️  Traces: %s\n", strings.Join(item.Detail.Traces, ", "))
			}

			// PDF Link
			if item.Detail.PDFLink != "" {
				fmt.Printf("     📄 PDF: %s\n", item.Detail.PDFLink)
			}
		}
	}

	fmt.Printf("\n%s\n", strings.Repeat("═", 60))
	fmt.Printf("Total: %d items\n", len(menu.Items))
}