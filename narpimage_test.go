package NARPImage

import (
	"path"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_cutBytesOfUint16(t *testing.T) {
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
			t.Logf("cutBytesOfUint16(%v) => %v,%v,%v, successfully", bt.in, bt.outB, bt.outL, bt.outR)
		}
	}
}

func Test_putBytesToUint16(t *testing.T) {
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
			t.Logf("putBytesToUint16(%v) => %v, successfully", bt.in, bt.out)
		}
	}
}

func getTestImagesFilenames(t *testing.T) []string {
	filenames, err := filepath.Glob("./testimages/*.jpg")
	if err != nil {
		t.Fatal(err)
	}

	fileNames := []string{}
	for _, f := range filenames {
		filename, _ := filepath.Abs(f)
		extension := path.Ext(filename)
		if extension == ".jpg" {
			filename := filename[0:len(filename)-len(extension)] + "."
			fileNames = append(fileNames, filename)
		}
	}
	return fileNames
}

func testConstructFromJpgFile(s string, narpimg *NARPImage, t *testing.T) {
	t.Logf("Constructing NARP image in memory from jpg file <%sjpg>.\n", s)
	err := narpimg.ConstructFromJpgFile(s+"jpg", false)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Constructed NARP image in memory from jpg file <%sjpg> ; bounds: %v, %v.\n",
		s, narpimg.Size.X, narpimg.Size.Y)
	t.Logf("Number of\n\t\t pixels = %v,\n\t\t keys = %v.\nGain in reduction of pixel objects: %v%%.\n",
		int(narpimg.Size.X)*int(narpimg.Size.Y), len(narpimg.NARPixels),
		100-100*len(narpimg.NARPixels)/(int(narpimg.Size.X)*int(narpimg.Size.Y)))
}

func testSave(s string, narpimg *NARPImage, t *testing.T) {
	err := narpimg.Save(s+"narp", true)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Saved NARP image from memory to file <%snarp>.\n", s)
}

func testLoad(s string, narpimgAfterLoading *NARPImage, t *testing.T) {
	err := narpimgAfterLoading.Load(s + "narp")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Loaded NARP image from file <%snarp> to memory.\n", s)
}

func testDeconstructToPngFile(s string, narpimg *NARPImage, t *testing.T) {
	err := narpimg.DeconstructToPngFile(s + "png")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Saved NARP image from memory to file <%spng>.\n", s)
}

func Test_ImageFiles(t *testing.T) {
	fileNames := getTestImagesFilenames(t)
	for _, s := range fileNames {
		short := filepath.Base(s)
		narpimg := NARPImage{}
		narpimgAfterLoading := NARPImage{}

		t.Run("ConstructFromJpgFile::"+short, func(t1 *testing.T) {
			testConstructFromJpgFile(s, &narpimg, t1)
		})
		t.Run("Save::"+short, func(t *testing.T) {
			testSave(s, &narpimg, t)
		})
		t.Run("Load::"+short, func(t *testing.T) {
			testLoad(s, &narpimgAfterLoading, t)

			if reflect.DeepEqual(narpimg, narpimgAfterLoading) {
				t.Logf("Loaded NARP image is the same as the previous one in memory, as expected.\n")
			} else {
				t.Fatalf("Loaded NARP image is different from the previous one in memory.\n")
			}
		})
		t.Run("DeconstructToPngFile::"+short, func(t *testing.T) {
			testDeconstructToPngFile(s, &narpimg, t)
		})
	}
}
