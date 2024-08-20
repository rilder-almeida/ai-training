package ai

import (
	"bytes"
	"strings"

	"github.com/fogleman/gg"
)

/*
|🟢|🟢|🟢|🟢|🟢|🟢|🟢|
|🟢|🟢|🟢|🟢|🟢|🟢|🟢|
|🟢|🟢|🟢|🟢|🟢|🟢|🟢|
|🟢|🟢|🟢|🔴|🟢|🟢|🟢|
|🟢|🟢|🟢|🔵|🟢|🟢|🟢|
|🔵|🟢|🟢|🔵|🔴|🟢|🟢|
*/

func generateImage(boardData string) ([]byte, error) {
	const rows = 6
	const cols = 7
	const width = 190
	const height = 165

	y := float64(20)
	gap := float64(25)

	dc := gg.NewContext(width, height)
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	markersByRow := strings.Split(boardData, "\n")
	for row := range rows {
		x := float64(20)

		markersByCol := strings.Split(markersByRow[row], "|")
		for col := 1; col <= cols; col++ {
			switch markersByCol[col] {
			case "🟢":
				dc.SetRGB(0, 1, 0)
			case "🔵":
				dc.SetRGB(0, 0, 1)
			case "🔴":
				dc.SetRGB(1, 0, 0)
			}

			dc.DrawCircle(x, y, 10)
			dc.Fill()

			x += gap
		}

		y += gap
	}

	var b bytes.Buffer
	if err := dc.EncodePNG(&b); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
