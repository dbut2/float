// Package utils provides display formatting utilities shared across the application.
package utils

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// printer formats numbers with English locale thousand separators (e.g. 1,234.56).
var printer = message.NewPrinter(language.English)

// currencySymbols maps ISO 4217 codes to their conventional display symbols.
var currencySymbols = map[string]string{
	"AUD": "$",
	"USD": "US$",
	"EUR": "€",
	"GBP": "£",
	"JPY": "¥",
	"CNY": "¥",
	"KRW": "₩",
	"SGD": "S$",
	"NZD": "NZ$",
	"HKD": "HK$",
	"CAD": "C$",
	"TWD": "NT$",
	"CHF": "Fr.",
	"SEK": "kr",
	"NOK": "kr",
	"DKK": "kr",
	"THB": "฿",
	"IDR": "Rp",
	"VND": "₫",
	"INR": "₹",
	"MYR": "RM",
	"PHP": "₱",
	"BRL": "R$",
	"MXN": "MX$",
	"ZAR": "R",
}

// ZeroDecimalCurrencies contains ISO 4217 codes for currencies with no minor units.
var ZeroDecimalCurrencies = map[string]bool{
	"JPY": true, "KRW": true, "IDR": true, "VND": true,
	"BIF": true, "GNF": true, "MGA": true, "PYG": true,
	"RWF": true, "UGX": true, "XAF": true, "XOF": true,
}

func symbolFor(code string) string {
	if s, ok := currencySymbols[code]; ok {
		return s
	}
	return code + "\u00a0" // non-breaking space fallback: "XXX 1,234.56"
}

// FormatAmount formats an amount in a currency's base units.
// For zero-decimal currencies (JPY, KRW, etc.) the base unit is the major unit.
// For all others, base units are cents (÷100).
// Negative values are prefixed with "-". Positive values have no sign prefix.
// e.g. FormatAmount(-148900, "AUD") → "-$1,489.00"
// e.g. FormatAmount(4987300, "JPY") → "¥4,987,300"
func FormatAmount(baseUnits int64, currencyCode string) string {
	return formatMajor(baseUnitToMajor(baseUnits, currencyCode), currencyCode, false)
}

// FormatSignedAmount is like FormatAmount but always includes the sign.
// Positive values are prefixed with "+". Negative values with "-".
// e.g. FormatSignedAmount(520000, "AUD") → "+$5,200.00"
// e.g. FormatSignedAmount(-8543, "AUD")  → "-$85.43"
func FormatSignedAmount(baseUnits int64, currencyCode string) string {
	return formatMajor(baseUnitToMajor(baseUnits, currencyCode), currencyCode, true)
}

// FormatForeignBalance converts an AUD amount to a foreign currency display string.
// rate is "1 AUD = rate foreignCurrencyCode".
// Negative values are prefixed with "-". Positive values have no sign prefix.
func FormatForeignBalance(audCents int64, rate float64, currencyCode string) string {
	return formatMajor(float64(audCents)/100.0*rate, currencyCode, false)
}

func baseUnitToMajor(baseUnits int64, currencyCode string) float64 {
	if ZeroDecimalCurrencies[currencyCode] {
		return float64(baseUnits)
	}
	return float64(baseUnits) / 100.0
}

func formatMajor(major float64, currencyCode string, includePositiveSign bool) string {
	sym := symbolFor(currencyCode)
	negative := major < 0
	if negative {
		major = -major
	}

	var number string
	if ZeroDecimalCurrencies[currencyCode] {
		number = printer.Sprintf("%.0f", major)
	} else {
		number = printer.Sprintf("%.2f", major)
	}

	switch {
	case negative:
		return "-" + sym + number
	case includePositiveSign:
		return "+" + sym + number
	default:
		return sym + number
	}
}
