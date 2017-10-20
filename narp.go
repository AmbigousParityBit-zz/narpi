//not a regular pixel image
package NARPImage

import (
	"fmt"
	"image/color"
	"log"
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
	if pixel.HSize == 0 {
		s[0] = make([]rune, 1)
		s[0][0] = 'o'
		return s
	}

	vsize := uint16(0)
	for i := uint8(0); i < hsize; i++ {
		if va, ok := pixel.VSize[i]; ok {
			v := putBytesToUint16(va) + 1
			if v > vsize {
				vsize = v
			}
		}
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
		r = "nil"
	}
	return r
}

func (pixel *NotARegularPixel) Print(prefix string) {
	s := pixel.RunesArray()

	log.Printf("______________________________________________________________________________________")
	log.Printf("(NotARegularPixel):: %vcolor:%v, size(horizontal x max_vertical):%vx%v, verticals: \n%v \n", prefix, pixel.Color, len(s), len(s[0]), printVerticals(pixel.VSize))
	log.Printf("======================================================================================")

	for j := 0; j < len(s[0]); j++ {
		st := ""
		for i := 0; i < len(s); i++ {
			st += string(s[i][j])
		}
		log.Println(st)
	}
	log.Printf("======================================================================================")
}

func getRGBA8(imageClr color.Color) (uint8, uint8, uint8) {
	r, g, b, _ := imageClr.RGBA()
	return uint8(r / 257), uint8(g / 257), uint8(b / 257)
}

func colorsEqual(imageClr color.Color, narpColor RGB8) bool {
	r, g, b := getRGBA8(imageClr)
	if r == narpColor.R && g == narpColor.G && b == narpColor.B {
		return true
	}
	return false
}
