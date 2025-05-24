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
	"strconv"
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

// Function to simulate a long startup time with actual computation
func simulateLongStartup(limit int) {
	start := time.Now()
	count := 0
	// log out limit to the console
	fmt.Printf("Limit: %d\n", limit)

	// Naive prime number calculation to consume CPU
	for num := 2; count < limit; num++ {
		if isPrime(num) {
			count++
			// print out the count to the console
			fmt.Printf("Count: %d\n", count)
		}
		if time.Since(start) > 60*time.Second { // Ensure we run for about 1 minute
			break
		}
	}

	// Mark startup as complete
	atomic.StoreInt32(&startupComplete, 1)
	setProbeTimestamp("startupProbe")

	// Create a file to indicate that the startup is complete
	_, err := os.Create("/tmp/startup-file")
	// write to the file and date
	err = os.WriteFile("/tmp/startup-file", []byte(`Startup complete at ` + time.Now().Format("2006-01-02T15:04:05")), 0644)
	if err != nil {
		log.Fatalf("Failed to create startup complete file: %v", err)
	}
	
}

// Helper function to check if a number is prime
func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Valkyrie Application</h1>")
	fmt.Fprintf(w, "<br>")
	fmt.Fprintf(w, "<p>Welcome to the Valkyrie application. The application is used to trigger Kubernetes liveness, readiness, and startup probes.</p>")
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
	fmt.Fprintf(w, "<strong>Liveness Status: </strong><span id='liveness-indicator' class='status-starting'>starting</span>")
	fmt.Fprintf(w, "<br><strong>Readiness Status: </strong><span id='readiness-indicator' class='status-starting'>not ready</span>")

	// Add CSS for colors
	fmt.Fprintf(w, "<style>")
	fmt.Fprintf(w, ".status-up { color: green; font-weight: bold; }")
	fmt.Fprintf(w, ".status-down { color: red; font-weight: bold; }")
	fmt.Fprintf(w, ".status-starting { color: orange; font-weight: bold; }")
	fmt.Fprintf(w, "</style>")

	// JavaScript
	fmt.Fprintf(w, "<script>")
	fmt.Fprintf(w, "function toggleLivenessFailure() { fetch('/toggle-liveness-failure').then(response => response.text()).then(data => { console.log(data); updateStatus(); }); }")
	fmt.Fprintf(w, "function toggleReadinessFailure() { fetch('/toggle-readiness-failure').then(response => response.text()).then(data => { console.log(data); updateStatus(); }); }")

	// Function to update the status dynamically
	fmt.Fprintf(w, "function updateStatus() {")

	// Fetch Liveness Status
	fmt.Fprintf(w, "fetch('/liveness-health').then(response => {")
	fmt.Fprintf(w, "    var livenessElement = document.getElementById('liveness-indicator');")
	fmt.Fprintf(w, "    if (!response.ok) {") 
	fmt.Fprintf(w, "        livenessElement.innerText = 'down';")
	fmt.Fprintf(w, "        livenessElement.classList.remove('status-up', 'status-starting');") 
	fmt.Fprintf(w, "        livenessElement.classList.add('status-down');") 
	fmt.Fprintf(w, "        return;") 
	fmt.Fprintf(w, "    }") 
	fmt.Fprintf(w, "    return response.text();")
	fmt.Fprintf(w, "}).then(data => {")
	fmt.Fprintf(w, "    if (data) {")
	fmt.Fprintf(w, "        var livenessElement = document.getElementById('liveness-indicator');")
	fmt.Fprintf(w, "        livenessElement.innerText = data;")
	fmt.Fprintf(w, "        livenessElement.classList.remove('status-down', 'status-starting');")
	fmt.Fprintf(w, "        if (data === 'up') { livenessElement.classList.add('status-up'); }")
	fmt.Fprintf(w, "        else { livenessElement.classList.add('status-starting'); }")
	fmt.Fprintf(w, "    }")
	fmt.Fprintf(w, "});")

	// Fetch Readiness Status
	fmt.Fprintf(w, "fetch('/readiness-health').then(response => {")
	fmt.Fprintf(w, "    var readinessElement = document.getElementById('readiness-indicator');")
	fmt.Fprintf(w, "    if (!response.ok) {")
	fmt.Fprintf(w, "        readinessElement.innerText = 'not ready';")
	fmt.Fprintf(w, "        readinessElement.classList.remove('status-up', 'status-starting');") 
	fmt.Fprintf(w, "        readinessElement.classList.add('status-down');") 
	fmt.Fprintf(w, "        return;") 
	fmt.Fprintf(w, "    }")
	fmt.Fprintf(w, "    return response.text();")
	fmt.Fprintf(w, "}).then(data => {")
	fmt.Fprintf(w, "    if (data) {")
	fmt.Fprintf(w, "        var readinessElement = document.getElementById('readiness-indicator');")
	fmt.Fprintf(w, "        readinessElement.innerText = data;")
	fmt.Fprintf(w, "        readinessElement.classList.remove('status-down', 'status-starting');")
	fmt.Fprintf(w, "        if (data === 'ready') { readinessElement.classList.add('status-up'); }")
	fmt.Fprintf(w, "        else { readinessElement.classList.add('status-starting'); }")
	fmt.Fprintf(w, "    }")
	fmt.Fprintf(w, "});")

	fmt.Fprintf(w, "}")

	// Automatically update status every 2 seconds
	fmt.Fprintf(w, "setInterval(updateStatus, 2000);")
	fmt.Fprintf(w, "</script>")
}
// Handler to check the liveness of the application
func livenessHealthHandler(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&simulateLivenessFailure) == 1 {
		// sleep for 2 seconds to simulate a slow response
		time.Sleep(2 * time.Second)
		http.Error(w, `down`, http.StatusInternalServerError)
		return
	}
	// If the application is still starting up, return a 503
	if atomic.LoadInt32(&startupComplete) == 0 {
		http.Error(w, `down`, http.StatusServiceUnavailable)
		return
	}
	// If the application has started up, return a 200
	setProbeTimestamp("livenessProbe") // Capture liveness timestamp
	// return 200
	fmt.Fprintf(w, `up`)
}

// Handler to check the readiness of the application
func readinessHealthHandler(w http.ResponseWriter, r *http.Request) {
	// sleep for 2 seconds to simulate a slow response
	if atomic.LoadInt32(&simulateReadinessFailure) == 1 {
		http.Error(w, `not ready`, http.StatusServiceUnavailable)
		return
	}
	if atomic.LoadInt32(&startupComplete) == 0 {
		http.Error(w, `not ready`, http.StatusServiceUnavailable)
		return
	}
	fmt.Fprintf(w, `ready`)
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
	primeTimeCountStr := os.Getenv("PRIME_NUMBER_COUNT") // this comes from the deployment.yaml file
    
	// Convert the string to an integer
	primeTimeCount, err := strconv.Atoi(primeTimeCountStr)
	if err != nil {
		log.Fatalf("Error converting PRIME_NUMBER_COUNT to integer: %v", err)
	}
    
	// Start the long startup process in a separate goroutine
	go simulateLongStartup(primeTimeCount)

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
