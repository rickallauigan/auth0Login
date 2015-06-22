package goauth2appengine

import (
	"appengine"
	"appengine/urlfetch"
	"code.google.com/p/goauth2/oauth"
	"html/template"
	"net/http"
)

// Cache all of the HTML files in the templates directory so that we only have to hit disk once.
var cached_templates = template.Must(template.ParseGlob("goauth2appengine/templates/*.html"))

// Global Variables used during OAuth protocol flow of authentication.
var (
	code  = ""
	token = ""
)

var domain = "bukidutility.auth0.com"

// OAuth2 configuration.
var oauthCfg = &oauth.Config{
	// ClientId:     "39841708487-lpttv3dlfg5u7ci5fm2du2me6l4hkj3v.apps.googleusercontent.com",
	// ClientSecret: "YNy5amFmFfUj9NcYGU5XRenx",
	// AuthURL:      "https://accounts.google.com/o/oauth2/auth",
	// RedirectURL:  "http://localhost:8080/oauth2callback",
	// TokenURL:     "https://accounts.google.com/o/oauth2/token",
	//Scope:        "https://www.googleapis.com/auth/userinfo.profile",

	ClientId:     "mhqhf8fTNZKtDDZRdukygwWTbybVVHbC",
	ClientSecret: "ZJzvq_3kGkmRjG3KOAQGweiLsvKytuRjKrB_l2cMcCUlo10oKpMNGzGQn8ns7B4E",
	AuthURL:      "https://bukidutility.auth0.com/authorize",
	RedirectURL:  "http://localhost:8080/oauth2callback",
	TokenURL:     "https://bukidutility.auth0.com/oauth/token",
	Scope:        "openid",
	//Endpoint: oauth2.Endpoint{
	//	AuthURL:  "https://" + domain + "/authorize",
	//	TokenURL: "https://" + domain + "/oauth/token",
	//},
}

// This is the URL that Google has defined so that an authenticated application may obtain the user's info in json format.
//const profileInfoURL = "https://www.googleapis.com/oauth2/v1/userinfo?alt=json"
const profileInfoURL = "https://bukidutility.auth0.com/userinfo"

// This is where Google App Engine sets up which handler lives at the root url.
func init() {
	// Immediately enter the main app.
	main()
}

func main() {

	// Setup application handlers.
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/authorize", handleAuthorize)

	// Google will redirect to this page to return your code, so handle it appropriately
	http.HandleFunc("/oauth2callback", handleOAuth2Callback)

}

// Root directory handler.
func handleRoot(rw http.ResponseWriter, req *http.Request) {

	err := cached_templates.ExecuteTemplate(rw, "notAuthenticated.html", nil)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
	}

}

// Start the authorization process.
func handleAuthorize(rw http.ResponseWriter, req *http.Request) {

	// Get the Google URL which shows the Authentication page to the user.
	url := oauthCfg.AuthCodeURL("")

	// Redirect user to that page.
	http.Redirect(rw, req, url, http.StatusFound)
}

// Function that handles the callback from the Google server.
func handleOAuth2Callback(rw http.ResponseWriter, req *http.Request) {

	// Initialize an appengine context.
	c := appengine.NewContext(req)

	// Retrieve the code from the response.
	code := req.FormValue("code")

	// Configure OAuth's http.Client to use the appengine/urlfetch transport
	// that all Google App Engine applications have to use for outbound requests.
	t := &oauth.Transport{Config: oauthCfg, Transport: &urlfetch.Transport{Context: c}}

	// Exchange the received code for a token.
	token, err := t.Exchange(code)
	if err != nil {
		c.Errorf("%v", err)
	}

	// Now get user data based on the Transport which has the token.
	resp, _ := t.Client().Get(profileInfoURL)
	buf := make([]byte, 1024)
	resp.Body.Read(buf)

	// Log the token.
	c.Infof("Token: %s", token)

	// Render the user's information.

	err = cached_templates.ExecuteTemplate(rw, "userInfo.html", string(buf))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
	}
}
