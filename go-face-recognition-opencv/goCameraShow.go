package main

import (
	"fmt"
	"gocv.io/x/gocv"
	"gopkg.in/eapache/queue.v1"
	"image/color"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

/*
func main() {
	// set to use a video capture device 0
	deviceID := 0
	//deviceID := "rtsp://192.168.0.10/live/live0"
	//deviceID := "rtsp://admin:cmiot123@192.168.0.100/"

	// open webcam
	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer webcam.Close()
	fmt.Println("open cam ok")

	// open display window
	window := gocv.NewWindow("Face Detect")
	defer window.Close()
	fmt.Println("NewWindow ok")

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	// color for the rect when faces detected
	blue := color.RGBA{0, 0, 255, 0}

	// load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load("data/haarcascade_frontalface_default.xml") {
		fmt.Println("Error reading cascade file: data/haarcascade_frontalface_default.xml")
		return
	}

	//for ffmpeg push to rtmp server
	width := int(webcam.Get(gocv.VideoCaptureFrameWidth))
	height := int(webcam.Get(gocv.VideoCaptureFrameHeight))
	fps := int(webcam.Get(gocv.VideoCaptureFPS))

	cmdArgs :=fmt.Sprintf("%s %s %s %d %s %s",
		"ffmpeg -y -an -f rawvideo -vcodec rawvideo -pix_fmt bgr24 -s",
		fmt.Sprintf("%dx%d", width, height),
		"-r",
		fps,
		"-i - -c:v libx264 -pix_fmt yuv420p -preset ultrafast -f flv",
		"rtmp://192.168.0.30:1935/live/movie",
	)
	fmt.Printf("cmdargs:%s\n",cmdArgs)
	list := strings.Split(cmdArgs, " ")
	cmd := exec.Command(list[0], list[1:]...)
	cmdIn, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	defer cmdIn.Close()
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("start reading camera device: %v\n", deviceID)
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("cannot read device %v\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}
		fmt.Println("read frame ok")

		// detect faces
		rects := classifier.DetectMultiScale(img)
		fmt.Printf("found %d faces\n", len(rects))

		// draw a rectangle around each face on the original image
		for _, r := range rects {
			gocv.Rectangle(&img, r, blue, 3)
		}

		// show the image in the window, and wait 1 millisecond
		window.IMShow(img)
		window.WaitKey(1)

		//push to rtmp server
		cnt,err :=cmdIn.Write([]byte(img.ToBytes()))
		//cnt,err :=cmdIn.Write(img.ToBytes())
		if err !=nil {
			fmt.Printf("%v",err)
			os.Exit(0)
		}else{
			fmt.Printf("send cnt=%d\n",cnt)
		}
	}
}*/

var wg sync.WaitGroup

func main() {
	frameQueue := queue.New()
	argsChan := make(chan string)

	go getFrameFromCamera(frameQueue,argsChan)
	wg.Add(1)

	time.Sleep(1)

	go recFaceAndPushToRtmp(frameQueue,argsChan)
	wg.Add(1)

	wg.Wait()
	fmt.Println("main exit...")
}

func getFrameFromCamera(queue *queue.Queue,wArgsChan chan<- string) {
	// set src
	deviceID := 0
	//deviceID := "rtsp://admin:cmiot123@192.168.0.100/"

	// open webCam
	webCam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer webCam.Close()
	fmt.Println("open cam ok")

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	//for ffmpeg push to rtmp server
	// get webCam ops:width/height/fpss
	width := int(webCam.Get(gocv.VideoCaptureFrameWidth))
	height := int(webCam.Get(gocv.VideoCaptureFrameHeight))
	fps := int(webCam.Get(gocv.VideoCaptureFPS))

	cmdArgs :=fmt.Sprintf("%s %s %s %d %s %s",
		"ffmpeg -y -an -f rawvideo -vcodec rawvideo -pix_fmt bgr24 -s",
		fmt.Sprintf("%dx%d", width, height),
		"-r",
		fps,
		"-i - -c:v libx264 -pix_fmt yuv420p -preset ultrafast -f flv",
		"rtmp://192.168.0.30:1935/live/movie",
	)
	//fmt.Printf("cmdargs:%s\n",cmdArgs)
	wArgsChan <-cmdArgs
	fmt.Printf("send cmdargs to push routine ok\n")

	for {
		if webCam.IsOpened() {
			// read frame from cam
			if ok := webCam.Read(&img); !ok {
				fmt.Printf("cannot read device %v\n", deviceID)
				break
			}
			if img.Empty() {
				continue
			}
			fmt.Println("read frame ok")

			// resize 320*240
			//gocv.Resize(img,&dstImg,dstImg,320,240,gocv.InterpolationCubic)

			// put frame into queue
			queue.Add(img)
		} else {
			fmt.Println("camera has been closed!")
			break
		}
	}

	wg.Done()
}

func recFaceAndPushToRtmp(queue *queue.Queue,rArgsChan <-chan string) {
	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	// open display window
	window := gocv.NewWindow("Face Detect")
	defer window.Close()
	fmt.Println("NewWindow ok")

	// color for the rect when faces detected
	blue := color.RGBA{0, 0, 255, 0}

	// load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load("data/haarcascade_frontalface_default.xml") {
		fmt.Println("Error reading cascade file: data/haarcascade_frontalface_default.xml")
		return
	}

	//for ffmpeg push to rtmp server
	cmdArgs := <-rArgsChan
	list := strings.Split(cmdArgs, " ")
	cmd := exec.Command(list[0], list[1:]...)
	cmdIn, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	defer cmdIn.Close()
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	for {
		if queue.Length() > 0 {
			queueImg := queue.Get(0)
			switch qImg := queueImg.(type) {
				case gocv.Mat:
					img = qImg
				default:
					continue
			}

			// detect faces
			rects := classifier.DetectMultiScale(img)
			fmt.Printf("found %d faces\n", len(rects))

			// draw a rectangle around each face on the original image
			for _, r := range rects {
				gocv.Rectangle(&img, r, blue, 3)
			}

			// show the image in the window, and wait 1 millisecond
			window.IMShow(img)
			window.WaitKey(1)

			//push to rtmp server
			cnt, err := cmdIn.Write([]byte(img.ToBytes()))
			if err != nil {
				fmt.Printf("%v", err)
				break
			} else {
				fmt.Printf("send cnt=%d\n", cnt)
			}
		}
	}

	wg.Done()
}
