package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

type Pokemon struct {
	FullName   string `json:"full_name"`
	Name       string `json:"name"`
	Number     string `json:"number"`
	Type       string `json:"type"`
	Total      string `json:"total"`
	HP         string `json:"hp"`
	Attack     string `json:"attack"`
	Defense    string `json:"defense"`
	SpAtk      string `json:"sp_atk"`
	SpDef      string `json:"sp_def"`
	Speed      string `json:"speed"`
	DetailPath string `json:"detail_path"`
	BaseExp    string `json:"base_exp"`
}

func main() {
	// Create a new collector
	c := colly.NewCollector(
		colly.AllowedDomains("pokemondb.net"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	// Collector for detail pages
	detailCollector := c.Clone()

	// Add delay between requests to simulate human-like browsing
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		RandomDelay: 2 * time.Second,
	})

	var (
		pokemons []Pokemon
		mutex    sync.Mutex
	)

	// Scrape Pokémon details
	c.OnHTML("table#pokedex tr", func(e *colly.HTMLElement) {
		// Skip header row
		if e.ChildText("th") != "" {
			return
		}

		number := e.ChildText("td:nth-child(1)")

		// Dừng lại sau khi đạt đến Pokémon có ID là 20
		if number > "0003" {
			return
		}

		// Lấy tên đầy đủ, bao gồm cả dạng Mega nếu có
		name := e.ChildText("td:nth-child(2) a")
		megaForm := e.ChildText("td:nth-child(2) small")
		fullName := name
		if megaForm != "" {
			fullName = megaForm
		}

		pokemon := Pokemon{
			FullName: fullName,
			Name:     e.ChildText("td:nth-child(2) a"),
			Number:   e.ChildText("td:nth-child(1)"),
			Type:     e.ChildText("td:nth-child(3)"),
			Total:    e.ChildText("td:nth-child(4)"),

			HP:         e.ChildText("td:nth-child(5)"),
			Attack:     e.ChildText("td:nth-child(6)"),
			Defense:    e.ChildText("td:nth-child(7)"),
			SpAtk:      e.ChildText("td:nth-child(8)"),
			SpDef:      e.ChildText("td:nth-child(9)"),
			Speed:      e.ChildText("td:nth-child(10)"),
			DetailPath: e.Request.AbsoluteURL(e.ChildAttr("td:nth-child(2) a", "href")),
		}

		// Validate data to ensure it's not empty
		if pokemon.Name == "" || pokemon.Number == "" {
			log.Printf("Incomplete data for Pokémon: skipping")
			return
		}

		mutex.Lock()
		pokemons = append(pokemons, pokemon)
		mutex.Unlock()

		log.Printf("Scraped: %+v", pokemon)

		// Visit the detail page for each Pokémon
		detailCollector.Visit(pokemon.DetailPath)
	})

	// Scrape additional details from the detail page
	detailCollector.OnHTML("h2:contains('Training') + .vitals-table", func(e *colly.HTMLElement) {
		// Tìm hàng chứa "Base Exp." và lấy giá trị
		baseExp := e.ChildText("tr:contains('Base Exp.') td")

		// Cập nhật Base Exp. cho Pokémon hiện tại
		if baseExp != "" {
			mutex.Lock()
			for i := range pokemons {
				if pokemons[i].DetailPath == e.Request.URL.String() {
					pokemons[i].BaseExp = baseExp
					log.Printf("Base Exp. found for %s: %s", pokemons[i].FullName, baseExp)
					break
				}
			}
			mutex.Unlock()
		}
	})

	// Error handling
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error on URL %s: %v", r.Request.URL, err)
	})

	// Visit the Pokémon database page
	url := "https://pokemondb.net/pokedex/all"
	log.Printf("Visiting URL: %s", url)
	if err := c.Visit(url); err != nil {
		log.Fatalf("Failed to visit URL %s: %v", url, err)
	}

	// Save to JSON
	if len(pokemons) == 0 {
		log.Println("No data was scraped. Check website or selectors.")
		return
	}

	file, err := os.Create("pokemons.json")
	if err != nil {
		log.Fatalf("Could not create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(pokemons); err != nil {
		log.Fatalf("Could not write to file: %v", err)
	}

	fmt.Printf("Scraping completed. Scraped %d Pokémon. Data saved to pokemons.json\n", len(pokemons))
}
