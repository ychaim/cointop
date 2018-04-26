package cointop

import (
	"fmt"
	"strings"
	"time"

	"github.com/miguelmota/cointop/pkg/color"
	"github.com/miguelmota/cointop/pkg/fcache"
	"github.com/miguelmota/cointop/pkg/termui"
)

var oneWeek = (time.Hour * 24) * 7

func (ct *Cointop) updateChart() error {
	maxX := ct.maxtablewidth - 3
	coin := ct.selectedCoinName()
	ct.chartPoints(maxX, coin)
	if len(ct.chartpoints) != 0 {
		ct.chartview.Clear()
	}
	var body string
	for i := range ct.chartpoints {
		var s string
		for j := range ct.chartpoints[i] {
			p := ct.chartpoints[i][j]
			s = fmt.Sprintf("%s%c", s, p.Ch)
		}
		body = fmt.Sprintf("%s%s\n", body, s)

	}
	ct.update(func() {
		fmt.Fprint(ct.chartview, color.White(body))
	})

	return nil
}

func (ct *Cointop) chartPoints(maxX int, coin string) error {
	chart := termui.NewLineChart()
	chart.Height = 10
	chart.AxesColor = termui.ColorWhite
	chart.LineColor = termui.ColorCyan
	chart.Border = false

	now := time.Now()
	secs := now.Unix()
	start := secs - int64(oneWeek.Seconds())
	end := secs

	var data []float64
	cachekey := strings.ToLower(coin)
	if cachekey == "" {
		cachekey = "globaldata"
	}

	cached, found := ct.cache.Get(cachekey)
	if found {
		// cache hit
		data, _ = cached.([]float64)
		ct.debuglog("soft cache hit")
	}

	if len(data) == 0 {
		if coin == "" {
			graphData, err := ct.api.GetGlobalMarketGraphData(start, end)
			if err != nil {
				return nil
			}
			for i := range graphData.MarketCapByAvailableSupply {
				data = append(data, graphData.MarketCapByAvailableSupply[i][1]/1E9)
			}
		} else {
			graphData, err := ct.api.GetCoinGraphData(coin, start, end)
			if err != nil {
				return nil
			}
			for i := range graphData.PriceUSD {
				data = append(data, graphData.PriceUSD[i][1])
			}
		}

		ct.cache.Set(cachekey, data, 10*time.Second)
		go func() {
			_ = fcache.Set(cachekey, data, 24*time.Hour)
		}()
	}

	chart.Data = data
	termui.Body = termui.NewGrid()
	termui.Body.Width = maxX
	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(12, 0, chart),
		),
	)

	var points [][]termui.Cell
	// calculate layout
	termui.Body.Align()
	w := termui.Body.Width
	h := 10
	row := termui.Body.Rows[0]
	b := row.Buffer()
	for i := 0; i < h; i = i + 1 {
		var rowpoints []termui.Cell
		for j := 0; j < w; j = j + 1 {
			p := b.At(j, i)
			rowpoints = append(rowpoints, p)
		}
		points = append(points, rowpoints)
	}

	ct.chartpoints = points
	return nil
}

func (ct *Cointop) selectedCoinName() string {
	coin := ct.selectedcoin
	if coin != nil {
		return coin.Name
	}

	return ""
}

func (ct *Cointop) toggleCoinChart() error {
	highlightedcoin := ct.highlightedRowCoin()
	if ct.selectedcoin == highlightedcoin {
		ct.selectedcoin = nil
	} else {
		ct.selectedcoin = highlightedcoin
	}
	ct.updateChart()
	ct.updateMarketbar()
	return nil
}
