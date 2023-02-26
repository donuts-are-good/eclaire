package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/crypto/blake2b"
)

// init - we're mainly just setting up directory and file
// boilerplate items here that should run before main()
func init() {

	// initialize the server
	err := setupServer()
	if err != nil {
		log.Println("Error setting up server:", err)
		return
	}

}

// setupServer scans for and/or creates necessary boilerplate directories and the files that go in them.
// if www doesnt exist, it will be created along with the index and 404.html
// if www does exist, but those files are missing, it assumes you wanted it
// that way and leaves them alone.
func setupServer() error {

	// ./www is where all of the content you're hosting lives.
	// inside www is where you put folders matching the names of your domains.

	/*

		sidenote, conjecture, errata...
		why apache and nginx dont work like this, the world may never know,
		but this format always made the most sense.
		miss me with that symlinked config stuff.

		if there is no www folder, the user won't know to put anything there, so we should create it if it doesn't exist.

	*/

	// check for the www folder
	_, err := os.Stat("./www")
	if os.IsNotExist(err) {

		// do the actual making of the folder
		err := os.Mkdir("./www", 0755)
		if err != nil {
			return err
		}

		// nginx and apache always had a sweet it works page
		// that would tell attackers all the sweet details about
		// all the fun bits your server has hot and ready for them.
		// we should do like the cool kids do.
		html := []byte("<html><body><h1>Eclaire is working!</h1></body></html>")
		err = os.WriteFile("./www/index.html", html, 0644)
		if err != nil {
			return err
		}
		// custom 404's were pretty hot back in the 88x31 days,
		// and for whatever reason the time just was never available
		// to set them up always, so let's put this here but more as
		// an indicator, rather than a prompt to be artistic.
		// but that'd be cool.
		html404 := []byte("<html><body><h1>404</h1><h4>call the cops</h4></body></html>")

		// write the file, we dont use ioutil anymore for writing
		// in modern Go stuff because it's deprecated, so here we're
		// basically doing the same thing with os instead of io.
		err = os.WriteFile("./www/404.html", html404, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

// main is where all of the magic happens.
func main() {

	// new muxer
	mux := http.NewServeMux()

	// send requests to the handler
	mux.HandleFunc("/", domainHandler)

	// a cached request object
	type cachedItem struct {
		resp    *http.Response
		content []byte
	}

	// a poor man's makeshift cache
	cache := make(map[string]*cachedItem)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// debug: log the request
		logRequest(r)

		// for each request, we hash the path
		pathHash := hashPath(r.URL.Path)

		// for each entry in the cache, we do some validation
		cacheEntry, ok := cache[pathHash]

		if ok {

			// in our cachedItem, we're essentially looking at a http.Response
			// pointer with some extra stuff attached to it.
			// that response has header data, and in that header data, we're
			// storing last-modified as a string which can be parsed into time.
			// this was mainly done as an assumption it'd be less intensive than
			// calculating the hash then calculating modtime on the file itself
			// on disk for each request if a lookup will suffice.
			lastModifiedStr := cacheEntry.resp.Header.Get("Last-Modified")
			lastModifiedTime, err := time.Parse(http.TimeFormat, lastModifiedStr)
			if err != nil {
				log.Println("error checking modtime: ", err)
			}

			// here we are checking if the cache needs to be invalidated
			if file, err := os.Open(filepath.Join("www", r.URL.Path)); err == nil {
				defer file.Close()

				// stat is used to get modTime, which we compare to
				// After(lastModifiedTime)
				stat, err := file.Stat()
				if err == nil && stat.ModTime().After(lastModifiedTime) {

					// always knew there was a delete keyword but hadn't had a
					// whole lot of chances to use it where it made sense...
					// today is your day, my friend.
					delete(cache, pathHash)

				}
			}
		}

		// get the file that was asked for and server it to the user
		// unless there's a problem.
		transport := http.DefaultTransport
		client := &http.Client{Transport: transport}
		resp, err := client.Do(r)
		if err != nil {
			http.Error(w, "500 internal server error", http.StatusInternalServerError)
			return
		}

		// dont forget to shut the fridge.
		defer resp.Body.Close()

		// use io readall for the response body
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "500 internal server error", http.StatusInternalServerError)
			return
		}

		// set headers
		cacheEntry = &cachedItem{resp: resp, content: content}
		cacheEntry.resp.Header.Set("Cache-Control", "public, max-age=86400")
		cacheEntry.resp.Header.Set("Vary", "Accept-Encoding")
		cacheEntry.resp.Header.Set("Content-Encoding", "gzip")
		cacheEntry.resp.Header.Set("Last-Modified", time.Now().Format(http.TimeFormat))

		// at to our 'cache'
		cache[pathHash] = cacheEntry

		// range the header for the kv data
		for headerKey, headerKeyValue := range resp.Header {
			w.Header().Set(headerKey, headerKeyValue[0])
		}

		// write the header response code
		w.WriteHeader(resp.StatusCode)

		// send it, brah
		w.Write(content)

	})

	// certManager is for autocert letsencrypt
	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache("certs"),
	}

	// https server params
	// maybe these need to be tuned. not sure until we run the test
	// https://github.com/donuts-are-good/knockknock
	server := &http.Server{
		Addr:         ":https",
		Handler:      handler,
		TLSConfig:    certManager.TLSConfig(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// http(80) https(443)
	go func() {

		// http port server
		log.Println("Starting HTTP server on port 80...")
		err := http.ListenAndServe(":http", certManager.HTTPHandler(nil))
		if err != nil {
			log.Println("Error starting HTTP server:", err)
		}
	}()

	// https port server
	log.Println("Starting HTTPS server on port 443...")
	err := server.ListenAndServeTLS("", "")
	if err != nil {
		log.Println("Error starting HTTPS server:", err)
	}
}

// hashPath is part of the caching strategy
// blake2b was used just because it's fast
// and reasonably strong for a static site
// and it is present in the Go stdlib
func hashPath(path string) string {

	// make a new hash object
	h, _ := blake2b.New256(nil)

	// add the path to the hash object
	h.Write([]byte(path))

	// hash it
	hash := h.Sum(nil)

	// return the hash as hex
	return hex.EncodeToString(hash)

}

// splitDomainFromPort takes a hostname string in the form "domain.com:port" and returns just the domain name.
func splitDomainFromPort(host string) string {

	// split the hostname from the port by chopping at :
	parts := strings.Split(host, ":")

	// if we have example.tld:443, then we're looking for the example.tld part.
	return parts[0]

}

// logRequest logs the specified request to the console
func logRequest(r *http.Request) {
	fmt.Printf("[%s] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), r.Method, r.URL.Path)
}

// domainHandler checks what domain is being requested, and routes the request appropriately
func domainHandler(w http.ResponseWriter, r *http.Request) {

	// log the request
	logRequest(r)

	// get the domain being requested
	domain := splitDomainFromPort(r.Host)

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
