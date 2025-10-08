package main

import (
	"log"
	"net/http"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// The main function is the entry point where the app is configured and started.
// It is executed in 2 different environments: A client (the web browser) and a
// server.
func main() {
	// The first thing to do is to associate the component with a path.
	//
	// This is done by calling the Route() function,  which tells go-app what
	// component to display for a given path, on both client and server-side.
	app.Route("/", func() app.Composer { return &mapLibre{} })

	// Once the routes set up, the next thing to do is to either launch the app
	// or the server that serves the app.
	//
	// When executed on the client-side, the RunWhenOnBrowser() function
	// launches the app,  starting a loop that listens for app events and
	// executes client instructions. Since it is a blocking call, the code below
	// it will never be executed.
	//
	// When executed on the server-side, RunWhenOnBrowser() does nothing, which
	// lets room for server implementation without the need for precompiling
	// instructions.
	app.RunWhenOnBrowser()

	// Finally, launching the server that serves the app is done by using the Go
	// standard HTTP package.
	//
	// The Handler is an HTTP handler that serves the client and all its
	// required resources to make it work into a web browser. Here it is
	// configured to handle requests with a path that starts with "/".
	http.Handle("/", &app.Handler{
		Name:        "Cyber Trespasser",
		Description: "Connect the dots",
		Styles: []string{
			"https://unpkg.com/leaflet@1.9.4/dist/leaflet.css",
			"https://unpkg.com/maplibre-gl/dist/maplibre-gl.css",
			"web/app.css",
		},
		Scripts: []string{
			"https://unpkg.com/leaflet@1.9.4/dist/leaflet.js",
			"https://unpkg.com/maplibre-gl/dist/maplibre-gl.js",
			"https://unpkg.com/@maplibre/maplibre-gl-leaflet/leaflet-maplibre-gl.js",
			"web/setupMap.js",
		},
	})

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}
