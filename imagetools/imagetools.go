package imagetools

import (
	"fmt"
	"github.com/pkg/errors"
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"
)

/*
Rebuild function to create a big image with smaller ones
*/
func Rebuild(t [][]image.Image) image.Image {
	xmax := 0
	ymax := 0
	for y := 0; y < len(t); y++ { // compte le nombre de colonnes
		ymax = ymax + t[y][0].Bounds().Dy()
	}
	for x := 0; x < len(t[0]); x++ {
		xmax = xmax + t[0][x].Bounds().Dx()
	}
	//rectangle for the big image
	r := image.Rectangle{image.Point{0, 0}, image.Point{xmax, ymax}}

	rgba := image.NewRGBA(r)

	xi := 0
	yi := 0

	for y := 0; y < len(t); y++ { //y utilisé en hauteur
		xi = 0
		for x := 0; x < len(t[y]); x++ { //x utilisé en largeur
			pi := image.Point{xi, yi}
			ri := image.Rectangle{pi, pi.Add(t[y][x].Bounds().Size())}

			draw.Draw(rgba, ri, t[y][x], image.Point{0, yi}, draw.Src)

			xi = xi + t[y][x].Bounds().Dx()
		}
		yi = yi + t[y][0].Bounds().Dy()
	}

	return rgba
}

/*
RebuildChevauchement function pour refabriquer une image a partir de petites et d'un chevauchement
*/
func RebuildChevauchement(t [][]image.Image, pixs int) image.Image {
	xmax := 0
	ymax := 0
	for y := 0; y < len(t); y++ { // compte le nombre de colonnes
		ymax = ymax + t[y][0].Bounds().Dy()
	}
	// on enleve le chevauchement
	ymax -= pixs * (len(t) - 1)
	for x := 0; x < len(t[0]); x++ {
		xmax = xmax + t[0][x].Bounds().Dx()
	}
	//pas de chevauchement en x
	//rectangle for the big image
	r := image.Rectangle{image.Point{0, 0}, image.Point{xmax, ymax}}

	rgba := image.NewRGBA(r)

	xi := 0
	yi := 0

	for y := 0; y < len(t); y++ { //y utilisé en hauteur
		xi = 0
		for x := 0; x < len(t[y]); x++ { //x utilisé en largeur
			pi := image.Point{xi, yi}
			ri := image.Rectangle{pi, pi.Add(t[y][x].Bounds().Size())}

			draw.Draw(rgba, ri, t[y][x], image.Point{0, yi}, draw.Src)

			xi = xi + t[y][x].Bounds().Dx()
		}
		yi = yi + t[y][0].Bounds().Dy() - pixs
		if y == 0 {
			yi += int(pixs / 2)
		}
	}

	return rgba
}

/*
Open is a function to open a file as image
*/
func Open(filepath string) (image.Image, error) {

	infile, err := os.Open(filepath)

	if err != nil {
		error := errors.Wrap(err, "Open failed with error :")
		return nil, error
	}
	defer infile.Close()

	img, _, err := image.Decode(infile)
	if err != nil {
		error := errors.Wrap(err, "Open failed with error :")
		return nil, error
	}
	return img, nil
}

/*
Export create a png file based on the image file given
*/
func Export(img image.Image, name string) error {

	// Encode the image to the new file
	newFileName := name
	newfile, err := os.Create(newFileName)

	if err != nil {
		error := errors.Wrap(err, "Export failed with error :")
		return error
	}
	defer newfile.Close()

	err = png.Encode(newfile, img)

	if err != nil {
		error := errors.Wrap(err, "Export Failed with error :")
		return error
	}

	return nil
}

/*
Crop l'image en plusieurs images
*/
func Crop(img image.Image, nbSplit int) [][]image.Image {
	//traitements sur l'image
	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y

	slice := int(math.Floor(float64(h) / float64(nbSplit)))

	//Si on est dans le cas ou la division n'est pas entiere
	if h%nbSplit != 0 {
		slice++
	}

	//handle case where h = nbSplit * slice
	slices := make([][]image.Image, nbSplit)
	for y := range slices {
		slices[y] = make([]image.Image, 1)
	}

	cpt := 0

	for y := 0; y < h; y = y + slice {

		//create a subImage
		rect := image.Rect(0, y, w, min(y+slice, h))
		imgSliced := cropImage(img, rect)

		slices[cpt][0] = imgSliced
		cpt++
	}

	return slices
}

/*
CropChevauchement crop l'image en plusieurs images avec un chaevauchement entre elles
*/
func CropChevauchement(img image.Image, nbSplit int, pixs int) [][]image.Image {
	//traitements sur l'image
	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y

	slice := int(math.Floor(float64(h) / float64(nbSplit)))

	//Si on est dans le cas ou la division n'est pas entiere
	if h%nbSplit != 0 {
		slice++
	}

	//handle case where h = nbSplit * slice
	slices := make([][]image.Image, nbSplit)
	for y := range slices {
		slices[y] = make([]image.Image, 1)
	}

	cpt := 0

	for y := 0; y < h; y = y + slice {

		//create a subImage
		rect := image.Rect(0, max(0, y-pixs), w, min(y+slice, h))
		imgSliced := cropImage(img, rect)

		slices[cpt][0] = imgSliced
		cpt++
	}

	return slices
}

func cropImage(img image.Image, cropRect image.Rectangle) (cropImg image.Image) {
	//Interface for asserting whether `img`
	//implements SubImage or not.
	//This can be defined globally.
	type CropableImage interface {
		image.Image
		SubImage(r image.Rectangle) image.Image
	}

	if p, ok := img.(CropableImage); ok {
		// Call SubImage. This should be fast,
		// since SubImage (usually) shares underlying pixel.
		cropImg = p.SubImage(cropRect)
	} else if cropRect = cropRect.Intersect(img.Bounds()); !cropRect.Empty() {
		fmt.Println("else")
		// If `img` does not implement `SubImage`,
		// copy (and silently convert) the image portion to RGBA image.
		rgbaImg := image.NewRGBA(cropRect)
		for y := cropRect.Min.Y; y < cropRect.Max.Y; y++ {
			for x := cropRect.Min.X; x < cropRect.Max.X; x++ {
				rgbaImg.Set(x, y, img.At(x, y))
			}
		}
		cropImg = rgbaImg
	} else {
		// Return an empty RGBA image
		cropImg = &image.RGBA{}
	}

	return cropImg
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
