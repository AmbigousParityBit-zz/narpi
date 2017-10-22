package narpi

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func _TestcutBytesOfUint16(t *testing.T) {
	var bytesCutTests = []struct {
		in   uint16
		outB bool
		outL uint8
		outR uint8
	}{
		{60116, true, 234, 212},
		{14231, true, 111, 23},
		{16383, true, 255, 255},
		{255, false, 255, 255},
	}

	for _, bt := range bytesCutTests {
		if outB_, outL_, outR_ := cutBytesOfUint16(bt.in); bt.outB != outB_ &&
			bt.outL != outL_ && bt.outR != outR_ {
			t.Fatalf("cutBytesOfUint16(%v) => %v,%v,%v, want %v,%v,%v", bt.in, outB_, outL_, outR_,
				bt.outB, bt.outL, bt.outR)
		} else {
			log.Printf("cutBytesOfUint16(%v) => %v,%v,%v, successfully", bt.in, bt.outB, bt.outL, bt.outR)
		}
	}
}

func _TestputBytesToUint16(t *testing.T) {
	var bytesPutTests = []struct {
		in  []uint8
		out uint16
	}{
		{[]uint8{122}, 122},
		{[]uint8{234, 212}, 60116},
		{[]uint8{111, 23}, 28439},
		{[]uint8{255, 255}, 65535},
	}

	for _, bt := range bytesPutTests {
		if v := putBytesToUint16(bt.in); v != bt.out {
			t.Fatalf("putBytesToUint16(%v) => %v, want %v", bt.in, v, bt.out)
		} else {
			log.Printf("putBytesToUint16(%v) => %v, successfully", bt.in, bt.out)
		}
	}
}

func getTestImagesFilenames(t *testing.T, ext string) []string {
	defer timeTrack(time.Now(), "getTestImagesFilenames")
	filenames, err := filepath.Glob("./testimages/*." + ext)
	if err != nil {
		t.Fatal(err)
	}

	fileNames := []string{}
	for _, f := range filenames {
		filename, _ := filepath.Abs(f)
		extension := path.Ext(filename)
		if extension == "."+ext {
			filename := filename[0:len(filename)-len(extension)] + "."
			fileNames = append(fileNames, filename)
		}
	}
	return fileNames
}

func deleteExtFiles(t *testing.T, ext string) {
	fileNames := getTestImagesFilenames(t, ext)
	for _, s := range fileNames {
		err := os.Remove(s + ext)
		if err != nil {
			t.Fatalf("couldn't delete file %v", s+ext)
		}
	}
	fileNames = getTestImagesFilenames(t, ext)
	if len(fileNames) > 0 {
		t.Fatalf("couldn't delete files <*%v> during preparations to tests", ext)
	}
	log.Printf("Successfully deleted all <*%v> files during preparations to tests.", ext)
}

func compareJpgPngFiles(t *testing.T, prfx string) {
	ext1, ext2 := "png", "jpg"

	if _, err := os.Stat(prfx + ext1); os.IsNotExist(err) {
		t.Fatalf("File <%s> doesn't exists.", prfx+ext1)
	}
	if _, err := os.Stat(prfx + ext2); os.IsNotExist(err) {
		t.Fatalf("File <%s> doesn't exists.", prfx+ext2)
	}

	img1, err := loadPng(prfx + ext1)
	if err != nil {
		t.Fatal(err)
	}
	img2, err := loadJpg(prfx + ext2)
	if err != nil {
		t.Fatal(err)
	}

	maxx := img1.Bounds().Max.X
	maxy := img1.Bounds().Max.Y

	if maxx != img2.Bounds().Max.X || maxy != img2.Bounds().Max.Y {
		t.Fatalf("Images <%s>, <%s> have different sizes.", prfx+ext1, prfx+ext2)
	}

	count := 0
	if !reflect.DeepEqual(img1, img2) {
		for x := 0; x < maxx; x++ {
			for y := 0; y < maxy; y++ {
				//if !reflect.DeepEqual(img1.At(x, y), img2.At(x, y)) {
				r1 := img1.At(x, y)
				r2 := img2.At(x, y)
				if r1 != r2 {
					t.Log(r1, r2)
					count++
				}
				//}
			}
		}
	}

	if count == 0 {
		log.Printf("Success: images <%s>, <%s> are identical.", prfx+ext1, prfx+ext2)
	} else {
		t.Fatalf("Error: images <%s>, <%s> differ in %v pixels (%.2f%% different).", prfx+ext1, prfx+ext2,
			count, float32(count)/float32(maxx*maxy)*100)
	}
}

func testConstructFromJpgFile(s string, narpimg *NARPImage, t *testing.T) {
	defer timeTrack(time.Now(), "testConstructFromJpgFile")
	log.Printf("Constructing NARP image in memory from jpg file <%sjpg>.\n", s)
	err := narpimg.LoadJpg(s+"jpg", false)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Constructed NARP image in memory from jpg file <%sjpg> ; bounds: %v, %v.\n",
		s, narpimg.Size.X, narpimg.Size.Y)
	log.Printf("Number of\n\t\t pixels = %v,\n\t\t keys = %v.\nGain in reduction of pixel objects: %v%%.\n",
		int(narpimg.Size.X)*int(narpimg.Size.Y), len(narpimg.NARPixels),
		100-100*len(narpimg.NARPixels)/(int(narpimg.Size.X)*int(narpimg.Size.Y)))
}

func testConstructFromPngFile(s string, narpimg *NARPImage, t *testing.T) {
	defer timeTrack(time.Now(), "testConstructFromPngFile")
	log.Printf("Constructing NARP image in memory from png file <%spng>.\n", s)
	err := narpimg.LoadJpg(s+"png", false)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Constructed NARP image in memory from png file <%spng> ; bounds: %v, %v.\n",
		s, narpimg.Size.X, narpimg.Size.Y)
	log.Printf("Number of\n\t\t pixels = %v,\n\t\t keys = %v.\nGain in reduction of pixel objects: %v%%.\n",
		int(narpimg.Size.X)*int(narpimg.Size.Y), len(narpimg.NARPixels),
		100-100*len(narpimg.NARPixels)/(int(narpimg.Size.X)*int(narpimg.Size.Y)))
}

func testSave(s string, narpimg *NARPImage, t *testing.T) {
	defer timeTrack(time.Now(), "Save")
	err := narpimg.Save(s+"narp", true)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Saved NARP image from memory to file <%snarp>.\n", s)
}

func testLoad(s string, narpimgAfterLoading *NARPImage, t *testing.T) {
	defer timeTrack(time.Now(), "Load")
	err := narpimgAfterLoading.Load(s + "narp")
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Loaded NARP image from file <%snarp> to memory.\n", s)
}

func testDeconstructToPngFile(s string, narpimg *NARPImage, t *testing.T) {
	defer timeTrack(time.Now(), "testDeconstructToPngFile")
	err := narpimg.Png(s + "png")
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Saved NARP image from memory to file <%spng>.\n", s)
}

func testDeconstructToJpgFile(s string, narpimg *NARPImage, t *testing.T) {
	defer timeTrack(time.Now(), "testDeconstructToJpgFile")
	err := narpimg.Jpg(s + "jpg")
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Saved NARP image from memory to file <%sjpg>.\n", s)
}

func TestImageFilesJpgToNARPIToPng(t *testing.T) {
	fileNames := getTestImagesFilenames(t, "jpg")
	deleteExtFiles(t, "png")
	for _, s := range fileNames {
		t.Run("ConstructFromJpgFile-Save-Load-DeconstructToPngFile::"+filepath.Base(s), func(t *testing.T) {
			narpimg := NARPImage{}
			narpimgAfterLoading := NARPImage{}

			testConstructFromJpgFile(s, &narpimg, t)
			testSave(s, &narpimg, t)
			testLoad(s, &narpimgAfterLoading, t)
			if reflect.DeepEqual(narpimg, narpimgAfterLoading) {
				log.Printf("Loaded NARP image is the same as the previous one in memory, as expected.\n")
			} else {
				t.Fatalf("Loaded NARP image is different from the previous one in memory.\n")
			}
			testDeconstructToPngFile(s, &narpimgAfterLoading, t)

			//compareJpgPngFiles(t, s)
		})
	}
}

func _TestImageFilesPngToNARPIToJpg(t *testing.T) {
	fileNames := getTestImagesFilenames(t, "png")
	deleteExtFiles(t, "jpg")
	for _, s := range fileNames {
		t.Run("ConstructFromPngFile-Save-Load-DeconstructToJpgFile::"+filepath.Base(s), func(t *testing.T) {
			narpimg := NARPImage{}
			narpimgAfterLoading := NARPImage{}

			testConstructFromPngFile(s, &narpimg, t)
			testSave(s, &narpimg, t)
			testLoad(s, &narpimgAfterLoading, t)
			if reflect.DeepEqual(narpimg, narpimgAfterLoading) {
				log.Printf("Loaded NARP image is the same as the previous one in memory, as expected.\n")
			} else {
				t.Fatalf("Loaded NARP image is different from the previous one in memory.\n")
			}
			testDeconstructToJpgFile(s, &narpimgAfterLoading, t)

			compareJpgPngFiles(t, s)
		})
	}
}
