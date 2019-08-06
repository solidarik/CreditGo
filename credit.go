package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/leekchan/accounting"
)

const percRate = 9.75 / 12 / 100
const dtFormat = "02.01.2006"

var p = fmt.Println
var f = fmt.Printf

// Pay Структура внесенных платежей
type Pay struct {
	DateStr string `json:"date"`
	Amount  int    `json:"amount"`
}

// PayWithDate Структуру платежей с датой
type PayWithDate struct {
	Pay
	Date time.Time
}

// Data Массив внешних данных
type Data struct {
	Pays         []Pay `json:"pays"`
	PaysWithDate []PayWithDate
	Credit       int    `json:"credit"`
	StartDateStr string `json:"start_date"`
	Years        int    `json:"years"`
	PercRateStr  string `json:"perc_rate"`
	PercRate     float32
}

// Strategy Стратегия погашения кредита
type Strategy struct {
	Name    string
	OverPay int
	Benefit int
}

func calcCurrentMonthPay(credit int, months int) int {
	var mpPlus1 = math.Pow(percRate+1, float64(months))
	var monthPay = float64(credit) * percRate * mpPlus1 / (mpPlus1 - 1)
	return int(math.Round(monthPay))
}

func calcOverPay(monthPay int, months int) int {
	return int(monthPay * months)
}

func readDataFromFile(filename string) Data {
	var file, _ = ioutil.ReadFile(filename)
	var data Data
	var pays []PayWithDate
	json.Unmarshal(file, &data)
	for i := 0; i < len(data.Pays); i++ {
		var dateParse, _ = time.Parse(dtFormat, data.Pays[i].DateStr)
		var payWithDate = PayWithDate{Pay: data.Pays[i], Date: dateParse}
		pays = append(pays, payWithDate)
	}
	data.PaysWithDate = pays
	return data
}

func printCaption(s string) {
	f("\n")
	var str = ">>>>>>>>>>>>>>> " + strings.ToUpper(s) + " <<"
	var size = utf8.RuneCountInString(str)
	for i := size; i < 80; i++ {
		str += "<"
	}
	f("%s\n", str)
}

func calcStupidStrategy(credit int, months int) Strategy {
	var monthPay = calcCurrentMonthPay(credit, months)
	var overPay = calcOverPay(monthPay, months)

	return Strategy{Name: "Платим каждый месяц без доп. взносов", OverPay: overPay}
}

func calcFiftyStrategy(credit int, months int) Strategy {
	// var accum = 0
	// var startDate = time.Parse(dtFormat, "17.07.2019")
	return Strategy{Name: "Платим как только накопили 50 тыс."}
}

func calcCurrentStrategy(credit int, months int) Strategy {
	return Strategy{Name: "Платим с учетом текущих досрочных платежей"}
}

func calcTimeStrategy(credit int, months int) Strategy {
	return Strategy{Name: "Платим на погашение (!) срока платежа"}
}

func main() {
	ac := accounting.Accounting{Symbol: "₽ ", Precision: 2}
	var strategies []Strategy

	var data = readDataFromFile("data.json")
	printCaption("Наши взносы")
	for i := 0; i < len(data.PaysWithDate); i++ {
		var t3 = data.PaysWithDate[i].Date
		f("%02d-%02d-%02d\t%s\n", t3.Year(), t3.Month(), t3.Day(), ac.FormatMoney(data.PaysWithDate[i].Amount))
	}

	var startCredit = data.Credit
	var months = 12 * data.Years

	strategies = append(strategies, calcStupidStrategy(startCredit, months))
	strategies = append(strategies, calcFiftyStrategy(startCredit, months))
	strategies = append(strategies, calcCurrentStrategy(startCredit, months))
	strategies = append(strategies, calcTimeStrategy(startCredit, months))
	for i := 0; i < len(strategies); i++ {
		var s = strategies[i]
		printCaption("Стратегия " + s.Name)
		f("Выгода: %s\n", ac.FormatMoney(s.Benefit))
		f("Переплата: %s\n", ac.FormatMoney(s.OverPay))
	}

}
