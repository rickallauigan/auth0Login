package main

import (
	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
	"code.google.com/p/goauth2/oauth"
	"crypto/sha256"
	_ "crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	//"github.com/dgrijalva/jwt-go"
	"github.com/astaxie/beego/session"
	"github.com/gorilla/mux"
	"html/template"
	//"io"
	"github.com/stretchr/objx"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type users struct {
	UID            string
	DateRegistered time.Time
	Token          string
}

type variables struct {
	ID           int       `json:"id"`
	VariableName string    `json:"variable name"`
	Description  string    `json:"description"`
	Unit         string    `json:"unit"`
	Validation   string    `json:"validation"`
	DateAdded    time.Time `json:"date added"`
}

type newVariables struct {
	ID           int    `json:"id"`
	VariableName string `json:"variable name"`
	Unit         string `json:"unit"`
}

type values struct {
	TimeCreated time.Time `json:"time created"`
	TimeStored  time.Time `json:"time stored"`
	DeviceUse   string    `json:"device use"`
	VariableID  string    `json:"variable id"`
	FarmID      string    `json:"farm id"`
	Token       string    `json:"token"`
	Value       string    `json:"value"`
}

type recipes struct {
	Recipeid    string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	RecipeVars  string    `json:"recipe variables"`
	DateCreated time.Time `json:"date created"`
}
type RecipeV map[string]string

type recipeVars struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	DataType string `json:"data type"`
}

// Cache all of the HTML files in the templates directory so that we only have to hit disk once.
var cached_templates = template.Must(template.ParseGlob("goauth2appengine/templates/*.html"))

// Global Variables used during OAuth protocol flow of authentication.
//var (
//	code  = ""
//  token = ""
//)

var (
	GlobalSessions *session.Manager
)
var domain = "bukidutility.auth0.com"

// OAuth2 configuration.
var oauthCfg = &oauth.Config{
	ClientId:     "mhqhf8fTNZKtDDZRdukygwWTbybVVHbC",
	ClientSecret: "ZJzvq_3kGkmRjG3KOAQGweiLsvKytuRjKrB_l2cMcCUlo10oKpMNGzGQn8ns7B4E",
	AuthURL:      "https://bukidutility.auth0.com/authorize",
	//RedirectURL:  "http://localhost:8080/oauth2callback",
	RedirectURL: "http://bukidutility.appspot.com/oauth2callback",
	TokenURL:    "https://bukidutility.auth0.com/oauth/token",
	Scope:       "profile",
}

const profileInfoURL = "https://bukidutility.auth0.com/userinfo?alt=json"

// This is where Google App Engine sets up which handler lives at the root url.
func init() {

	GlobalSessions, _ = session.NewManager("memory", `{"cookieName":"gosessionid","gclifetime":3600}`)
	go GlobalSessions.GC()

	// Immediately enter the main app.
	main()
}

func main() {

	router := mux.NewRouter().StrictSlash(true)

	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	router.PathPrefix("/static/css/").Handler(http.StripPrefix("/static/css/", http.FileServer(http.Dir("static/css/"))))
	router.PathPrefix("/static/fonts/").Handler(http.StripPrefix("/static/fonts/", http.FileServer(http.Dir("static/fonts/"))))
	router.PathPrefix("/static/images/").Handler(http.StripPrefix("/static/images/", http.FileServer(http.Dir("static/images/"))))
	router.PathPrefix("/static/js/min/").Handler(http.StripPrefix("/static/js/min/", http.FileServer(http.Dir("static/js/min/"))))
	router.PathPrefix("/static/js/vendor/").Handler(http.StripPrefix("/static/js/vendor/", http.FileServer(http.Dir("static/js/vendor/"))))
	router.PathPrefix("/static/js/").Handler(http.StripPrefix("/static/js/", http.FileServer(http.Dir("static/js/"))))

	router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("public/"))))

	router.HandleFunc("/oauth2callback", handleOAuth2Callback)
	router.HandleFunc("/authorize", handleAuthorize)
	router.HandleFunc("/user", UserHandler)
	router.HandleFunc("/registeriot", registeriotHandler)
	router.HandleFunc("/variabledata", variableDataHandler)

	router.HandleFunc("/main", mainHandler)
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/", homeHandler)

	router.HandleFunc("/api/v1/farm/{farmID}", createHandler)
	router.HandleFunc("/api/v1/farm/", farmIndexHandler)
	router.HandleFunc("/api/v1/recipe/{recipeID}", recipeValHandler)
	router.HandleFunc("/api/v1/recipe/", recipeIndexHandler)
	router.HandleFunc("/api/v1/", apiIndexHandler)
	router.HandleFunc("/master/", masterListHandler)
	router.HandleFunc("/createrecipe/", createRecipeHandler)

	http.Handle("/", router)

}

func masterListHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/master/" {
		errorHandler(w, r, http.StatusNotFound, "")
		return
	}
	c := appengine.NewContext(r)

	if r.Method == "GET" {
		q := datastore.NewQuery("tblVariables") //.Filter("VariableName =", "PH")

		var vars []variables
		if _, err := q.GetAll(c, &vars); err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "")
			return
		}

		page := template.Must(template.ParseFiles(
			"static/_base.html",
			"static/variables.html",
		))
		if err := page.Execute(w, vars); err != nil {
			errorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}

	} else {
		c := appengine.NewContext(r)
		q, err := datastore.NewQuery("tblVariables").KeysOnly().Count(c) //.Filter("Name =", recipeID) //.Order("-DateAdded")
		if err != nil {
			fmt.Fprintf(w, `count err: %s`, err)
			return
		}

		vars := variables{
			ID:           q + 1,
			VariableName: r.FormValue("variableName"),
			Description:  r.FormValue("variableDescription"),
			Unit:         r.FormValue("variableUnit"),
			Validation:   r.FormValue("variableValidation"),
			DateAdded:    time.Now(),
		}

		key, err := datastore.Put(c, datastore.NewIncompleteKey(c, "tblVariables", nil), &vars)
		if err != nil {
			panic(key)
			errorHandler(w, r, http.StatusInternalServerError, "")
			return
		}

		http.Redirect(w, r, "/master/", http.StatusFound)
	}
}

func variableDataHandler(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	q := datastore.NewQuery("tblDATA").Order("-TimeCreated")

	var value []values
	if _, err := q.GetAll(c, &value); err != nil {
		panic(err)
		//errorHandler(w, r, http.StatusInternalServerError, "")
		return
	}

	if r.URL.Path != "/variabledata" {
		errorHandler(w, r, http.StatusNotFound, "")
		return
	}
	page := template.Must(template.ParseFiles(
		"static/_base.html",
		"static/variableData.html",
	))

	if err := page.Execute(w, value); err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

}

func registeriotHandler(w http.ResponseWriter, r *http.Request) {

	session, _ := GlobalSessions.SessionStart(w, r)
	defer session.SessionRelease(w)

	response, _ := json.MarshalIndent(session.Get("profile"), "", " ")
	m := objx.MustFromJSON(string(response))
	user_id := m.Get("user_id").Str()

	if user_id != "" {

		c := appengine.NewContext(r)
		q := datastore.NewQuery("tblUsers").Filter("UID =", user_id)

		var user []users
		if _, err := q.GetAll(c, &user); err != nil {
			//panic(err)
			errorHandler(w, r, http.StatusInternalServerError, "")
			return
		}

		if r.URL.Path != "/registeriot" {
			errorHandler(w, r, http.StatusNotFound, "")
			return
		}

		page := template.Must(template.ParseFiles(
			"static/_base.html",
			"static/registeriot.html",
		))

		if err := page.Execute(w, user); err != nil {
			errorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}
	}
}
func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, data interface{}) {

	response, _ := json.MarshalIndent(data, "", " ")
	//fmt.Fprintln(w, string(response))

	m := objx.MustFromJSON(string(response))
	user_id := m.Get("user_id").Str()
	//fmt.Fprintln(w, user_id)

	hash := sha256.New()
	hash.Write([]byte(user_id))
	sha256_hash := hex.EncodeToString(hash.Sum(nil))

	c := appengine.NewContext(r)
	q, err := datastore.NewQuery("tblUsers").Filter("UID =", user_id).Count(c) //.Order("-DateAdded")
	if err != nil {
		//fmt.Fprintf(w, `count err: %s`, err)
		return
	}
	if q == 0 {
		us := users{
			UID:            user_id,
			DateRegistered: time.Now(),
			Token:          sha256_hash,
		}

		k, err := datastore.Put(c, datastore.NewIncompleteKey(c, "tblUsers", nil), &us)
		if err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "")
			fmt.Print(k)
			return
		}
	}

	cwd, _ := os.Getwd()
	t, err := template.ParseFiles(filepath.Join(cwd, "/static/index.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	//hash := sha256.New()
	//hash.Write([]byte(user_id))
	//sha256_hash := hex.EncodeToString(hash.Sum(nil))
	//fmt.Fprintln(w, sha256_hash)
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := GlobalSessions.SessionStart(w, r)
	defer session.SessionRelease(w)
	session.Get("profile")

	//if session.Get("profile") == nil {
	//	http.Redirect(w, r, "/", http.StatusMovedPermanently)
	//}

	RenderTemplate(w, r, "user", session.Get("profile"))
}

// Root directory handler.
func loginHandler(rw http.ResponseWriter, req *http.Request) {

	err := cached_templates.ExecuteTemplate(rw, "notAuthenticated.html", nil)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
	}

}

func IsAuthenticated(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	session, _ := GlobalSessions.SessionStart(w, r)
	defer session.SessionRelease(w)
	if session.Get("profile") == nil {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	} else {
		next(w, r)
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
func handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {

	// Initialize an appengine context.
	c := appengine.NewContext(r)

	// Retrieve the code from the response.
	code := r.FormValue("code")

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

	raw, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var profile map[string]interface{}
	if err := json.Unmarshal(raw, &profile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, _ := GlobalSessions.SessionStart(w, r)
	defer session.SessionRelease(w)

	//session.Set("id_token", token.Extra("id_token"))
	session.Set("access_token", token.AccessToken)
	session.Set("profile", profile)

	http.Redirect(w, r, "/user", http.StatusMovedPermanently) // Redirect to logged in page

}

func farmIndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {

		c := appengine.NewContext(r)
		q := datastore.NewQuery("tblDATA")

		var value []values
		if _, err := q.GetAll(c, &value); err != nil {
			panic(err)
			//errorHandler(w, r, http.StatusInternalServerError, "")
			return
		}

		response, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			panic(err)
			//errorHandler(w, r, http.StatusInternalServerError, "")
			return
		}
		fmt.Fprintln(w, string(response))

	} else {
		fmt.Fprintln(w, "ACCESS DENIED!")
		return
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		//w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		//w.WriteHeader(http.StatusOK)

		vars := mux.Vars(r)
		farmID := vars["farmID"]

		c := appengine.NewContext(r)
		q := datastore.NewQuery("tblDATA").Filter("FarmID =", farmID) //.Order("-DateAdded")
		var value []values
		if _, err := q.GetAll(c, &value); err != nil {
			panic(err)
			//errorHandler(w, r, http.StatusInternalServerError, "")
			return
		}

		response, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "")
			return
		}
		fmt.Fprint(w, string(response))

	} else {
		//w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		//w.WriteHeader(http.StatusOK)

		vars := mux.Vars(r)
		farmID := vars["farmID"]

		decoder := json.NewDecoder(r.Body)

		var v values
		err := decoder.Decode(&v)
		if err != nil {
			panic(err)
			//errorHandler(w, r, http.StatusInternalServerError, "")
			return
		}

		vals := values{
			TimeCreated: time.Now(),
			TimeStored:  time.Now(),
			VariableID:  farmID,
			DeviceUse:   "M",
			FarmID:      farmID,
			Token:       r.Header.Get("X-Farm-Token"),
			Value:       v.Value,
		}

		c := appengine.NewContext(r)

		key, err := datastore.Put(c, datastore.NewIncompleteKey(c, "tblDATA", nil), &vals)
		if err != nil {
			panic(err)
			panic(key)
			//errorHandler(w, r, http.StatusNotFound, "")
			return
		}

		response, err := json.MarshalIndent(vals, "", "  ")
		if err != nil {
			panic(err)
			//errorHandler(w, r, http.StatusInternalServerError, "")
		}
		fmt.Fprint(w, string(response))

	}
}

func createRecipeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/createrecipe/" {
		errorHandler(w, r, http.StatusNotFound, "")
		//return
	}
	if r.Method == "GET" {

		page := template.Must(template.ParseFiles(
			"static/_base.html",
			"static/createRecipe.gtpl",
		))

		if err := page.Execute(w, nil); err != nil {
			errorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}

	} else {

		c := appengine.NewContext(r)
		q, err := datastore.NewQuery("tblRecipes").KeysOnly().Count(c) //.Filter("Name =", recipeID) //.Order("-DateAdded")
		if err != nil {
			fmt.Fprintf(w, `count err: %s`, err)
			return
		}
		//i := strconv.Itoa(q)
		recipe := recipes{
			Recipeid:    "recipe" + strconv.Itoa(q),
			Name:        r.FormValue("recipename"),
			Description: r.FormValue("description"),
			DateCreated: time.Now(),
			//RecipeVars:  r.FormValue("var1") + "," + r.FormValue("var2") + "," + r.FormValue("var3") + "," + r.FormValue("var4") + "," + r.FormValue("var5"),

			//RecipeVars:  recipeVars,
			//RecipeVars: "{\"id\":\"" + r.FormValue("var1") + "\"}, {\"id\":\"" + r.FormValue("var2") + "\"}, {\"id\":\"" + r.FormValue("var3") + "\"}, {\"id\":\"" + r.FormValue("var4") + "\"}, {\"id\":\"" + r.FormValue("var5") + "\"}",
			//RecipeVars: "\"" + r.FormValue("var1") + "\", \"" + r.FormValue("var2") + "\", \"" + r.FormValue("var3") + "\", \"" + r.FormValue("var4") + "\", \"" + r.FormValue("var5") + "\"",
			//RecipeVars: r.FormValue("var1"),
		}

		key, err := datastore.Put(c, datastore.NewIncompleteKey(c, "tblRecipes", nil), &recipe)
		if err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "")
			panic(key)
			return
		}

		http.Redirect(w, r, "/createrecipe/", http.StatusFound)
	}
}

func recipeIndexHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/api/v1/recipe/" {
		errorHandler(w, r, http.StatusNotFound, "")
		return
	}

	if r.Method == "GET" {

		c := appengine.NewContext(r)
		q := datastore.NewQuery("tblRecipes").Order("Recipeid")

		var recipe []recipes
		if _, err := q.GetAll(c, &recipe); err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "")
			return
		}

		response, err := json.MarshalIndent(recipe, "", "  ")
		if err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "")
			return
		}
		fmt.Fprintln(w, string(response))

	} else {
		fmt.Fprint(w, "{\n  'method':'post',\n  'details':'access denied'\n}\n")
		return
	}
}

func recipeValHandler(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)

	if r.Method == "GET" {
		vars := mux.Vars(r)
		recipeID := vars["recipeID"]
		q := datastore.NewQuery("tblRecipes").Filter("Recipeid =", recipeID)
		var recipe []recipes
		if _, err := q.GetAll(c, &recipe); err != nil {
			panic(err)
			//errorHandler(w, r, http.StatusInternalServerError, "")
			return
		}

		// response, err := json.MarshalIndent(recipe, "", "  ")
		// if err != nil {
		// 	errorHandler(w, r, http.StatusInternalServerError, "")
		// 	return
		// }
		// fmt.Fprintln(w, string(response))

		for _, rec := range recipe {

			data := rec.RecipeVars
			res := strings.Split(data, ",")
			//fmt.Fprintln(w, res)
			//dc := strconv.Itoa(rec.DateCreated)
			fmt.Fprintln(w, "{\"recipe id\": \""+rec.Recipeid+"\",")
			fmt.Fprintln(w, "\"name\": \""+rec.Name+"\",")
			fmt.Fprintln(w, "\"description\": \""+rec.Description+"\",")
			//fmt.Fprintln(w, "\"date created\" :\""+dc+"\"")
			fmt.Fprintln(w, "\"recipe vars\": [")

			for count, i := range res {
				var s string = i
				id, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					panic(err)
				}
				q := datastore.NewQuery("tblVariables").Filter("ID =", id)

				var vars []variables
				if _, err := q.GetAll(c, &vars); err != nil {
					errorHandler(w, r, http.StatusInternalServerError, "")
					return
				}
				for _, vars2 := range vars {
					fmt.Fprint(w, "{ \"id\" :\"", vars2.ID, "\",")
					fmt.Fprint(w, "\"name\" :\"", vars2.VariableName, "\",")
					fmt.Fprint(w, "\"validation\" :\"", vars2.Validation, "\",")
					fmt.Fprint(w, "\"unit\" :\"", vars2.Unit, "\"}")
					if count+1 != len(res) {
						fmt.Fprint(w, ",\n")
					}
				}

			}
			fmt.Fprintln(w, "]}")

		}

	} else {
		fmt.Fprint(w, "{\n  'method':'post',\n  'details':'access denied'\n}\n")
		return
	}
}

func apiIndexHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/api/v1/" {
		errorHandler(w, r, http.StatusNotFound, "")
		return
	}

	if r.Method == "GET" {
		fmt.Fprintln(w, "BUKID UTILITY API VERSION 1.0")
		fmt.Fprintln(w, time.Now())

	} else {
		fmt.Fprint(w, "{\n  'method':'post',\n  'details':'access denied'\n}\n")
		return
	}

}

func homeHandler(w http.ResponseWriter, r *http.Request) {

	session, _ := GlobalSessions.SessionStart(w, r)
	defer session.SessionRelease(w)

	response, _ := json.MarshalIndent(session.Get("profile"), "", " ")
	m := objx.MustFromJSON(string(response))
	user_id := m.Get("user_id").Str()

	if user_id != "" {
		if r.URL.Path != "/" {

			errorHandler(w, r, http.StatusNotFound, "")
			return
		}

		page := template.Must(template.ParseFiles(
			"static/index.html",
		))
		if err := page.Execute(w, nil); err != nil {
			errorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func mainHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/main" {
		errorHandler(w, r, http.StatusNotFound, "")
		return
	}

	page := template.Must(template.ParseFiles(
		"static/_base.html",
		"static/main.html",
	))

	if err := page.Execute(w, nil); err != nil {
		errorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

}
func errorHandler(w http.ResponseWriter, r *http.Request, status int, err string) {
	w.WriteHeader(status)
	switch status {

	case http.StatusNotFound:
		page := template.Must(template.ParseFiles(
			"static/_base.html",
			"static/404.html",
		))
		if err := page.Execute(w, nil); err != nil {
			errorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}

	case http.StatusInternalServerError:
		page := template.Must(template.ParseFiles(
			"static/_base.html",
			"static/500.html",
		))
		if err := page.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
