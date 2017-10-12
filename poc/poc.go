package main

import (
	"fmt"
	"narpimage"
	"reflect"
)

func main() {
	s := "./testitem."

	narpimg := new(narpimage.NARPImage)

	err := narpimg.ConstructFromJpgFile(s+"jpg", true)
	fmt.Println()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Constructed NARP image in memory from jpg file <%sjpg> ; bounds: %v, %v.\n",
		s, narpimg.Size.X, narpimg.Size.Y)

	err = narpimg.Save(s+"narp", true)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Saved NARP image from memory to file <%snarp>.\n", s)

	narpimgAfterLoading := new(narpimage.NARPImage)
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
