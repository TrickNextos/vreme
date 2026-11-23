package main

import (
	"context"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"sync"
	"time"

	"github.com/TrickNextos/vreme/vreme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "vreme",
		Usage: "Prikaz vremenskih postaj, meritev in poročila",
		Flags: []cli.Flag{
			&cli.FloatFlag{
				Name:  "minTemp",
				Usage: "Najnižja dovoljena temperatura meritve (°C)",
				Value: -50,
			},
			&cli.FloatFlag{
				Name:  "maxTemp",
				Usage: "Najvišja dovoljena temperatura meritve (°C)",
				Value: 50,
			},
			&cli.IntFlag{
				Name:  "stMeritev",
				Usage: "Najmanjše število meritev za izračun statistike",
				Value: 3,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runTUI(
				cmd.Float("minTemp"),
				cmd.Float("maxTemp"),
				int(cmd.Int("stMeritev")),
			)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func dummyTemperatura(minTemp, maxTemp float64) float64 {
	lo := math.Max(minTemp, -10.0)
	hi := math.Min(maxTemp, 35.0)
	if lo > hi {
		lo, hi = minTemp, maxTemp
	}
	val := lo + rand.Float64()*(hi-lo)
	return math.Round(val*10) / 10
}

func runTUI(minTemp, maxTemp float64, stMeritev int) error {
	var mu sync.Mutex
	order := []string{"LJ01", "PO01", "RG01"}

	postaje := make(map[string]vreme.Postaja)
	vreme.DodajPostajo(postaje, "LJ01", "Zelezniska postaja", "Ljubljana")
	vreme.DodajPostajo(postaje, "PO01", "Fancy hotel", "Portoroz")
	vreme.DodajPostajo(postaje, "RG01", "Glavna vlecnica", "Rogla")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := tview.NewApplication()

	cardLJ := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft)
	cardLJ.SetBorder(true).SetTitle("  LJ01 - Ljubljana  ")

	cardPO := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft)
	cardPO.SetBorder(true).SetTitle("  PO01 - Portoroz  ")

	cardRG := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft)
	cardRG.SetBorder(true).SetTitle("  RG01 - Rogla  ")

	reportView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetChangedFunc(func() {
			app.Draw()
		})
	reportView.SetBorder(true).SetTitle("  Poročilo in meritve  ")

	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[gray]Press [white]q[gray] to quit").
		SetTextAlign(tview.AlignLeft)

	cardsRow := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(cardLJ, 0, 1, false).
		AddItem(cardPO, 0, 1, false).
		AddItem(cardRG, 0, 1, false)

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(cardsRow, 6, 0, false).
		AddItem(reportView, 0, 1, true).
		AddItem(footer, 1, 0, false)

	refreshUI := func() {
		mu.Lock()
		singleLJ := map[string]vreme.Postaja{"LJ01": postaje["LJ01"]}
		singlePO := map[string]vreme.Postaja{"PO01": postaje["PO01"]}
		singleRG := map[string]vreme.Postaja{"RG01": postaje["RG01"]}
		mu.Unlock()

		app.QueueUpdateDraw(func() {
			cardLJ.SetText(vreme.IzpisMeritev(singleLJ))
			cardPO.SetText(vreme.IzpisMeritev(singlePO))
			cardRG.SetText(vreme.IzpisMeritev(singleRG))
			reportView.SetText(vreme.IzpisPorocila(postaje, stMeritev))
		})
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q', 'Q':
			cancel()
			app.Stop()
			return nil
		}
		return event
	})

	go func() {
		refreshUI()
		for {
			delay := time.Duration(1000+rand.IntN(2001)) * time.Millisecond
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
				id := order[rand.IntN(len(order))]
				novaTemp := dummyTemperatura(minTemp, maxTemp)

				mu.Lock()
				vreme.DodajMeritev(postaje, id, novaTemp, minTemp, maxTemp)
				mu.Unlock()

				refreshUI()
			}
		}
	}()

	return app.SetRoot(layout, true).Run()
}
