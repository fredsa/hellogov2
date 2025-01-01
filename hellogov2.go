package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"google.golang.org/appengine/v2"
	"google.golang.org/appengine/v2/user"
)

// https://cloud.google.com/appengine/docs/standard/go/runtime#environment_variables
const PORT = "PORT"
const GOOGLE_CLOUD_PROJECT = "GOOGLE_CLOUD_PROJECT" // The Google Cloud project ID associated with your application.
const GAE_APPLICATION = "GAE_APPLICATION"           // App id, with prefix.
const GAE_ENV = "GAE_ENV"                           // `standard` in production.
const GAE_RUNTIME = "GAE_RUNTIME"                   // Runtime in `app.yaml`.
const GAE_VERSION = "GAE_VERSION"                   // App version.
const DUMMY_APP_ID = "my-app-id"

var runningLocally = false

func init() {
	// Register handlers in init() per `appengine.Main()` documentation.
	http.HandleFunc("/", myHandler)
}

func main() {
	if os.Getenv(GAE_APPLICATION) == "" {
		runningLocally = true

		// Returned by `appengine.AppID(ctx)`.
		_ = os.Setenv(GAE_APPLICATION, "hellogov2")

		// Runtime from `app.yaml`.
		_ = os.Setenv(GAE_RUNTIME, "go123")

		// Deployed version.
		_ = os.Setenv(GAE_VERSION, "my-version")

		// App Engine standard environemnt.
		_ = os.Setenv(GAE_ENV, "standard")

		// Optionally, override default port 8080.
		_ = os.Setenv(PORT, "4200")

		fmt.Printf("Calling appengine.Main() to listen on port %s", os.Getenv(PORT))
	}

	// Standard App Engine APIs require `appengine.Main` to have been called.
	// appengine.Main()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	// App Engine context for the in-flight HTTP request.
	ctx := appengine.NewContext(r)

	fmt.Fprintf(w, "\nUser APIs:\n")
	checkUserAPIs(w, ctx)

	fmt.Fprintf(w, "\nApp Engine APIs:\n")
	checkAppEngineAPIs(w, ctx)
}

func checkUserAPIs(w http.ResponseWriter, ctx context.Context) {
	// Running locally => nil
	// Server inadvertently imported "google.golang.org/appengine/user" => nil
	// Server correctly imported "google.golang.org/appengine/v2/user" => `fredsa` (=the logged in user)
	// Server ListenAndServe() => nil
	fmt.Fprintf(w, "user.Current(ctx)=%v\n", user.Current(ctx))

	// Running locally => (always) false
	// Server inadvertently imported "google.golang.org/appengine/user" => (always) false
	// Server correctly imported "google.golang.org/appengine/v2/user" => true
	// Server ListenAndServe() => (always) false
	fmt.Fprintf(w, "user.IsAdmin(ctx)=%v\n", user.IsAdmin(ctx))

	if !runningLocally {
		// Running locally => err `service bridge HTTP failed: Post "http://appengine.googleapis.internal:10001/rpc_http": dial tcp: lookup appengine.googleapis.internal: no such host`
		// Server inadvertently imported "google.golang.org/appengine/user" => err `API error 2 (user: NOT_ALLOWED)`
		// Server correctly imported "google.golang.org/appengine/v2/user" => `https://hellogov2.appspot.com/_ah/conflogin?continue=https://hellogov2.appspot.com/`
		// Server ListenAndServe() => err `API error 2 (user: NOT_ALLOWED)`
		loginURL, err := user.LoginURL(ctx, "/")
		fmt.Fprintf(w, "user.LoginURL(ctx, \"/\")=%v err=%v\n", loginURL, err)
	}

	if !runningLocally {
		// Running locally => err `service bridge HTTP failed: Post "http://appengine.googleapis.internal:10001/rpc_http": dial tcp: lookup appengine.googleapis.internal: no such host`
		// Server inadvertently imported "google.golang.org/appengine/user" => err `API error 2 (user: NOT_ALLOWED)`
		// Server correctly imported "google.golang.org/appengine/v2/user" => `https://hellogov2.appspot.com/_ah/conflogin?continue=https://hellogov2.appspot.com/`
		// Server ListenAndServe() => err `API error 2 (user: NOT_ALLOWED)`
		logoutURL, err := user.LoginURL(ctx, "/")
		fmt.Fprintf(w, "user.LogoutURL(ctx, \"/\")=%v err=%v\n", logoutURL, err)
	}
}

func checkAppEngineAPIs(w http.ResponseWriter, ctx context.Context) {
	// Running locally => value of `os.Getenv("GAE_APPLICATION")`
	// Server appengine.Main() => `hellogov2`
	// Server ListenAndServe() => `hellogov2`
	fmt.Fprintf(w, "appengine.AppID(ctx)=%v\n", appengine.AppID(ctx))

	if !runningLocally {
		// Running locally => `` and logs `Get "http://metadata/computeMetadata/v1/instance/zone": dial tcp: lookup metadata: no such host`
		// Server appengine.Main() => `us-west1-8`
		// Server ListenAndServe() => `us-west1-8`
		fmt.Fprintf(w, "appengine.Datacenter(ctx)=%v\n", appengine.Datacenter(ctx))
	}

	if !runningLocally {
		// Running locally => ""
		// Server appengine.Main() => "hellogov2.uw.r.appspot.com"
		// Server ListenAndServe() => ""
		fmt.Fprintf(w, "appengine.DefaultVersionHostname(ctx)=%v\n", appengine.DefaultVersionHostname(ctx))
	}

	if !runningLocally {
		// Running locally => panics `http: panic serving [::1]:64902: Metadata fetch failed for 'instance/attributes/gae_backend_instance': Get "http://metadata/computeMetadata/v1/instance/attributes/gae_backend_instance": dial tcp: lookup metadata: no such host`
		// Server appengine.Main() => `0066d924808f85e59480f4f834d89809739e28d68d0471e54a81ecfdd776e886ec44b9294369e983466180a7ef8a9dd24967183bc382a400c8a9d3f8f483ef2fac91a655321eeb3743b798cf97ca`
		// Server ListenAndServe() => `0066d92480c4c3a8bdfdfa6a133c0632b04135a3d6d71c6827c45af31aa60f2717e8d7ac036f652d3af9605884e802e97b84373919c769148bac28b145c9e756d6ee1abd6719c72ceaf233c27de2f1`
		fmt.Fprintf(w, "appengine.InstanceID()=%v\n", appengine.InstanceID())
	}

	// Running locally => true
	// Server appengine.Main() => true
	// Server ListenAndServe() => true
	fmt.Fprintf(w, "appengine.IsAppEngine()=%v\n", appengine.IsAppEngine())

	// Running locally => false
	// Server appengine.Main() => false
	// Server ListenAndServe() => false
	fmt.Fprintf(w, "appengine.IsDevAppServer()=%v\n", appengine.IsDevAppServer())

	// Running locally => false
	// Server appengine.Main() => false
	// Server ListenAndServe() => false
	fmt.Fprintf(w, "appengine.IsFlex()=%v\n", appengine.IsFlex())

	// Running locally => true
	// Server appengine.Main() => true
	// Server ListenAndServe() => true
	fmt.Fprintf(w, "appengine.IsSecondGen()=%v\n", appengine.IsSecondGen())

	// Running locally => true
	// Server appengine.Main() => true
	// Server ListenAndServe() => true
	fmt.Fprintf(w, "appengine.IsStandard()=%v\n", appengine.IsStandard())

	if !runningLocally {
		// Running locally => panics `http: panic serving [::1]:65140: Metadata fetch failed for 'instance/attributes/gae_backend_name': Get "http://metadata/computeMetadata/v1/instance/attributes/gae_backend_name": dial tcp: lookup metadata: no such host`
		// Server appengine.Main() => `default`
		// Server ListenAndServe() => `default`
		fmt.Fprintf(w, "appengine.ModuleName(ctx)=%v\n", appengine.ModuleName(ctx))
	}

	// Running locally => ""
	// Server appengine.Main() => "677455b700ff0cb62a9b86bdbb00017a75777e73617565722d7064612d6465760001323032343132333174313232383135000100"
	// Server ListenAndServe() => ""
	fmt.Fprintf(w, "appengine.RequestID(ctx)=%v\n", appengine.RequestID(ctx))

	// Running locally => "standard"
	// Server appengine.Main() => "standard"
	// Server ListenAndServe() => "standard"
	fmt.Fprintf(w, "appengine.ServerSoftware()=%v\n", appengine.ServerSoftware())

	if !runningLocally {
		// Running locally => err `service bridge HTTP failed: Post "http://appengine.googleapis.internal:10001/rpc_http": dial tcp: lookup appengine.googleapis.internal: no such host`
		// Server appengine.Main() => "hellogov2@appspot.gserviceaccount.com"
		// Server ListenAndServe() => "hellogov2@appspot.gserviceaccount.com"
		serviceAccount, err := appengine.ServiceAccount(ctx)
		fmt.Fprintf(w, "appengine.ServiceAccount(ctx)=%v err=%v\n", serviceAccount, err)
	}

	if !runningLocally {
		// Running locally => panics `http: panic serving [::1]:49339: Metadata fetch failed for 'instance/attributes/gae_backend_version': Get "http://metadata/computeMetadata/v1/instance/attributes/gae_backend_version": dial tcp: lookup metadata: no such host`
		// Server appengine.Main() => "20241231t122815.465917320654064374"
		// Server ListenAndServe() => "listenandserve.465923307189784710"
		fmt.Fprintf(w, "appengine.VersionID(ctx)=%v\n", appengine.VersionID(ctx))
	}
}
