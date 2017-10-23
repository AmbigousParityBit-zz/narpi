// notaregularpixel package
package narpi

import (
	"image"
	"log"
	"os"
	"time"
)

const FileExt = ".narp"

func init() {
	log.SetFlags(log.Lshortfile)
	log.SetPrefix("(NARPImage):: ")
	log.SetOutput(os.Stderr)
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func (narp NotARegularPixel) drawNARP(img *image.RGBA, narpx int, narpy int) {
	s := narp.RunesArray()

	for j := 0; j < len(s[0]); j++ {
		for i := 0; i < len(s); i++ {
			if s[i][j] == 'o' {
				firstb := (narpy+j)*img.Stride + (narpx+i)*4
				img.Pix[firstb] = narp.Color.R
				img.Pix[firstb+1] = narp.Color.G
				img.Pix[firstb+2] = narp.Color.B
				img.Pix[firstb+3] = 255
			}
		}
	}
}

func (narp NotARegularPixel) markVisited(narpx int, narpy int, visited *[][]bool, vislenX, vislenY int) {
	if len(*visited) == 0 || len(*visited) != vislenX || len((*visited)[0]) != vislenY {
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

func showProgress(curr, max int, show bool) {
	if !show {
		return
	}

	progress := float32(curr) / float32(max) * 100.0
	if int(progress)%13 == 0 {
		log.Printf("Progress: %.2f%% \r", progress)
	}
}

func (narpimage *NARPImage) putToNarpImage(img *image.RGBA, showprogress bool) error {
	if img == nil {
		log.Panicln("putToNarpImage: Underlying image to construct from is nil")
	}

	var visited [][]bool

	narpimage.Size.X, narpimage.Size.Y = uint16(img.Bounds().Max.X), uint16(img.Bounds().Max.Y)
	boundsmin := struct{ X, Y uint16 }{uint16(img.Bounds().Min.X), uint16(img.Bounds().Min.Y)}
	sy := int(narpimage.Size.Y)
	sx := int(narpimage.Size.X)

	for y := int(boundsmin.Y); y < sy; y++ {
		if showprogress {
			showProgress(y, int(narpimage.Size.Y)-1, showprogress)
		}
		for x := int(boundsmin.X); x < sx; x++ {
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

func getNARP(x int, y int, img *image.RGBA, visited *[][]bool) (narp *NotARegularPixel) {
	firstb := y*img.Stride + x*4
	r := img.Pix[firstb]
	g := img.Pix[firstb+1]
	b := img.Pix[firstb+2]

	narp = &NotARegularPixel{
		HSize: 0, VSize: map[uint8][]uint8{}, Color: RGB8{r, g, b}}
	hsize := -1
	maxx := img.Rect.Max.X
	l := len(*visited)

	for xH := x; xH < maxx && colorsEqual(img, xH, y, narp.Color) && hsize < 253; xH++ {
		if l == 0 || !((*visited)[xH][y]) {
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

func getVerticalFloodCount(x int, y int, img *image.RGBA, visited *[][]bool) (verticals *[]uint8) {
	firstb := y*img.Stride + x*4
	r := img.Pix[firstb]
	g := img.Pix[firstb+1]
	b := img.Pix[firstb+2]

	color := RGB8{r, g, b}
	vsize := uint16(0)
	maxy := img.Bounds().Max.Y

	for yV := y + 1; yV < maxy && colorsEqual(img, x, yV, color); yV++ {
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
