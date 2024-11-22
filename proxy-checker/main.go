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
	HTTP_PROXY_LIST_URL  = "https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/http.txt"
	SOCKS4_PROXY_LIST_URL = "https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/socks4.txt"
	SOCKS5_PROXY_LIST_URL = "https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/socks5.txt"

	// Working Proxy Filenames
	HTTP_WORKING_PROXY_FILENAME  = "http_working_proxies.txt"
	SOCKS4_WORKING_PROXY_FILENAME = "socks4_working_proxies.txt"
	SOCKS5_WORKING_PROXY_FILENAME = "socks5_working_proxies.txt"
)

func main() {
	// Create a Resty Client
	client := resty.New()

	// Create a request to fetch HTTP proxy list
	httpProxyResponse, err := client.R().Get(HTTP_PROXY_LIST_URL)
	if err != nil || httpProxyResponse.StatusCode() != http.StatusOK {
		log.Println("Getting HTTP Proxy List failed:", err, httpProxyResponse.StatusCode())
	}

	// Create a request to fetch SOCKS4 proxy list
	socks4ProxyResponse, err := client.R().Get(SOCKS4_PROXY_LIST_URL)
	if err != nil || socks4ProxyResponse.StatusCode() != http.StatusOK {
		log.Println("Getting SOCKS4 Proxy List failed:", err, socks4ProxyResponse.StatusCode())
	}

	// Create a request to fetch SOCKS5 proxy list
	socks5ProxyResponse, err := client.R().Get(SOCKS5_PROXY_LIST_URL)
	if err != nil || socks5ProxyResponse.StatusCode() != http.StatusOK {
		log.Println("Getting SOCKS5 Proxy List failed:", err, socks5ProxyResponse.StatusCode())
	}

	// Split the proxy lists into arrays of strings, only if the responses were successful
	var httpProxies, socks4Proxies, socks5Proxies []string
	if httpProxyResponse.StatusCode() == http.StatusOK {
		httpProxies = strings.Split(string(httpProxyResponse.Body()), "\n")
	}
	if socks4ProxyResponse.StatusCode() == http.StatusOK {
		socks4Proxies = strings.Split(string(socks4ProxyResponse.Body()), "\n")
	}
	if socks5ProxyResponse.StatusCode() == http.StatusOK {
		socks5Proxies = strings.Split(string(socks5ProxyResponse.Body()), "\n")
	}

	// Remove existing proxy files if they exist
	removeFileIfExists(HTTP_WORKING_PROXY_FILENAME)
	removeFileIfExists(SOCKS4_WORKING_PROXY_FILENAME)
	removeFileIfExists(SOCKS5_WORKING_PROXY_FILENAME)

	// Create a WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Loop over the HTTP proxies and launch a goroutine for each one
	for _, httpProxyServer := range httpProxies {
		wg.Add(1) // Increment the WaitGroup counter

		go func(proxy string) {
			defer wg.Done() // Decrement the WaitGroup counter when this goroutine completes
			processProxy(proxy, "http://", HTTP_WORKING_PROXY_FILENAME)
		}(httpProxyServer)
	}

	// Loop over the SOCKS4 proxies and launch a goroutine for each one
	for _, socks4ProxyServer := range socks4Proxies {
		wg.Add(1) // Increment the WaitGroup counter

		go func(proxy string) {
			defer wg.Done() // Decrement the WaitGroup counter when this goroutine completes
			processProxy(proxy, "socks4://", SOCKS4_WORKING_PROXY_FILENAME)
		}(socks4ProxyServer)
	}

	// Loop over the SOCKS5 proxies and launch a goroutine for each one
	for _, socks5ProxyServer := range socks5Proxies {
		wg.Add(1) // Increment the WaitGroup counter

		go func(proxy string) {
			defer wg.Done() // Decrement the WaitGroup counter when this goroutine completes
			processProxy(proxy, "socks5://", SOCKS5_WORKING_PROXY_FILENAME)
		}(socks5ProxyServer)
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
		fmt.Println("Failed to open proxy file:", proxyFilename)
		return
	}
	defer f.Close()

	// Write the proxy server to the file
	if _, err := f.WriteString(proxy + "\n"); err != nil {
		fmt.Println("Failed to write to file:", proxyFilename, "with error:", err)
	}
}

// Helper function to remove a file if it exists
func removeFileIfExists(filename string) {
	if _, err := os.Stat(filename); err == nil {
		fmt.Println("Removing", filename)
		if err := os.Remove(filename); err != nil {
			fmt.Println("Failed to remove", filename)
		}
	}
}
