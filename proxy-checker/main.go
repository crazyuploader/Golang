package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	// Proxy Server List URL
	HTTP_PROXY_LIST_URL  = "https://www.proxyscan.io/download?type=http"
	HTTPS_PROXY_LIST_URL = "https://www.proxyscan.io/download?type=https"

	// Working Proxy Filename
	WorkingProxyFilename = "working_proxies.txt"
)

func main() {
	// Create a Resty Client
	client := resty.New()

	// Create a request
	httpProxyResponse, _ := client.R().Get(HTTP_PROXY_LIST_URL)
	httpsProxyResponse, _ := client.R().Get(HTTPS_PROXY_LIST_URL)

	// Check if the HTTP Proxy List response has errors
	if httpProxyResponse.StatusCode() != http.StatusOK {
		log.Fatalln("Getting HTTP Proxy List failed:", httpProxyResponse.StatusCode())
	}

	// Check if the HTTPS Proxy List response has errors
	if httpsProxyResponse.StatusCode() != http.StatusOK {
		log.Fatalln("Getting HTTPS Proxy List failed:", httpsProxyResponse.StatusCode())
	}

	// Split the proxy lists into arrays of strings
	httpProxies := strings.Split(string(httpProxyResponse.Body()), "\n")
	httpsProxies := strings.Split(string(httpsProxyResponse.Body()), "\n")

	// Remove working proxies file if it exists
	if _, err := os.Stat(WorkingProxyFilename); err == nil {
		fmt.Println("Removing", WorkingProxyFilename)
		fmt.Println("")
		if err := os.Remove(WorkingProxyFilename); err != nil {
			fmt.Println("Failed to remove", WorkingProxyFilename)
		}
	}

	// Create a WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Loop over the HTTP proxies and launch a goroutine for each one
	for _, httpProxyServer := range httpProxies {
		wg.Add(1) // Increment the WaitGroup counter

		go func(proxy string) {
			defer wg.Done() // Decrement the WaitGroup counter when this goroutine completes
			processProxy(proxy, "http://", WorkingProxyFilename)
		}(httpProxyServer)
	}

	// Loop over the HTTPS proxies and launch a goroutine for each one
	for _, httpsProxyServer := range httpsProxies {
		wg.Add(1) // Increment the WaitGroup counter

		go func(proxy string) {
			defer wg.Done() // Decrement the WaitGroup counter when this goroutine completes
			processProxy(proxy, "https://", WorkingProxyFilename)
		}(httpsProxyServer)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	fmt.Println("All requests completed.")
}

// Function to check whether or not a proxy works properly
func processProxy(proxy, prefix, proxyFilename string) {
	// Create a new Resty client for each request
	client := resty.New()

	// Set the proxy and timeout settings for the client
	client.SetProxy(fmt.Sprintf("%s%s", prefix, proxy))
	client.SetTimeout(10 * time.Second)

	// Make the HTTP request using the client
	resp, err := client.R().Get("https://api.devjugal.com/ip")
	if err != nil {
		return // Skip this request if there was an error
	}

	// Print the response if the request was successful
	fmt.Println("Proxy Server:", proxy)
	fmt.Println(resp)
	fmt.Println("")

	// Open the working proxies list file
	f, err := os.OpenFile(proxyFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		fmt.Println("Failed to open proxy.txt file")
	}
	defer f.Close()

	// Write the proxy server to the file
	if _, err := f.WriteString(proxy + "\n"); err != nil {
		fmt.Println("Failed to write proxy.txt file, with error:", err)
	}
}
