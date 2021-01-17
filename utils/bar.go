package utils

import (
	"fmt"
	"github.com/dustin/go-humanize"
)

type progressBar struct {
	percent int64  // progress
	cur     int64  // current
	total   int64  // total value for progress
	rate    string //
	symbol  string // the fill value for progress bar
}

func (bar *progressBar) getPercent() int64 {
	return int64(float32(bar.cur) / float32(bar.total) * 100)
}

func (bar *progressBar) setSymbol(symbol string) {
	bar.symbol = symbol
}

func NewBar(start, total int64) *progressBar {
	bar := &progressBar{}
	bar.cur = start
	bar.total = total
	if bar.symbol == "" {
		bar.symbol = "█"
	}
	bar.percent = bar.getPercent()
	for i := 0; i < int(bar.percent); i += 2 {
		bar.rate += bar.symbol // initial progress position
	}
	return bar
}

func NewCustomBar(start, total int64, symbol string) *progressBar {
	bar := NewBar(start, total)
	bar.setSymbol(symbol)
	return bar
}

func (bar *progressBar) Show(cur int64) {
	bar.cur = cur
	last := bar.percent
	bar.percent = bar.getPercent()
	if bar.percent != last && bar.percent%2 == 0 {
		bar.rate += bar.symbol
	}
	fmt.Printf("\r[%-50s]%3d%% %8v/%v", bar.rate, bar.percent,
		humanize.Bytes(uint64(bar.cur)), humanize.Bytes(uint64(bar.total)))
}
