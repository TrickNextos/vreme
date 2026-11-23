// Package vreme zagotavlja podatkovne strukture in funkcije za upravljanje
// vremenskih postaj, dodajanje meritev ter izpis statistike in poročil.
package vreme

import (
	"fmt"
	"strings"
)

// Postaja predstavlja posamezno vremensko postajo in njene zbrane meritve.
type Postaja struct {
	ime     string
	kraj    string
	meritve []float64
}

// DodajPostajo v slovar postaj doda novo postajo z podano oznako, imenom in krajem.
func DodajPostajo(postaje map[string]Postaja, oznaka, ime, kraj string) {
	postaje[oznaka] = Postaja{
		ime:  ime,
		kraj: kraj,
	}
}

// DodajMeritev doda novo temperaturno meritev k postaji z dano oznako,
// če je temperatura znotraj dovoljenega območja [minTemp, maxTemp].
func DodajMeritev(postaje map[string]Postaja, oznaka string, temperatura float64, minTemp, maxTemp float64) {
	if minTemp > temperatura || temperatura > maxTemp {
		fmt.Printf("Temperatura %.1f je izven obmocja [%.1f, %.1f]\n", temperatura, minTemp, maxTemp)
		return
	}

	p, ok := postaje[oznaka]
	if !ok {
		fmt.Println("Postaja z oznako", oznaka, "ne obstaja")
		return
	}

	p.meritve = append(p.meritve, temperatura)
	postaje[oznaka] = p
}

func statistika(postaje map[string]Postaja, oznaka string, stMeritev int) (najnizja, najvisja, povprecna float64, ok bool) {
	p, ok := postaje[oznaka]
	if !ok {
		return
	}

	if len(p.meritve) < stMeritev {
		ok = false
		return
	}

	najvisja = p.meritve[0]
	najnizja = p.meritve[0]
	sum := 0.0

	for _, meritev := range p.meritve {
		najvisja = max(meritev, najvisja)
		najnizja = min(meritev, najnizja)
		sum += meritev
	}
	povprecna = sum / float64(len(p.meritve))
	return
}

// IzpisMeritev vrne oblikovan niz z vsemi meritvami za posamezne postaje.
func IzpisMeritev(postaje map[string]Postaja) string {
	var b strings.Builder
	b.WriteString("MERITVE:\n")
	for id, p := range postaje {
		fmt.Fprintf(&b, "%s - %s (%s) %v\n", id, p.ime, p.kraj, p.meritve)
	}
	return b.String()
}

// IzpisPorocila vrne oblikovano poročilo s statistiko temperatur za vse postaje.
func IzpisPorocila(postaje map[string]Postaja, stMeritev int) string {
	var b strings.Builder
	b.WriteString("POROCILO:\n")
	for id, p := range postaje {
		minV, maxV, avgV, ok := statistika(postaje, id, stMeritev)
		fmt.Fprintf(&b, "%s: ", p.ime)
		if !ok {
			b.WriteString("premalo meritev\n")
			continue
		}

		var temp string
		switch {
		case 25 <= avgV:
			temp = "Vroče"
		case 15 <= avgV && avgV < 25:
			temp = "Toplo"
		case 5 <= avgV && avgV < 15:
			temp = "Hladno"
		case avgV < 5:
			temp = "Mrzlo"
		default:
			temp = ""
		}
		fmt.Fprintf(&b, "min %.1f °C, max %.1f °C, povprečje %.1f °C -> %s\n", minV, maxV, avgV, temp)
	}
	return b.String()
}
