package main

import (
	"fmt"
	"reflect"

	"github.com/AmbigousParityBit/NARPImage"
)

func main() {
	s := "./testitem."

	narpimg := new(NARPImage.NARPImage)

	err := narpimg.ConstructFromJpgFile(s+"jpg", true)
	fmt.Println()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Constructed NARP image in memory from jpg file <%sjpg> ; bounds: %v, %v.\n",
		s, narpimg.Size.X, narpimg.Size.Y)
	fmt.Printf("Number of\n\t pixels = %v,\n\t keys = %v.\nGain in reduction of pixel objects: %v%%.\n",
		int(narpimg.Size.X)*int(narpimg.Size.Y), len(narpimg.NARPixels),
		100-100*len(narpimg.NARPixels)/(int(narpimg.Size.X)*int(narpimg.Size.Y)))

	err = narpimg.Save(s+"narp", true)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Saved NARP image from memory to file <%snarp>.\n", s)
	/*
		count := 0
		for point, pixel := range narpimg.NARPixels {
			fmt.Println(point, "::\t", pixel)
			count++
			if count == 10 {
				//			return
			}
		}
	*/

	fmt.Println()

	narpimgAfterLoading := new(NARPImage.NARPImage)
	err = narpimgAfterLoading.Load(s + "narp")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Loaded NARP image from file <%snarp> to memory.\n", s)
	if reflect.DeepEqual(narpimg, narpimgAfterLoading) {
		fmt.Printf("Loaded NARP image is the same as the previous one in memory.\n")

	} else {
		fmt.Printf("Loaded NARP image is different from the previous one in memory.\n")
	}
}
