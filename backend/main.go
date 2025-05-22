package main

import (
	"context"
	"encoding/json" // Added
	"fmt"
	"log"
	"net/http"
	"os"

	"example.com/k12platform/ent"
	"example.com/k12platform/ent/migrate"
	"example.com/k12platform/ent/user" // Added for role constants

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// UserCreatePayload defines the expected JSON structure for creating a user
type UserCreatePayload struct {
	Username  string `json:"username"`
	Password  string `json:"password"` // In a real app, hash this securely
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email"`
	Role      string `json:"role,omitempty"` // e.g., "student", "teacher", "admin"
}

// createUserHandler handles requests to POST /users
func createUserHandler(client *ent.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload UserCreatePayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		// Basic validation
		if payload.Username == "" || payload.Password == "" || payload.Email == "" {
			http.Error(w, "Username, password, and email are required", http.StatusBadRequest)
			return
		}
		
		// Determine the role
		role := user.RoleStudent // Default role
		if payload.Role != "" {
			switch payload.Role {
			case "teacher":
				role = user.RoleTeacher
			case "admin":
				role = user.RoleAdmin
			case "student":
				// already default
			default:
                log.Printf("Invalid role specified: %s, defaulting to student", payload.Role)
			}
		}

		// INSECURE: Password should be hashed in a real application.
		u, err := client.User.
			Create().
			SetUsername(payload.Username).
			SetHashedPassword(payload.Password). // Store as is (INSECURE)
			SetFirstName(payload.FirstName).
			SetLastName(payload.LastName).
			SetEmail(payload.Email).
			SetRole(role).
			Save(r.Context())

		if err != nil {
			if ent.IsConstraintError(err) {
				http.Error(w, fmt.Sprintf("Failed to create user (constraint violation, e.g., username or email already exists): %v", err), http.StatusConflict)
			} else {
				http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(u)
	}
}

func main() {
	// Initialize Ent Client
	client, err := ent.Open("sqlite3", "file:k12platform.db?cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()

	// Run the auto migration tool.
	ctx := context.Background()
	if err := client.Schema.Create(
		ctx,
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	// HTTP Handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		count, err := client.User.Query().Count(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to count users: %v", err), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Hello, World! DB Connection Successful. Users in DB: %d", count)
	})

	http.HandleFunc("/users", createUserHandler(client))

	// Server Startup
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
