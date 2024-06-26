package main

import (
	"bufio"
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
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

// Action represents a user action
type Action struct {
	Timestamp  time.Time
	ActionType string
	X          int
	Y          int
	Button     string
}

var actions []Action

func recordMouseEvents() {
	evChan := hook.Start()
	defer hook.End()

	for ev := range evChan {
		if ev.Kind == hook.MouseMove {
			actions = append(actions, Action{
				Timestamp:  time.Now(),
				ActionType: "move",
				X:          int(ev.X),
				Y:          int(ev.Y),
			})
		} else if ev.Kind == hook.MouseDown || ev.Kind == hook.MouseUp {
			button := "left"
			if ev.Button == hook.MouseMap["mright"] {
				button = "right"
			}
			actionType := "click"
			if ev.Clicks == 2 {
				actionType = "double_click"
			}
			actions = append(actions, Action{
				Timestamp:  time.Now(),
				ActionType: actionType,
				X:          int(ev.X),
				Y:          int(ev.Y),
				Button:     button,
			})
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func saveActionsToFile(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	for _, action := range actions {
		fmt.Fprintf(file, "%s,%s,%d,%d,%s\n", action.Timestamp.Format(time.RFC3339), action.ActionType, action.X, action.Y, action.Button)
	}
}

func loadActionsFromFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) != 5 {
			continue
		}

		timestamp, err := time.Parse(time.RFC3339, parts[0])
		if err != nil {
			log.Printf("Failed to parse timestamp: %v", err)
			continue
		}

		x, err := strconv.Atoi(parts[2])
		if err != nil {
			log.Printf("Failed to parse X coordinate: %v", err)
			continue
		}

		y, err := strconv.Atoi(parts[3])
		if err != nil {
			log.Printf("Failed to parse Y coordinate: %v", err)
			continue
		}

		actions = append(actions, Action{
			Timestamp:  timestamp,
			ActionType: parts[1],
			X:          x,
			Y:          y,
			Button:     parts[4],
		})
	}
}

func replayActions() {
	startTime := actions[0].Timestamp
	for _, action := range actions {
		time.Sleep(action.Timestamp.Sub(startTime))
		startTime = action.Timestamp

		switch action.ActionType {
		case "move":
			robotgo.MoveMouse(action.X, action.Y)
		case "click":
			robotgo.MouseClick(action.Button)
		case "double_click":
			robotgo.MouseClick(action.Button, true)
		}
	}
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
		log.Println("Invalid request method:", r.Method)
		return
	}

	command := r.URL.Query().Get("cmd")
	if command == "" {
		http.Error(w, "cmd parameter is required", http.StatusBadRequest)
		log.Println("cmd parameter is required")
		return
	}

	log.Printf("Received command: %s", command)

	switch command {
	case "open_chrome":
		log.Println("Executing command: open_chrome")
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
			log.Println("x and y parameters are required")
			return
		}
		log.Printf("Moving mouse to (%s, %s)\n", x, y)
		robotgo.MoveMouse(atoi(x), atoi(y))
		fmt.Fprintf(w, "Mouse moved to (%s, %s)", x, y)
	case "click_mouse":
		button := r.URL.Query().Get("button")
		if button == "" {
			button = "left"
		}
		log.Printf("Clicking %s mouse button\n", button)
		robotgo.MouseClick(button)
		fmt.Fprintf(w, "Mouse %s button clicked", button)
	case "double_click_mouse":
		button := r.URL.Query().Get("button")
		if button == "" {
			button = "left"
		}
		log.Printf("Double clicking %s mouse button\n", button)
		robotgo.MouseClick(button, true)
		fmt.Fprintf(w, "Mouse %s button double clicked", button)
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
			log.Println("type parameter is required")
			return
		}

		fileName := fmt.Sprintf("%s/screenshot_%d.png", sharedDir, time.Now().Unix())
		log.Printf("Screenshot will be saved to: %s", fileName)

		var img image.Image

		if screenshotType == "fullscreen" {
			sx, sy := robotgo.GetScreenSize()
			log.Printf("Screen Size: %dx%d\n", sx, sy)
			log.Println("Capturing full screen image")
			img = robotgo.CaptureImg(0, 0, sx, sy)
		} else if screenshotType == "mouse" {
			x, y := robotgo.GetMousePos()
			width, height := 300, 200 // Specify the size around the mouse pointer
			log.Printf("Capturing image around mouse at (%d, %d)\n", x, y)
			log.Println("Capturing image around mouse pointer")
			img = robotgo.CaptureImg(x-int(width/2), y-int(height/2), width, height)
		} else {
			http.Error(w, "Invalid type parameter", http.StatusBadRequest)
			log.Println("Invalid type parameter")
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
		log.Println("Unknown command")
	}
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func main() {
	go recordMouseEvents()

	// Run for a fixed amount of time
	time.Sleep(10 * time.Second)

	saveActionsToFile("actions.txt")
	fmt.Println("Actions recorded and saved to actions.txt")

	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/command", handleCommand)
	ip := getLocalIP()
	fmt.Printf("Starting server on :8081. Container IP: %s\n", ip)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
