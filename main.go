package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/cobra"
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

func fetchRates(url string) ([]Rate, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rates []Rate
	if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
		return nil, err
	}
	return rates, nil
}

func Convert(rates []Rate, from, to string, amount float64) (float64, error) {
	var fromSum, toSum float64
	var foundFrom, foundTo bool

	for _, rate := range rates {
		if rate.Code == from {
			fromSum = rate.CBPrice
			foundFrom = true
		}
		if rate.Code == to {
			toSum = rate.CBPrice
			foundTo = true
		}
	}

	if !foundFrom || !foundTo {
		return 0, fmt.Errorf("conversion rate from %s to %s not found", from, to)
	}

	conAmount := amount * (toSum * fromSum)
	return conAmount, nil
}

func ListRates(rates []Rate) {
	table := t.NewWriter(os.Stdout)
	headers := []string{"Currency", "Central Bank Price (CBPrice)", "Code", "Date"}
	table.SetHeader(headers)

	for _, rate := range rates {
		table.Append([]string{rate.Title, fmt.Sprintf("%.2f", rate.CBPrice), rate.Code, rate.Date})
	}

	table.SetBorder(true)
	table.SetCenterSeparator("|")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("-")

	table.SetHeaderColor(
		t.Colors{t.FgHiMagentaColor, t.Bold},
		t.Colors{t.FgHiMagentaColor, t.Bold},
		t.Colors{t.FgHiMagentaColor, t.Bold},
		t.Colors{t.FgHiMagentaColor, t.Bold},
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
	var ratesURL = "https://nbu.uz/en/exchange-rates/json"

	var rootCmd = &cobra.Command{
		Use:   "currency_converter",
		Short: "convert currency using exchange rates",
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all conversion rates",
		Run: func(cmd *cobra.Command, args []string) {
			rates, err := fetchRates(ratesURL)
			if err != nil {
				color.Red("Failed to fetch rates: %v", err)
				os.Exit(1)
			}
			ListRates(rates)
		},
	}

	var convertCmd = &cobra.Command{
		Use:   "convert [amount] [from_currency] [to_currency]",
		Args:  cobra.ExactArgs(3),
		Short: "Convert currency from one to another",
		Run: func(cmd *cobra.Command, args []string) {
			amountStr := args[0]
			fromCurrency := args[1]
			toCurrency := args[2]

			amount, err := strconv.ParseFloat(amountStr, 64)
			if err != nil {
				color.Red("Invalid amount: %v", err)
				os.Exit(1)
			}

			rates, err := fetchRates(ratesURL)
			if err != nil {
				color.Red("Failed to fetch rates: %v", err)
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
				t.Colors{t.FgHiMagentaColor, t.Bold},
				t.Colors{t.FgHiMagentaColor, t.Bold},
				t.Colors{t.FgHiMagentaColor, t.Bold},
				t.Colors{t.FgHiMagentaColor, t.Bold},
			)
			table.SetColumnColor(
				t.Colors{t.FgHiWhiteColor},
				t.Colors{t.FgHiBlueColor},
				t.Colors{t.FgHiGreenColor},
				t.Colors{t.FgHiCyanColor},
			)

			table.Render()

			color.Cyan("\nThank you for using the Currency Converter!\n")
		},
	}

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(convertCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
