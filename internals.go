// notaregularpixel package
package NARPImage

import (
	"image"
	"image/color"
	"log"
	"os"
)

const FileExt = ".narp"

func init() {
	log.SetFlags(log.Lshortfile)
	log.SetPrefix("(NARPImage):: ")
	log.SetOutput(os.Stderr)
}

func drawAndMark(img *image.RGBA, x, y uint16, color color.Color, visited *[][]bool) {
	img.Set(int(x), int(y), color)
	(*visited)[x][y] = true
}

func (narp NotARegularPixel) markVisited(narpx int, narpy int, visited *[][]bool, vislenX, vislenY int) {
	if len(*visited) == 0 || len(*visited) != vislenX || len((*visited)[0]) != vislenX {
		initVisitedArray(visited, vislenX, vislenY)
	}

	s := narp.RunesArray()
	for j := 0; j < len(s[0]); j++ {
		for i := 0; i < len(s); i++ {
			if s[i][j] == 'o' {
				(*visited)[narpx+i][narpy+j] = true
			}
		}
	}

}

func (narpimage *NARPImage) initNARPImage() {
	narpimage.NARPixels = []NotARegularPixel{}
	narpimage.Size = struct{ X, Y uint16 }{0, 0}
	narpimage.Version = "0.6"
}

func initVisitedArray(visited *[][]bool, lenX, lenY int) {
	*visited = make([][]bool, lenX)
	for k := range *visited {
		(*visited)[k] = make([]bool, lenY)
	}
}

func showProgress(curr, max uint16, show bool) {
	if !show {
		return
	}

	progress := float32(curr) / float32(max) * 100.0
	if int(progress)%13 == 0 {
		log.Printf("Progress: %.2f%% \r", progress)
	}
}

func (narpimage *NARPImage) putToNarpImage(img image.Image, showprogress bool) error {
	if img == nil {
		log.Panicln("putToNarpImage: Underlying image to construct from is nil")
	}

	var visited [][]bool

	narpimage.Size.X, narpimage.Size.Y = uint16(img.Bounds().Max.X), uint16(img.Bounds().Max.Y)
	boundsmin := struct{ X, Y uint16 }{uint16(img.Bounds().Min.X), uint16(img.Bounds().Min.Y)}

	for y := boundsmin.Y; y < narpimage.Size.Y; y++ {
		showProgress(y, narpimage.Size.Y-1, showprogress)
		for x := boundsmin.X; x < narpimage.Size.X; x++ {
			if len(visited) == 0 || !(visited[x][y]) {
				narp := getNARP(x, y, img, &visited)
				narp.markVisited(int(x), int(y), &visited, int(narpimage.Size.X), int(narpimage.Size.Y))
				narpimage.NARPixels = append(narpimage.NARPixels, *narp)
			}
		}
	}

	if showprogress {
		log.Println()
	}

	return nil
}

func getNARP(x uint16, y uint16, img image.Image, visited *[][]bool) (narp *NotARegularPixel) {
	r, g, b := getRGBA8(img.At(int(x), int(y)))
	narp = &NotARegularPixel{
		HSize: 0, VSize: map[uint8][]uint8{}, Color: RGB8{r, g, b}}
	hsize := -1

	for xH := x; colorsEqual(img.At(int(xH), int(y)), narp.Color) && xH < uint16(img.Bounds().Max.X) && hsize < 253; xH++ {
		if len(*visited) == 0 || !((*visited)[xH][y]) {
			verticals := getVerticalFloodCount(xH, y, img, visited)
			if verticals != nil {
				hsize++
				narp.VSize[uint8(hsize)] = append(narp.VSize[uint8(hsize)], *verticals...)
			}
		}
	}
	if hsize == -1 {
		hsize++
	}
	narp.HSize = uint8(hsize)

	return narp
}

func putBytesToUint16(lr []uint8) (v uint16) {
	if len(lr) == 0 {
		return 0
	}
	if len(lr) == 1 {
		v = uint16(lr[0])
	} else {
		v = uint16(lr[0])
		v = v << 8
		v = v + uint16(lr[1])
	}

	return v
}

func cutBytesOfUint16(v uint16) (b bool, left uint8, right uint8) {
	if v > 255 {
		left := uint8((v & 240) >> 4)
		right := uint8(v & 15)
		return true, left, right
	}
	return false, 0, 0
}

func getVerticalFloodCount(x uint16, y uint16, img image.Image, visited *[][]bool) (verticals *[]uint8) {
	r, g, b := getRGBA8(img.At(int(x), int(y)))
	color := RGB8{r, g, b}
	vsize := uint16(0)

	for yV := y + 1; yV < uint16(img.Bounds().Max.Y) && colorsEqual(img.At(int(x), int(yV)), color); yV++ {
		if len(*visited) == 0 || !((*visited)[x][yV]) {
			vsize++
		}
	}

	if vsize == 0 {
		return nil
	}

	verticals = &[]uint8{}
	cutOrNot, left, right := cutBytesOfUint16(vsize)
	if cutOrNot {
		*verticals = append(*verticals, left)
		*verticals = append(*verticals, right)
	} else {
		*verticals = append(*verticals, uint8(vsize))
	}
	return verticals
}
