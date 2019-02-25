package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
)

const (
	SquareSizePixels   = 60
	SquareBorderPixels = 3
	SquareFoodRadius   = SquareSizePixels / 4
)

// From github.com/fogleman/gg
func parseHexColor(x string) color.Color {
	var r, g, b, a uint8

	x = strings.TrimPrefix(x, "#")
	a = 255
	if len(x) == 3 {
		format := "%1x%1x%1x"
		fmt.Sscanf(x, format, &r, &g, &b)
		r |= r << 4
		g |= g << 4
		b |= b << 4
	}
	if len(x) == 6 {
		format := "%02x%02x%02x"
		fmt.Sscanf(x, format, &r, &g, &b)
	}
	if len(x) == 8 {
		format := "%02x%02x%02x%02x"
		fmt.Sscanf(x, format, &r, &g, &b, &a)
	}
	return color.RGBA{r, g, b, a}
}

func loadRawImageAsset(filename string) *image.Image {
	f, err := os.Open(fmt.Sprintf("render/assets/%s", filename))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	assetImage, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	return &assetImage
}

func loadImageAsset(filename string, w int, h int, rot int) *image.Image {
	srcImage := loadRawImageAsset(filename)

	var dstImage image.Image
	dstImage = imaging.Resize(*srcImage, w, h, imaging.Lanczos)

	if rot == 180 {
		dstImage = imaging.FlipH(dstImage)
	} else if rot == 90 {
		dstImage = imaging.Rotate90(dstImage)
	} else if rot == 270 {
		dstImage = imaging.Rotate270(dstImage)
	}

	return &dstImage
}

func drawWatermark(dc *gg.Context) {
	watermarkImage := loadRawImageAsset("watermark.png")
	dc.DrawImageAnchored(*watermarkImage, dc.Width()/2, dc.Height()/2, 0.5, 0.5)
}

func drawEmptySquare(dc *gg.Context, x int, y int) {
	dc.SetRGB255(240, 240, 240)
	dc.DrawRectangle(
		float64(x*SquareSizePixels+SquareBorderPixels),
		float64(y*SquareSizePixels+SquareBorderPixels),
		float64(SquareSizePixels-SquareBorderPixels*2),
		float64(SquareSizePixels-SquareBorderPixels*2),
	)
	dc.Fill()
}

func drawFood(dc *gg.Context, x int, y int) {
	dc.SetRGB255(255, 92, 117)
	dc.DrawCircle(
		float64(x*SquareSizePixels+SquareSizePixels/2),
		float64(y*SquareSizePixels+SquareSizePixels/2),
		SquareFoodRadius,
	)
	dc.Fill()
}

func drawSnakeImage(filename string, dc *gg.Context, x int, y int, hexColor string, direction string) {
	drawEmptySquare(dc, x, y)

	var rotation int
	switch direction {
	case "right":
		rotation = 0
	case "down":
		rotation = 270
	case "left":
		rotation = 180
	case "up":
		rotation = 90
	}

	maskImage := loadImageAsset(
		filename,
		SquareSizePixels-SquareBorderPixels*2,
		SquareSizePixels-SquareBorderPixels*2,
		rotation,
	)

	dst := dc.Image().(draw.Image)
	dstRect := image.Rect(
		x*SquareSizePixels+SquareBorderPixels,
		y*SquareSizePixels+SquareBorderPixels,
		(x+1)*SquareSizePixels-SquareBorderPixels,
		(y+1)*SquareSizePixels-SquareBorderPixels,
	)

	srcImage := &image.Uniform{parseHexColor(hexColor)}

	draw.DrawMask(dst, dstRect, srcImage, image.ZP, *maskImage, image.ZP, draw.Over)
}

func drawSnakeBody(dc *gg.Context, x int, y int, hexColor string) {
	dc.SetHexColor(hexColor)
	dc.DrawRectangle(
		float64(x*SquareSizePixels+SquareBorderPixels),
		float64(y*SquareSizePixels+SquareBorderPixels),
		float64(SquareSizePixels-SquareBorderPixels*2),
		float64(SquareSizePixels-SquareBorderPixels*2),
	)
	dc.Fill()
}

func drawBoard(b *Board) image.Image {
	dc := gg.NewContext(SquareSizePixels*b.Width, SquareSizePixels*b.Height)

	dc.SetRGB255(255, 255, 255)
	dc.Clear()

	// Draw empty squares under watermark
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			switch b.Squares[x][y].Content {
			case BoardSquareFood:
				drawEmptySquare(dc, x, y)
			case BoardSquareEmpty:
				drawEmptySquare(dc, x, y)
			}
		}
	}

	drawWatermark(dc)

	// Draw food and snakes over watermark
	var snakeAsset string
	for x, row := range b.Squares {
		for y, square := range row {
			switch square.Content {
			case BoardSquareSnakeHead:
				snakeAsset = fmt.Sprintf("heads/%s.png", square.SnakeType)
				drawSnakeImage(snakeAsset, dc, x, y, square.HexColor, square.Direction)
			case BoardSquareSnakeBody:
				drawSnakeBody(dc, x, y, square.HexColor)
			case BoardSquareSnakeTail:
				snakeAsset = fmt.Sprintf("tails/%s.png", square.SnakeType)
				drawSnakeImage(snakeAsset, dc, x, y, square.HexColor, square.Direction)
			case BoardSquareFood:
				drawFood(dc, x, y)
			}
		}
	}

	return dc.Image()
}
