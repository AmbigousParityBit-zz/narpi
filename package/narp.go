//not a regular pixel image
package narpi

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"sort"
)

type RGB8 = struct{ R, G, B uint8 }

type NotARegularPixel struct {
	HSize uint8             // horizontal flood size
	Color RGB8              // color of pixel
	VSize map[uint8][]uint8 // map has cells with count of vertical pixels of the same color
	//
	//	ex.: HSize=3, VSize[2]=[2,1,2] means area of the same color:
	//		X	X	X
	//			X
	//			X
	//			X
}

func (pixel *NotARegularPixel) RunesArray() (s [][]rune) {
	hsize := pixel.HSize + 1
	s = make([][]rune, hsize)

	vsize := uint16(0)
	for i := uint8(0); i < hsize; i++ {
		if va, ok := pixel.VSize[i]; ok {
			v := putBytesToUint16(va) + 1
			if v > vsize {
				vsize = v
			}
		}
	}
	if vsize == 0 {
		vsize++
	}
	for i := uint8(0); i < hsize; i++ {
		s[i] = make([]rune, vsize)
	}

	for j := uint16(0); j < vsize; j++ {
		for i := uint8(0); i < hsize; i++ {
			if val, ok := pixel.VSize[i]; ok {
				vs := putBytesToUint16(val) + 1
				if j < vs {
					s[i][j] = 'o'
				} else {
					s[i][j] = '.'
				}
			} else {
				s[i][0] = 'o'
			}
		}
	}

	return s
}

func printVerticals(m map[uint8][]uint8) (r string) {
	var keys []int
	for k := range m {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	r = ""
	for _, k := range keys {
		r += fmt.Sprintf("%v:%d ", k+1, putBytesToUint16(m[uint8(k)])+1)
	}

	if r == "" {
		r = "empty"
	}
	return r
}

func (pixel *NotARegularPixel) Print(prefix string) {
	s := pixel.RunesArray()

	log.Printf("--------------------------------------------------------------------------------------")
	log.Printf("(NotARegularPixel):: %vcolor:%v, hsize:%v, size(runes, human friendly):%vx%v, \n\nvsize:  %v\nverticals(human friendly):  %v\n",
		prefix, pixel.Color, pixel.HSize, len(s), len(s[0]), pixel.VSize, printVerticals(pixel.VSize))

	for j := 0; j < len(s[0]); j++ {
		st := ""
		for i := 0; i < len(s); i++ {
			st += string(s[i][j])
		}
		log.Println(st)
	}
	log.Printf("--------------------------------------------------------------------------------------")
}

func getRGBA8(rgba16 color.Color) (uint8, uint8, uint8) {
	r, g, b, _ := rgba16.RGBA()
	return uint8(r / 257), uint8(g / 257), uint8(b / 257)
}

func colorsEqual(img *image.RGBA, x, y int, rgb8 RGB8) bool {
	firstb := y*img.Stride + x*4

	if img.Pix[firstb] == rgb8.R && img.Pix[firstb+1] == rgb8.G && img.Pix[firstb+2] == rgb8.B {
		return true
	}
	return false
}

func (pixel *NotARegularPixel) ReadBytes(b *bytes.Buffer) {
	*pixel = NotARegularPixel{
		HSize: 0, VSize: map[uint8][]uint8{}, Color: RGB8{0, 0, 0}}
	var err error

	pixel.Color.R, err = b.ReadByte()
	if err != nil {
		log.Fatalf(err.Error())
	}

	pixel.Color.G, err = b.ReadByte()
	if err != nil {
		log.Fatalf(err.Error())
	}

	pixel.Color.B, err = b.ReadByte()
	if err != nil {
		log.Fatalf(err.Error())
	}

	pixel.HSize, err = b.ReadByte()
	if err != nil {
		log.Fatalf(err.Error())
	}

	if pixel.HSize > 0 {
		pixel.VSize = make(map[uint8][]uint8, pixel.HSize)
		for i := 0; i < pixel.HSize; i++ {
			flag, err = b.ReadByte()
			if err != nil {
				log.Fatalf(err.Error())
			}

		}

		{
		}
		hslice := byte(0)
		counter := byte(0)
		flag := byte(0)
		for i := byte(0); i < pixel.HSize; i = counter*8 + hslice {
			if len(pixel.VSize[i]) > 1 {
				flag = flag | byte(math.Pow(2, float64(i)))
			}
			hslice++
			if hslice == 8 {
				hslice = 0
				counter++
				b.WriteByte(flag)
				flag = 0
			}
		}

		for i := byte(0); i < pixel.HSize; i++ {
			l := len(pixel.VSize[i])
			if l > 0 {
				if l > 1 {
					b.WriteByte()
				}
				b.WriteByte(pixel.VSize[i][0])
				if l > 1 {
					b.WriteByte(pixel.VSize[i][1])
				}
			}
		}
	}

	log.Println(pixel.Color, pixel.HSize)
}

func (pixel *NotARegularPixel) Bytes() (b *bytes.Buffer) {
	b = new(bytes.Buffer)
	b.Reset()

	b.WriteByte(pixel.Color.R)
	b.WriteByte(pixel.Color.G)
	b.WriteByte(pixel.Color.B)
	b.WriteByte(pixel.HSize)
	if pixel.HSize > 0 {
		hslice := byte(0)
		counter := byte(0)
		flag := byte(0)
		for i := byte(0); i < pixel.HSize; i = counter*8 + hslice {
			if len(pixel.VSize[i]) > 1 {
				flag = flag | byte(math.Pow(2, float64(i)))
			}
			hslice++
			if hslice == 8 {
				hslice = 0
				counter++
				b.WriteByte(flag)
				flag = 0
			}
		}

		for i := byte(0); i < pixel.HSize; i++ {
			l := len(pixel.VSize[i])
			if l > 0 {
				b.WriteByte(pixel.VSize[i][0])
				if l > 1 {
					b.WriteByte(pixel.VSize[i][1])
				}
			}
		}
	}

	return b
}
