package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

// getDomainFromHost takes a hostname string in the form "domain.com:port" and returns just the domain name.
func getDomainFromHost(host string) string {

	// split the hostname from the port by chopping at :
	parts := strings.Split(host, ":")

	// if we have example.tld:443, then we're looking for the example.tld part.
	return parts[0]

}

// domainHandler checks what domain is being requested, and routes the request appropriately
func domainHandler(w http.ResponseWriter, r *http.Request) {

	// get the domain being requested
	domain := getDomainFromHost(r.Host)

	// check if we have a directory for it
	path := filepath.Join("./www", domain)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {

			/*

				custom 404's, a short story.

				The way we do 404's are in order of importance. What this means is if there is a custom 404.html in eclaire/www/donutblog.com/ then we'll use that first. If there isn't, we'll move up one directory and see if there is a generic 404.html in eclaire/www/. If there isn't one there either, we just serve the generic http 404 response.

				We do also support 500 internal server errors, which should be rare, so those didn't get any special treatment, just an http 500 response when things go south.

			*/

			// this is the site-specific custom 404
			_, err := os.Stat(filepath.Join("./www", domain, "404.html"))
			if err == nil {
				http.ServeFile(w, r, filepath.Join("./www", domain, "404.html"))
				return
			}
			// this is the 404 that comes with the server
			_, err = os.Stat("./www/404.html")
			if err == nil {
				http.ServeFile(w, r, "./www/404.html")
				return
			}

			// this is the value-brand 404 page
			http.Error(w, "404 page not found", http.StatusNotFound)
			return
		}

		// if anything goes wrong, just run around screaming
		// http 500 error
		// it didn't make sense to make anything elaborate bc
		// how much can really go wrong with a static site?

		// ...

		// as soon as those words came off the fingertips,
		// there was regret and a sense of foreshadowing.
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
		return
	}

	// server up the content we -can- find
	filePath := filepath.Join(path, r.URL.Path)
	_, err = os.Stat(filePath)
	if err != nil {

		if os.IsNotExist(err) {

			// this is the site-specific custom 404
			_, err := os.Stat(filepath.Join(path, "404.html"))
			if err == nil {
				http.ServeFile(w, r, filepath.Join(path, "404.html"))
				return
			}

			// serve the default 404 page if it exists
			_, err = os.Stat("./www/404.html")
			if err == nil {
				http.ServeFile(w, r, "./www/404.html")
				return
			}

			// serve the complimentary http 404 response
			http.Error(w, "404 page not found", http.StatusNotFound)
			return
		}

		//
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
		return
	}

	// let it rip
	http.FileServer(http.Dir(path)).ServeHTTP(w, r)

}

func setupServer() error {
	// Create the www directory if it doesn't exist
	_, err := os.Stat("./www")
	if os.IsNotExist(err) {
		err := os.Mkdir("./www", 0755)
		if err != nil {
			return err
		}
		// Create a simple demo HTML page for Eclaire
		html := []byte("<html><body><h1>Eclaire is working!</h1></body></html>")
		err = os.WriteFile("./www/index.html", html, 0644)
		if err != nil {
			return err
		}
		// Create a custom HTML page for Eclaire
		html404 := []byte("<html><body><h1>404</h1><h4>call the cops</h4></body></html>")
		err = os.WriteFile("./www/404.html", html404, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {

	// initialize the server
	err := setupServer()
	if err != nil {
		fmt.Println("Error setting up server:", err)
		return
	}

}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", domainHandler)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Set("Content-Encoding", "gzip")
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache("certs"),
	}

	server := &http.Server{
		Addr:         ":https",
		Handler:      handler,
		TLSConfig:    certManager.TLSConfig(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Println("Starting HTTP server on port 80...")
		err := http.ListenAndServe(":http", certManager.HTTPHandler(nil))
		if err != nil {
			fmt.Println("Error starting HTTP server:", err)
		}
	}()

	fmt.Println("Starting HTTPS server on port 443...")
	err := server.ListenAndServeTLS("", "")
	if err != nil {
		fmt.Println("Error starting HTTPS server:", err)
	}
}
