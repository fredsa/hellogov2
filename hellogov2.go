package main

import (
	"fmt"
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
	// Running locally?
	if os.Getenv(GAE_APPLICATION) == "" {
		runningLocally = true

		// Returned by `appengine.AppID(ctx)`.
		_ = os.Setenv(GAE_APPLICATION, "myappid")

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
	appengine.Main()
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	// App Engine context for the in-flight HTTP request.
	ctx := appengine.NewContext(r)

	// Running locally => nil
	// Server inadvertently imported "google.golang.org/appengine/user" => nil
	// Server correctly imported "google.golang.org/appengine/v2/user" => `fredsa` (=the logged in user)
	fmt.Fprintf(w, "user.Current(ctx)=%v\n", user.Current(ctx))

	// Running locally => false
	// Server inadvertently imported "google.golang.org/appengine/user" => false
	// Server correctly imported "google.golang.org/appengine/v2/user" => true
	fmt.Fprintf(w, "user.IsAdmin(ctx)=%v\n", user.IsAdmin(ctx))

	if !runningLocally {
		// Running locally => err `service bridge HTTP failed: Post "http://appengine.googleapis.internal:10001/rpc_http": dial tcp: lookup appengine.googleapis.internal: no such host`
		// Server inadvertently imported "google.golang.org/appengine/user" => err `API error 2 (user: NOT_ALLOWED)`
		// Server correctly imported "google.golang.org/appengine/v2/user" => `https://myappid.appspot.com/_ah/conflogin?continue=https://myappid.appspot.com/`
		loginURL, err := user.LoginURL(ctx, "/")
		fmt.Fprintf(w, "user.LoginURL(ctx, \"/\")=%v err=%v\n", loginURL, err)
	}

	if !runningLocally {
		// Running locally => err `service bridge HTTP failed: Post "http://appengine.googleapis.internal:10001/rpc_http": dial tcp: lookup appengine.googleapis.internal: no such host`
		// Server inadvertently imported "google.golang.org/appengine/user" => err `API error 2 (user: NOT_ALLOWED)`
		// Server correctly imported "google.golang.org/appengine/v2/user" => `https://myappid.appspot.com/_ah/conflogin?continue=https://myappid.appspot.com/`
		logoutURL, err := user.LoginURL(ctx, "/")
		fmt.Fprintf(w, "user.LogoutURL(ctx, \"/\")=%v err=%v\n", logoutURL, err)
	}

	// Running locally => value `os.Getenv("GAE_APPLICATION")`
	// Server, `myappid`
	fmt.Fprintf(w, "appengine.AppID(ctx)=%v\n", appengine.AppID(ctx))

	if !runningLocally {
		// Running locally => `` and logs `Get "http://metadata/computeMetadata/v1/instance/zone": dial tcp: lookup metadata: no such host`
		// Server, `us-west1-1`
		fmt.Fprintf(w, "appengine.Datacenter(ctx)=%v\n", appengine.Datacenter(ctx))
	}

	if !runningLocally {
		// Running locally => ``
		// Server, `myappid.uw.r.appspot.com`
		fmt.Fprintf(w, "appengine.DefaultVersionHostname(ctx)=%v\n", appengine.DefaultVersionHostname(ctx))
	}

	if !runningLocally {
		// Running locally => panics `http: panic serving [::1]:64902: Metadata fetch failed for 'instance/attributes/gae_backend_instance': Get "http://metadata/computeMetadata/v1/instance/attributes/gae_backend_instance": dial tcp: lookup metadata: no such host`
		// Server, `0066d924808f85e59480f4f834d89809739e28d68d0471e54a81ecfdd776e886ec44b9294369e983466180a7ef8a9dd24967183bc382a400c8a9d3f8f483ef2fac91a655321eeb3743b798cf97ca`
		fmt.Fprintf(w, "appengine.InstanceID()=%v\n", appengine.InstanceID())
	}

	// Running locally => true
	// Server, true
	fmt.Fprintf(w, "appengine.IsAppEngine()=%v\n", appengine.IsAppEngine())

	// Running locally => false
	// Server, false
	fmt.Fprintf(w, "appengine.IsDevAppServer()=%v\n", appengine.IsDevAppServer())

	// Running locally => false
	// Server, false
	fmt.Fprintf(w, "appengine.IsFlex()=%v\n", appengine.IsFlex())

	// Running locally => true
	// Server, true
	fmt.Fprintf(w, "appengine.IsSecondGen()=%v\n", appengine.IsSecondGen())

	// Running locally => true
	// Server, true
	fmt.Fprintf(w, "appengine.IsStandard()=%v\n", appengine.IsStandard())

	if !runningLocally {
		// Running locally => panics `http: panic serving [::1]:65140: Metadata fetch failed for 'instance/attributes/gae_backend_name': Get "http://metadata/computeMetadata/v1/instance/attributes/gae_backend_name": dial tcp: lookup metadata: no such host`
		// Server, `default`
		fmt.Fprintf(w, "appengine.ModuleName(ctx)=%v\n", appengine.ModuleName(ctx))
	}

	// Running locally => “
	// Server, `677455b700ff0cb62a9b86bdbb00017a75777e73617565722d7064612d6465760001323032343132333174313232383135000100`
	fmt.Fprintf(w, "appengine.RequestID(ctx)=%v\n", appengine.RequestID(ctx))

	// Running locally => `standard`
	// Server, `standard`
	fmt.Fprintf(w, "appengine.ServerSoftware()=%v\n", appengine.ServerSoftware())

	if !runningLocally {
		// Running locally => err `service bridge HTTP failed: Post "http://appengine.googleapis.internal:10001/rpc_http": dial tcp: lookup appengine.googleapis.internal: no such host`
		// Server, `myappid@appspot.gserviceaccount.com`
		serviceAccount, err := appengine.ServiceAccount(ctx)
		fmt.Fprintf(w, "appengine.ServiceAccount(ctx)=%v err=%v\n", serviceAccount, err)
	}

	if !runningLocally {
		// Running locally => panics `http: panic serving [::1]:49339: Metadata fetch failed for 'instance/attributes/gae_backend_version': Get "http://metadata/computeMetadata/v1/instance/attributes/gae_backend_version": dial tcp: lookup metadata: no such host`
		// Server, `20241231t122815.465917320654064374`
		fmt.Fprintf(w, "appengine.VersionID(ctx)=%v\n", appengine.VersionID(ctx))
	}
}
