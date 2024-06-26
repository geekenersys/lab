package main

import (
	"fmt"
	"image"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/go-vgo/robotgo"
)

// New function for bitmap operations
func bitmap() {
	bit := robotgo.CaptureScreen()
	defer robotgo.FreeBitmap(bit)
	fmt.Println("abitMap...", bit)

	gbit := robotgo.ToBitmap(bit)
	fmt.Println("bitmap...", gbit.Width)

	gbitMap := robotgo.CaptureGo()
	fmt.Println("Go CaptureScreen...", gbitMap.Width)

	// img := robotgo.CaptureImg()
	img := robotgo.CaptureImg(0, 0, 1920, 1080)
	robotgo.SavePng(img, "save.png")

	num := robotgo.DisplaysNum()
	for i := 0; i < num; i++ {
		robotgo.DisplayID = i
		img1 := robotgo.CaptureImg()
		path1 := "save_" + strconv.Itoa(i)
		robotgo.SavePng(img1, path1+".png")
		robotgo.SaveJpeg(img1, path1+".jpeg", 50)

		img2 := robotgo.CaptureImg(10, 10, 20, 20)
		path2 := "test_" + strconv.Itoa(i)
		robotgo.SavePng(img2, path2+".png")
		robotgo.SaveJpeg(img2, path2+".jpeg", 50)
	}
}

// New function for color operations
func color() {
	color := robotgo.GetPixelColor(100, 200)
	fmt.Println("color----", color, "-----------------")

	clo := robotgo.GetPxColor(100, 200)
	fmt.Println("color...", clo)
	clostr := robotgo.PadHex(clo)
	fmt.Println("color...", clostr)

	rgb := robotgo.RgbToHex(255, 100, 200)
	rgbstr := robotgo.PadHex(robotgo.U32ToHex(rgb))
	fmt.Println("rgb...", rgbstr)

	hex := robotgo.HexToRgb(uint32(rgb))
	fmt.Println("hex...", hex)
	hexh := robotgo.PadHex(robotgo.U8ToHex(hex))
	fmt.Println("HexToRgb...", hexh)

	color2 := robotgo.GetPixelColor(10, 20)
	fmt.Println("color---", color2)
}

// New function for screen operations
func screen() {
	bitmap()

	sx, sy := robotgo.GetScreenSize()
	fmt.Println("get screen size: ", sx, sy)
	for i := 0; i < robotgo.DisplaysNum(); i++ {
		s1 := robotgo.ScaleF(i)
		fmt.Println("ScaleF: ", s1)
	}
	sx, sy = robotgo.GetScaleSize()
	fmt.Println("get screen scale size: ", sx, sy)

	color()
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		}
	}

	return "Unable to get IP"
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	ip := getLocalIP()
	fmt.Fprintf(w, "RobotGo HTTP Server. Container IP: %s", ip)
}

func handleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	command := r.URL.Query().Get("cmd")
	if command == "" {
		http.Error(w, "cmd parameter is required", http.StatusBadRequest)
		return
	}

	log.Printf("Received command: %s", command)

	switch command {
	case "open_chrome":
		go func() {
			err := exec.Command("chromium-browser", "--start-fullscreen").Start()
			if err != nil {
				log.Printf("Failed to open Chromium: %v", err)
			}
		}()
		fmt.Fprintf(w, "Opening Chromium in fullscreen mode")
	case "move_mouse":
		x := r.URL.Query().Get("x")
		y := r.URL.Query().Get("y")
		if x == "" || y == "" {
			http.Error(w, "x and y parameters are required", http.StatusBadRequest)
			return
		}
		fmt.Printf("Moving mouse to (%s, %s)\n", x, y)
		robotgo.MoveMouse(atoi(x), atoi(y))
		fmt.Fprintf(w, "Mouse moved to (%s, %s)", x, y)
	case "click_mouse":
		fmt.Println("Clicking mouse")
		robotgo.MouseClick("left")
		fmt.Fprintf(w, "Mouse clicked")
	case "take_screenshot":
		var sharedDir string
		if runtime.GOOS == "windows" {
			sharedDir = filepath.Join(".", "shared")
		} else {
			sharedDir = "/shared"
		}
		log.Printf("Creating shared directory: %s", sharedDir)
		err := os.MkdirAll(sharedDir, 0777)
		if err != nil {
			log.Printf("Failed to create directory: %v", err)
			http.Error(w, fmt.Sprintf("Failed to create directory: %v", err), http.StatusInternalServerError)
			return
		}

		screenshotType := r.URL.Query().Get("type")
		if screenshotType == "" {
			http.Error(w, "type parameter is required", http.StatusBadRequest)
			return
		}

		fileName := fmt.Sprintf("%s/screenshot_%d.png", sharedDir, time.Now().Unix())
		log.Printf("Screenshot will be saved to: %s", fileName)

		var img image.Image

		if screenshotType == "fullscreen" {
			sx, sy := robotgo.GetScreenSize()
			fmt.Printf("Screen Size: %dx%d\n", sx, sy)
			log.Println("Capturing full screen image")
			img = robotgo.CaptureImg(0, 0, sx, sy)
		} else if screenshotType == "mouse" {
			x, y := robotgo.GetMousePos()
			width, height := 300, 200 // Specify the size around the mouse pointer
			fmt.Printf("Capturing image around mouse at (%d, %d)\n", x, y)
			log.Println("Capturing image around mouse pointer")
			img = robotgo.CaptureImg(x-int(width/2), y-int(height/2), width, height)
		} else {
			http.Error(w, "Invalid type parameter", http.StatusBadRequest)
			return
		}

		// Save the captured screen to a file
		log.Println("Saving screenshot")
		err = robotgo.Save(img, fileName)
		if err != nil {
			log.Printf("Failed to save screenshot: %v", err)
			http.Error(w, fmt.Sprintf("Failed to save screenshot: %v", err), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Screenshot taken and saved to %s", fileName)
	default:
		http.Error(w, "Unknown command", http.StatusBadRequest)
	}
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func main() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/command", handleCommand)
	ip := getLocalIP()
	fmt.Printf("Starting server on :8081. Container IP: %s\n", ip)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
