package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/fatih/color"
	t "github.com/olekukonko/tablewriter"
)

type Rate struct {
	Title        string  `json:"title"`
	Code         string  `json:"code"`
	CBPrice      float64 `json:"cb_price,string"`
	NBUBuyPrice  *string `json:"nbu_buy_price"`
	NBUCellPrice *string `json:"nbu_cell_price"`
	Date         string  `json:"date"`
}

func loadRates(filename string) ([]Rate, error) {
	var rates []Rate

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &rates)
	return rates, err
}

func Convert(rates []Rate, from, to string, amount float64) (float64, error) {
	var fromRate, toRate float64
	var foundFrom, foundTo bool

	for _, rate := range rates {
		if rate.Code == from {
			fromRate = rate.CBPrice
			foundFrom = true
		}
		if rate.Code == to {
			toRate = rate.CBPrice
			foundTo = true
		}
	}

	if !foundFrom || !foundTo {
		return 0, fmt.Errorf("conversion rate from %s to %s not found", from, to)
	}

	convertedAmount := amount * (toRate / fromRate)
	return convertedAmount, nil
}
func ListRates(rates []Rate) {
	table := t.NewWriter(os.Stdout)
	headers := []string{"Currency", "Central Bank Price (CBPrice)", "Date", "Code"}
	table.SetHeader(headers)

	for _, rate := range rates {
		table.Append([]string{rate.Title, fmt.Sprintf("%.2f", rate.CBPrice), rate.Date, rate.Code})
	}

	table.SetBorder(true)
	table.SetCenterSeparator("|")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("-")

	table.SetHeaderColor(
		t.Colors{t.FgHiWhiteColor, t.Bold},
		t.Colors{t.FgHiCyanColor, t.Bold},
		t.Colors{t.FgHiGreenColor, t.Bold},
		t.Colors{t.FgHiBlueColor, t.Bold},
	)

	table.SetColumnColor(
		t.Colors{t.FgHiWhiteColor},
		t.Colors{t.FgCyanColor},
		t.Colors{t.FgHiGreenColor},
		t.Colors{t.FgHiBlueColor},
	)

	table.Render()
}

func main() {
	ratesFile := "rates.json"
	rates, err := loadRates(ratesFile)
	if err != nil {
		color.Red("Failed to load rates: %v", err)
		os.Exit(1)
	}

	listRates := flag.Bool("list", false, "List all available conversion rates")
	flag.Parse()

	if *listRates {
		ListRates(rates)
		return
	}

	args := flag.Args()
	if len(args) != 3 {
		color.Red("Usage: ./currency_converter <amount> <from_currency> <to_currency>")
		os.Exit(1)
	}

	amountStr := args[0]
	fromCurrency := args[1]
	toCurrency := args[2]

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		color.Red("Invalid amount: %v", err)
		os.Exit(1)
	}

	result, err := Convert(rates, fromCurrency, toCurrency, amount)
	if err != nil {
		color.Red("Conversion failed: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Conversion result: %.2f %s to %s\n", amount, fromCurrency, toCurrency)

	table := t.NewWriter(os.Stdout)
	table.SetHeader([]string{"Amount", "From Currency", "To Currency", "Converted Amount"})

	table.Append([]string{
		fmt.Sprintf("%.2f", amount),
		fromCurrency,
		toCurrency,
		fmt.Sprintf("%.2f", result),
	})

	table.SetBorder(true)
	table.SetCenterSeparator("|")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("-")
	table.SetHeaderColor(
		t.Colors{t.FgHiWhiteColor, t.Bold},
		t.Colors{t.FgHiCyanColor, t.Bold},
		t.Colors{t.FgHiCyanColor, t.Bold},
		t.Colors{t.FgHiGreenColor, t.Bold},
	)
	table.SetColumnColor(
		t.Colors{t.FgHiYellowColor},
		t.Colors{t.FgHiMagentaColor},
		t.Colors{t.FgHiBlueColor},
		t.Colors{t.FgHiRedColor},
	)

	table.Render()

	color.Cyan("\nThank you for using the Currency Converter!\n")
}
