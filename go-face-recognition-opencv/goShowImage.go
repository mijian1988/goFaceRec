package main

import (
	"fmt"
	"gocv.io/x/gocv"
	"image/color"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("How to run:\n\tshowimage [imgfile]")
		return
	}

	filename := os.Args[1]
	window := gocv.NewWindow("Hello")
	img := gocv.IMRead(filename, gocv.IMReadColor)
	if img.Empty() {
		fmt.Println("Error reading image from: %v", filename)
		return
	}

	///////////////////////////////////////////////////////////////
	// color for the rect when faces detected
	//blue := color.RGBA{0, 0, 255, 0}
	blue := color.RGBA{255, 0, 0, 0}
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load("data/haarcascade_frontalface_default.xml") {
		fmt.Println("Error reading cascade file: data/haarcascade_frontalface_default.xml")
		return
	}
	// detect faces
	rects := classifier.DetectMultiScale(img)
	fmt.Printf("found %d faces\n", len(rects))
	// draw a rectangle around each face on the original image
	for _, r := range rects {
		gocv.Rectangle(&img, r, blue, 1)
	}
	/////////////////////////////////////////////////////////////////////

	for {
		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}
