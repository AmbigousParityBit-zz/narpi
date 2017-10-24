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

func (narp NotARegularPixel) markVisited(narpx int, narpy int, visited *[][]bool, vislenX, vislenY int, lenvis *int) {
	if *lenvis == 0 || *lenvis != vislenX || len((*visited)[0]) != vislenY {
		initVisitedArray(visited, vislenX, vislenY)
		*lenvis = len(*visited)
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

func initVisitedArray(visited *[][]bool, lenX, lenY int) {
	*visited = make([][]bool, lenX)
	for k := range *visited {
		(*visited)[k] = make([]bool, lenY)
	}
}

func (narpimage *NARPImage) putToNarpImage(img *image.RGBA) error {
	defer timeTrack(time.Now(), "putToNarpImage")
	if img == nil {
		log.Panicln("putToNarpImage: Underlying image to construct from is nil")
	}

	var visited [][]bool

	narpimage.Size.X, narpimage.Size.Y = uint16(img.Bounds().Max.X), uint16(img.Bounds().Max.Y)
	boundsmin := struct{ X, Y uint16 }{uint16(img.Bounds().Min.X), uint16(img.Bounds().Min.Y)}
	sy := int(narpimage.Size.Y)
	sx := int(narpimage.Size.X)
	lenvis := len(visited)

	counter := 0
	if narpimage.NARPixels == nil || len(narpimage.NARPixels) == 0 {
		narpimage.NARPixels = make([]NotARegularPixel, sx*sy)
	}

	log.Println()
	ss := time.Now()
	for y := int(boundsmin.Y); y < sy; y++ {
		for x := int(boundsmin.X); x < sx; x++ {
			if lenvis == 0 || !(visited[x][y]) {
				narp := getNARP(x, y, img, &visited, lenvis) //, &narpimage.Colors)
				narp.markVisited(int(x), int(y), &visited, sx, sy, &lenvis)

				//narpimage.NARPixels = append(narpimage.NARPixels, *narp)
				narpimage.NARPixels[counter] = *narp
				counter++
			}
		}
	}
	log.Println("putToNarpImage(inner loop)::", time.Since(ss))
	log.Println()

	narpimage.NARPixels = narpimage.NARPixels[:counter]
	//log.Println("Number of colors: ", len(narpimage.Colors))

	return nil
}

//func getNARP(x int, y int, img *image.RGBA, visited *[][]bool, lenvis int, colors *map[RGB8]rune) (narp *NotARegularPixel) {
func getNARP(x int, y int, img *image.RGBA, visited *[][]bool, lenvis int) (narp *NotARegularPixel) {
	firstb := y*img.Stride + x*4
	r := img.Pix[firstb]
	g := img.Pix[firstb+1]
	b := img.Pix[firstb+2]

	narp = &NotARegularPixel{
		HSize: 0, VSize: map[uint8]uint8{}, Color: RGB8{r, g, b}}
	//(*colors)[narp.Color] = '.'
	hsize := -1
	maxx := img.Rect.Max.X

	for xH := x; xH < maxx && colorsEqual(img, xH, y, narp.Color) && hsize < 253; xH++ {
		if lenvis == 0 || !((*visited)[xH][y]) {
			vsize := getVerticalFloodCount(xH, y, img, visited)
			if vsize != 0 {
				hsize++
				narp.VSize[uint8(hsize)] = vsize
			}
		}
	}
	if hsize == -1 {
		hsize++
	}
	narp.HSize = uint8(hsize)
	/*
		if narp.HSize > 5 {
			b := narp.Bytes()
			s := fmt.Sprintf(" len=%v   :::   %x  ", b.Len(), b)
			log.Println()
			log.Println(s)
			log.Println()
			narp1 := NotARegularPixel{}
			log.Println(narp1)
			narp1.ReadBytes(b)
			log.Println(narp1)
			log.Println()
		}
	*/
	return narp
}

func putBytesToUint16(l, r uint8) (v uint16) {
	v = uint16(l)
	v = v << 8
	v = v + uint16(r)

	return v
}

func cutBytesOfUint16(v uint16) (l uint8, r uint8) {
	l, r = uint8(v>>8), uint8(v&0xff)

	return l, r
}

func getVerticalFloodCount(x int, y int, img *image.RGBA, visited *[][]bool) (vsize uint8) {
	firstb := y*img.Stride + x*4
	r := img.Pix[firstb]
	g := img.Pix[firstb+1]
	b := img.Pix[firstb+2]

	color := RGB8{r, g, b}
	vsize = uint8(0)
	maxy := img.Bounds().Max.Y
	lenvis := len(*visited)

	for yV := y + 1; yV < 255 && yV < maxy && colorsEqual(img, x, yV, color); yV++ {
		if lenvis == 0 || !((*visited)[x][yV]) {
			vsize++
		}
	}

	return vsize
}
