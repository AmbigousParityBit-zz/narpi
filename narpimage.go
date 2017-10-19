package NARPImage

// notaregularpixel package

import (
	"bytes"
	"encoding/gob"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"os"
)

type Point struct {
	X, Y uint16
}

type RGBColor struct {
	R, G, B uint8
}

type NotARegularPixel struct {
	HSize uint8             // horizontal flood size
	Color RGBColor          // color of pixel
	VSize map[uint8][]uint8 // map has cells with count of vertical pixels of the same color
	//
	//	ex.: HSize=3, VSize[2]=[2,1,2] means area of the same color:
	//		X	X	X
	//			X
	//			X
	//			X
	x, y uint16
}

type NARPImage struct {
	NARPixels []NotARegularPixel
	Size      Point
	Version   string
}

func init() {
	log.SetFlags(log.Lshortfile)
	log.SetPrefix("(NARPImage):: ")
	log.SetOutput(os.Stderr)
}

func drawAndMark(img *image.RGBA, x, y uint16, color color.Color, visited *[][]bool) {
	img.Set(int(x), int(y), color)
	(*visited)[x][y] = true
}

func (narpimage *NARPImage) DeconstructToPngFile(s string) error {
	img := image.NewRGBA(image.Rect(0, 0, int(narpimage.Size.X), int(narpimage.Size.Y)))
	var visited [][]bool
	initVisitedArray(&visited, int(narpimage.Size.X), int(narpimage.Size.Y))
	x, y := uint16(0), uint16(0)

	for _, narpixel := range narpimage.NARPixels {
		//printNARPixel(&narpixel, 5)
		if !visited[x][y] {
			printNARPixel(&narpixel, 0)
			if narpixel.x != x || narpixel.y != y {
				log.Panicf("mismatched coordinates, is: %v,%v => should be %v,%v", narpixel.x, narpixel.y, x, y)
			}
			color := color.RGBA{narpixel.Color.R, narpixel.Color.G, narpixel.Color.B, 255}
			drawAndMark(img, x, y, color, &visited)
			for h := uint8(0); h < narpixel.HSize; h++ {
				xH := x + uint16(h)
				drawAndMark(img, xH, y, color, &visited)
				if narpixel.VSize != nil && len(narpixel.VSize) > 0 {
					vsize := putBytesToUint16(narpixel.VSize[h])
					for v := uint16(0); v < vsize; v++ {
						yV := y + uint16(v)
						drawAndMark(img, xH, yV, color, &visited)
					}
				}
			}
		}
		for visited[x][y] {
			x++
			log.Print("inner loop", x)
			if x >= narpimage.Size.X {
				y++
				x = 0
			}
		}
		log.Print()
	}

	f, error := os.OpenFile(s, os.O_WRONLY|os.O_CREATE, 0666)
	if error != nil {
		return error
	}
	defer f.Close()
	png.Encode(f, img)

	return error
}

func printNARPixel(pixel *NotARegularPixel, hsizeThresh uint8) {
	if pixel == nil {
		log.Panicf("printNARPixel: Error, args are nil")
	}
	if pixel.HSize < hsizeThresh {
		return
	}

	s := ""
	log.Printf("Color: %v, horizontal size: %v, verticals: %v \n", pixel.Color, pixel.HSize, pixel.VSize)
	for k := uint8(0); k < pixel.HSize+1; k++ {
		s += "X"
	}
	log.Printf("X")

	vmax := uint16(0)
	for k := uint8(0); k < pixel.HSize+1; k++ {
		if pixel.VSize[k] != nil {
			vs := putBytesToUint16(pixel.VSize[k])
			if vs > vmax {
				vmax = vs
			}
		}
	}
	for v := uint16(1); v <= vmax; v++ {
		s = ""
		for k := uint8(0); k < pixel.HSize+1; k++ {
			if val, ok := pixel.VSize[k]; ok {
				vs := putBytesToUint16(val)
				if vs >= v {
					s += "X"
				} else {
					s += " "
				}
			}
		}
		log.Printf(s)
	}
}

func (narpimage *NARPImage) ConstructFromJpgFile(s string, showprogress bool) error {
	reader, err := os.Open(s)
	if err != nil {
		log.Println(err)
		return err
	}
	defer reader.Close()

	img, err := jpeg.Decode(reader)
	if err != nil {
		log.Println(err)
		return err
	}

	narpimage.initNARPImage()
	narpimage.putToNarpImage(img, showprogress)

	return nil
}

func (narpimage *NARPImage) Load(s string) error {
	file, err := os.Open(s)
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	narpimage.initNARPImage()

	err = gob.NewDecoder(file).Decode(narpimage)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

func (narpimage *NARPImage) Save(s string, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(s); !os.IsNotExist(err) {
			log.Fatalf("Save: error, file <%s> already exists", s)
			return err
		}
	}

	file, err := os.Create(s)
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	b := new(bytes.Buffer)
	err = gob.NewEncoder(b).Encode(narpimage)
	if err != nil {
		log.Println(err)
		return err
	}
	file.Write(b.Bytes())

	return err
}

func (narpimage *NARPImage) initNARPImage() {
	narpimage.NARPixels = []NotARegularPixel{}
	narpimage.Size = Point{0, 0}
	narpimage.Version = "0.5"
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
	initVisitedArray(&visited, img.Bounds().Max.X, img.Bounds().Max.Y)
	narpimage.Size.X, narpimage.Size.Y = uint16(img.Bounds().Max.X), uint16(img.Bounds().Max.Y)
	var boundsmin Point
	boundsmin.X, boundsmin.Y = uint16(img.Bounds().Min.X), uint16(img.Bounds().Min.Y)

	for y := boundsmin.Y; y < narpimage.Size.Y; y++ {
		showProgress(y, narpimage.Size.Y-1, showprogress)
		for x := boundsmin.X; x < narpimage.Size.X; x++ {
			if !(visited[x][y]) {
				narp := getNARP(x, y, img, &visited)
				narp.x = x
				narp.y = y
				narpimage.NARPixels = append(narpimage.NARPixels, *narp)
			}
		}
	}

	if showprogress {
		log.Println()
	}

	return nil
}

func getRGBAFFRange(imageClr color.Color) (uint8, uint8, uint8) {
	r, g, b, _ := imageClr.RGBA()
	return uint8(r / 257), uint8(g / 257), uint8(b / 257)
}

func compareColor(imageClr color.Color, narpColor RGBColor) bool {
	r, g, b := getRGBAFFRange(imageClr)
	if r == narpColor.R && g == narpColor.G && b == narpColor.B {
		return true
	}
	return false
}

func getNARP(x uint16, y uint16, img image.Image, visited *[][]bool) (narp *NotARegularPixel) {
	r, g, b := getRGBAFFRange(img.At(int(x), int(y)))
	narp = &NotARegularPixel{
		HSize: 0, VSize: map[uint8][]uint8{}, Color: RGBColor{r, g, b}}
	hsize := -1

	for xH := x; compareColor(img.At(int(xH), int(y)), narp.Color) && xH < uint16(img.Bounds().Max.X) && hsize < 253; xH++ {
		if !(*visited)[xH][y] {
			(*visited)[xH][y] = true
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
	r, g, b := getRGBAFFRange(img.At(int(x), int(y)))
	color := RGBColor{r, g, b}
	vsize := uint16(0)

	for yV := y + 1; yV < uint16(img.Bounds().Max.Y) && compareColor(img.At(int(x), int(yV)), color); yV++ {
		if !(*visited)[x][yV] {
			(*visited)[x][yV] = true
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
