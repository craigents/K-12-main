package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings" 

	"example.com/k12platform/ent"
	"example.com/k12platform/ent/migrate"
	"example.com/k12platform/ent/user"

	"github.com/casbin/casbin/v2"
	basicadapter "github.com/casbin/basic-adapter/v2"

	_ "github.com/mattn/go-sqlite3"
)

// ... (contextKey, UserCreatePayload, casbinEnforcer, BasicAuthMiddleware, Authorizer, createUserHandler remain the same) ...
type contextKey string
const userIDKey contextKey = "userID"
const userRoleKey contextKey = "userRole"

type UserCreatePayload struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email"`
	Role      string `json:"role,omitempty"`
}

var casbinEnforcer *casbin.Enforcer 

func BasicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		userRole := r.Header.Get("X-User-Role")

		if userID == "" || userRole == "" {
			log.Printf("BasicAuthMiddleware: X-User-ID or X-User-Role header not found. Attempting placeholder logic.")
			// Use r.Header.Get directly for placeholder check as userID/userRole might be partially set
			headerUserID := r.Header.Get("X-User-ID") // Get fresh header value
			if headerUserID == "user_admin_001" { 
				userID = "user_admin_001"
				userRole = "admin"
			} else if headerUserID == "user_teacher_001" {
				userID = "user_teacher_001"
				userRole = "teacher"
			} else {
				// Only set to anonymous_user if userID was truly empty from header
				if userID == "" { userID = "anonymous_user" }
				// Only set to anonymous if userRole was truly empty from header
				if userRole == "" { userRole = "anonymous" }
				log.Printf("BasicAuthMiddleware: Defaulting to UserID: %s, Role: %s", userID, userRole)
			}
		} else {
			log.Printf("BasicAuthMiddleware: Authenticated UserID: %s, Role: %s", userID, userRole)
		}
		
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		ctx = context.WithValue(ctx, userRoleKey, userRole)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Authorizer(e *casbin.Enforcer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roleCtx := r.Context().Value(userRoleKey)
			if roleCtx == nil {
				log.Println("Authorizer: User role not found in context. Denying access.")
				http.Error(w, "Forbidden: User role not available in context", http.StatusForbidden)
				return
			}
			userRole := roleCtx.(string)

			obj := r.URL.Path
			act := r.Method

			log.Printf("Authorizer: Checking permission for Role: %s, Path: %s, Method: %s", userRole, obj, act)

			allowed, err := e.Enforce(userRole, obj, act)
			if err != nil {
				log.Printf("Authorizer: Error during Enforce call: %v", err)
				http.Error(w, "Forbidden: Error checking permissions", http.StatusForbidden)
				return
			}

			if allowed {
				log.Printf("Authorizer: Access GRANTED for Role: %s, Path: %s, Method: %s", userRole, obj, act)
				next.ServeHTTP(w, r)
			} else {
				log.Printf("Authorizer: Access DENIED for Role: %s, Path: %s, Method: %s", userRole, obj, act)
				http.Error(w, "Forbidden: You don't have permission to access this resource.", http.StatusForbidden)
			}
		})
	}
}

func createUserHandler(client *ent.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Method check is now primarily handled by Authorizer based on policy.
		// if r.Method != http.MethodPost { ... } // This can be removed or kept as a fallback.

		var payload UserCreatePayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if payload.Username == "" || payload.Password == "" || payload.Email == "" {
			http.Error(w, "Username, password, and email are required", http.StatusBadRequest)
			return
		}
		
		role := user.RoleStudent 
		if payload.Role != "" {
			switch payload.Role {
			case "teacher":
				role = user.RoleTeacher
			case "admin":
				role = user.RoleAdmin
			case "student":
			default:
                log.Printf("Invalid role specified: %s, defaulting to student", payload.Role)
			}
		}

		u, err := client.User.
			Create().
			SetUsername(payload.Username).
			SetHashedPassword(payload.Password). 
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
	// ... (Ent Client and Casbin Enforcer initialization) ...
	client, err := ent.Open("sqlite3", "file:k12platform.db?cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()

	ctx := context.Background() // Main context for setup
	if err := client.Schema.Create(
		ctx, // Use main setup context for schema creation
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	adapter := basicadapter.NewAdapter("./casbin_policy.csv")
	enforcer, err := casbin.NewEnforcer("./casbin_model.conf", adapter)
	if err != nil {
		log.Fatalf("failed to create casbin enforcer: %v", err)
	}
	if err := enforcer.LoadPolicy(); err != nil {
		log.Fatalf("failed to load casbin policy: %v", err)
	}
	casbinEnforcer = enforcer

	// Define handlers
	rootHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use r.Context() for operations within a request handler.
		dbUserCount, _ := client.User.Query().Count(r.Context()) // Ignoring error for brevity
		fmt.Fprintf(w, "Hello, World! DB Users: %d. Casbin Initialized. Auth UserID: %s, Role: %s",
			dbUserCount, r.Context().Value(userIDKey), r.Context().Value(userRoleKey))
	})
	
	usersHandler := http.HandlerFunc(createUserHandler(client))
	coursesHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Course data. Auth UserID: %s, Role: %s. Method: %s",
			r.Context().Value(userIDKey), r.Context().Value(userRoleKey), r.Method)
	})

	// Apply middlewares
	// Root handler: only basic authentication
	http.Handle("/", BasicAuthMiddleware(rootHandler))

	// Users handler: basic authentication then Casbin authorization
	http.Handle("/users", BasicAuthMiddleware(Authorizer(casbinEnforcer)(usersHandler)))

	// Courses handler: basic authentication then Casbin authorization
	http.Handle("/courses", BasicAuthMiddleware(Authorizer(casbinEnforcer)(coursesHandler)))

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
