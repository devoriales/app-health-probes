// Author: Aleksandro Matejic, devoriales.com
// Version: 0.1
// Date: September 27, 2024
// Description: This Go program is a basic web server with liveness, readiness, and startup probes
// to simulate Kubernetes health checks for monitoring purposes.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	simulateLivenessFailure int32
	simulateReadinessFailure int32
	startupComplete int32
	probeTimestamps sync.Mutex
	timestamps = make(map[string]time.Time)
)

// Function to simulate a long startup time
func simulateLongStartup(seconds int) {
	time.Sleep(time.Duration(seconds) * time.Second)
	atomic.StoreInt32(&startupComplete, 1)
	setProbeTimestamp("startupProbe")
	_, err := os.Create("/tmp/startup-file")
	if err != nil {
		log.Fatalf("Failed to create startup complete file: %v", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Valkyrie Application</h1>")
	fmt.Fprintf(w, "<br>")
	fmt.Fprintf(w, "<br><a href='/liveness-health'>Liveness Health</a>")
	fmt.Fprintf(w, "<br><a href='/readiness-health'>Readiness Health</a>")
	fmt.Fprintf(w, "<br><a href='/timestamps'>Timestamps</a>")
	fmt.Fprintf(w, "<br><br>")
	fmt.Fprintf(w, "<label for='liveness-failure-toggle'>Simulate Liveness Failure:</label>")
	fmt.Fprintf(w, "<input type='checkbox' id='liveness-failure-toggle' onclick='toggleLivenessFailure()' %s>", getToggleChecked(&simulateLivenessFailure))
	fmt.Fprintf(w, "<br>")
	fmt.Fprintf(w, "<label for='readiness-failure-toggle'>Simulate Readiness Failure:</label>")
	fmt.Fprintf(w, "<input type='checkbox' id='readiness-failure-toggle' onclick='toggleReadinessFailure()' %s>", getToggleChecked(&simulateReadinessFailure))
	fmt.Fprintf(w, "<br><br>")
	fmt.Fprintf(w, "<strong>Liveness Status: </strong><span id='liveness-indicator'>%s</span>", getStatusIndicator(&simulateLivenessFailure, "liveness"))
	fmt.Fprintf(w, "<br><strong>Readiness Status: </strong><span id='readiness-indicator'>%s</span>", getStatusIndicator(&simulateReadinessFailure, "readiness"))
	fmt.Fprintf(w, "<script>")
	fmt.Fprintf(w, "function toggleLivenessFailure() { fetch('/toggle-liveness-failure').then(response => response.text()).then(data => { console.log(data); updateStatus(); }); }")
	fmt.Fprintf(w, "function toggleReadinessFailure() { fetch('/toggle-readiness-failure').then(response => response.text()).then(data => { console.log(data); updateStatus(); }); }")
	fmt.Fprintf(w, "function updateStatus() {")
	fmt.Fprintf(w, "fetch('/liveness-health').then(response => { if (response.ok) { return response.json(); } else { return response.text(); } }).then(data => { if (typeof data === 'object' && data.status) { document.getElementById('liveness-indicator').innerText = data.status; } else { document.getElementById('liveness-indicator').innerText = data; } });")
	fmt.Fprintf(w, "fetch('/readiness-health').then(response => { if (response.ok) { return response.json(); } else { return response.text(); } }).then(data => { if (typeof data === 'object' && data.status) { document.getElementById('readiness-indicator').innerText = data.status; } else { document.getElementById('readiness-indicator').innerText = data; } });")
	fmt.Fprintf(w, "}")
	fmt.Fprintf(w, "setInterval(updateStatus, 5000);")
	fmt.Fprintf(w, "</script>")
}

// Handler to check the liveness of the application
func livenessHealthHandler(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&simulateLivenessFailure) == 1 {
		http.Error(w, `{"status": "down"}`, http.StatusInternalServerError)
		return
	}
	// If the application is still starting up, return a 503
	if atomic.LoadInt32(&startupComplete) == 0 {
		http.Error(w, `{"status": "starting"}`, http.StatusServiceUnavailable)
		return
	}
	setProbeTimestamp("livenessProbe") // Capture liveness timestamp
	fmt.Fprintf(w, `{"status": "up"}`)
}

// Handler to check the readiness of the application
func readinessHealthHandler(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&simulateReadinessFailure) == 1 {
		http.Error(w, `{"status": "not ready"}`, http.StatusServiceUnavailable)
		return
	}
	if atomic.LoadInt32(&startupComplete) == 0 {
		http.Error(w, `{"status": "not ready"}`, http.StatusServiceUnavailable)
		return
	}
	fmt.Fprintf(w, `{"status": "ready"}`)
}

func toggleLivenessFailureHandler(w http.ResponseWriter, r *http.Request) {
	toggleFailure(&simulateLivenessFailure, w)
}

func toggleReadinessFailureHandler(w http.ResponseWriter, r *http.Request) {
	toggleFailure(&simulateReadinessFailure, w)
}

func toggleFailure(failureFlag *int32, w http.ResponseWriter) {
	if atomic.LoadInt32(failureFlag) == 0 {
		atomic.StoreInt32(failureFlag, 1)
		fmt.Fprintf(w, "Simulated failure mode activated.")
	} else {
		atomic.StoreInt32(failureFlag, 0)
		fmt.Fprintf(w, "Simulated failure mode deactivated.")
	}
}

// Handler to set and return the timestamps
func timestampsHandler(w http.ResponseWriter, r *http.Request) {
	probeTimestamps.Lock()
	defer probeTimestamps.Unlock()

	// Create a new map to hold the formatted timestamps
	formattedTimestamps := make(map[string]string)
	for key, value := range timestamps {
		// Format each timestamp to "YYYY-MM-DDTHH:MM:SS"
		formattedTimestamps[key] = value.Format("2006-01-02T15:04:05")
	}

	response, err := json.Marshal(formattedTimestamps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// Helper function to set timestamp for a given probe
func setProbeTimestamp(name string) {
	probeTimestamps.Lock()
	defer probeTimestamps.Unlock()
	// Only set the timestamp the first time it's called for each probe
	if _, exists := timestamps[name]; !exists {
		timestamps[name] = time.Now()
	}
}

// Helper function to determine if the toggle switch should be checked
func getToggleChecked(failureFlag *int32) string {
	if atomic.LoadInt32(failureFlag) == 1 {
		return "checked"
	}
	return ""
}

// Helper function to determine the status indicator text
func getStatusIndicator(failureFlag *int32, probeType string) string {
	if atomic.LoadInt32(failureFlag) == 1 {
		return "down"
	}
	if probeType == "readiness" && atomic.LoadInt32(&startupComplete) == 0 {
		return "not ready"
	}
	return "up"
}

func main() {
	go simulateLongStartup(6) // Simulate a long startup time

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/liveness-health", livenessHealthHandler)
	http.HandleFunc("/readiness-health", readinessHealthHandler)
	http.HandleFunc("/toggle-liveness-failure", toggleLivenessFailureHandler)
	http.HandleFunc("/toggle-readiness-failure", toggleReadinessFailureHandler)
	http.HandleFunc("/timestamps", timestampsHandler)

	fmt.Println("Starting server at port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
