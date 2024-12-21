package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuth
var googleOauthConfig = &oauth2.Config{
	ClientID:     "nothing",         // Replace with actual ID
	ClientSecret: "to see here",     // Replace with actual Secret
	RedirectURL:  "http://localhost:8080/auth/google/callback",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{"message": "Google OAuth Callback"})
}

// Python Script
func RunPythonHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("python3", "main.py")
	output, err := cmd.CombinedOutput()
	if err != nil {
		render.JSON(w, r, map[string]string{"error": fmt.Sprintf("Error running Python script: %v\nOutput: %s", err, output)})
		return
	}
	render.JSON(w, r, map[string]string{"output": string(output)})
}

// Bcrypt
func HashPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		render.JSON(w, r, map[string]string{"error": "Invalid request payload"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		render.JSON(w, r, map[string]string{"error": "Failed to hash password"})
		return
	}

	render.JSON(w, r, map[string]string{"hashed_password": string(hash)})
}

func ComparePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Password string `json:"password"`
		Hash     string `json:"hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		render.JSON(w, r, map[string]string{"error": "Invalid request payload"})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(request.Hash), []byte(request.Password))
	if err != nil {
		render.JSON(w, r, map[string]string{"message": "Password does not match"})
	} else {
		render.JSON(w, r, map[string]string{"message": "Password matches"})
	}
}

var request struct {
	Email    string `json:"email"`
	Username string `json:"name"`
	Password string `json:"password"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		render.JSON(w, r, map[string]string{"error": "Invalid request payload"})
		return
	}

	fmt.Printf("Email: %s, Username: %s, Password: %s\n", request.Email, request.Username, request.Password)
	render.JSON(w, r, request)
}

// Main
func main() {
	// Initialize the router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, map[string]string{"message": "Welcome to the DETECT Go API Server!"})
	})

	// OAuth Routes
	r.Get("/auth/google/login", GoogleLoginHandler)
	r.Get("/auth/google/callback", GoogleCallbackHandler)

	// Python Script Execution Route
	r.Get("/run-python", RunPythonHandler)

	// Bcrypt Routes
	r.Post("/hash-password", HashPasswordHandler)
	r.Post("/compare-password", ComparePasswordHandler)

	// Registration Route
	r.Post("/register", RegisterHandler)

	// Start the server
	log.Println("Starting server on http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
