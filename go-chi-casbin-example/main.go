package main

import (
	"context"
	"database/sql"
	"fmt"
	sqlxadapter "github.com/Blank-Xu/sqlx-adapter"
	"golang.org/x/crypto/bcrypt"

	"github.com/casbin/casbin/v2"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func finalizer(db *sqlx.DB) {
	err := db.Close()
	if err != nil {
		panic(err)
	}
}

var DB *sql.DB

func main() {
	dataSourceName := fmt.Sprintf("user=root password=pass host=localhost port=5432 dbname=casbin sslmode=disable")

	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Minute * 10)

	DB = db.DB
	runtime.SetFinalizer(db, finalizer)

	a, err := sqlxadapter.NewAdapter(db, "casbin_rule_test")
	if err != nil {
		panic(err)
	}

	router := chi.NewRouter()

	// load the casbin model and policy from files, database is also supported.
	e, err := casbin.NewEnforcer("authz_model.conf", a)
	//e, err := casbin.NewEnforcer("authz_model.conf", a)
	if err != nil {
		fmt.Println(err)
	}

	// Load the policy from DB.
	if err = e.LoadPolicy(); err != nil {
		log.Println("LoadPolicy failed, err: ", err)
	}

	router.Group(func(r chi.Router) {
		r.Route("/api/v1/admin", func(r chi.Router) {
			r.Use(tokenMiddleware)
			r.Use(Authorizer(e))
			r.Get("/index", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("bisa akses Get admin index"))
			})

			r.Post("/list", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("bisa akses Post admin list"))
			})

			r.Put("/update", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("bisa akses Put admin update"))
			})
		})

		r.Route("/api/v1/partner", func(r chi.Router) {
			r.Use(tokenMiddleware)
			r.Use(Authorizer(e))
			r.Get("/index", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("bisa akses Get partner index"))
			})

			r.Post("/list", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("bisa akses Post partner list"))
			})

			r.Put("/update", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("bisa akses Put partner update"))
			})
		})

		r.Route("/api/v1/superadmin", func(r chi.Router) {
			r.Use(tokenMiddleware)
			r.Use(Authorizer(e))
			r.Get("/index", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("bisa akses Get superadmin index"))
			})

			r.Post("/list", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("bisa akses Post superadmin list"))
			})

			r.Put("/update", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("bisa akses Put superadmin update"))
			})
		})
	})

	router.Post("/login", loginHandler)
	router.Post("/register", registerHandler)

	var srv http.Server
	srv.Addr = fmt.Sprintf("0.0.0.0:8085")
	srv.Handler = router

	log.Println("[API] start")
	idleConnectionClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
		}
		close(idleConnectionClosed)
	}()

	srv.ListenAndServe()

	<-idleConnectionClosed

	log.Println("[API] Bye")
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	role := r.FormValue("role")

	// Register user
	err := register(username, password, role)
	if err != nil {
		http.Error(w, "User registration failed", http.StatusInternalServerError)
		return
	}

	// Respond with a success message
	fmt.Fprintf(w, "User registered successfully!\n")
}

var secretKey = []byte("your-secret-key")

// CustomClaims represents the JWT claims
type CustomClaims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

func generateToken(userID uint64, username, role string) (string, string, error) {
	// Create Access Token
	accessTokenClaims := CustomClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(), // Token expires in 15 minutes
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString(secretKey)
	if err != nil {
		return "", "", err
	}

	// Create Refresh Token
	refreshTokenClaims := CustomClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(), // Refresh token expires in 7 days
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString(secretKey)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// User represents the user model
type User struct {
	ID       uint64
	Username string
	Password string
	Role     string
}

func register(username, password, role string) error {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("password hashing failed")
	}

	// Insert the user into the database
	_, err = DB.Exec("INSERT INTO users (username, password, role) VALUES ($1, $2, $3)", username, hashedPassword, role)
	if err != nil {
		return fmt.Errorf("user registration failed")
	}

	return nil
}

func authenticate(username, password string) (uint64, string, error) {
	var user User

	// Query the database to retrieve the user by username

	row := DB.QueryRow("SELECT id, username, password, role FROM users WHERE username = $1", username)
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		return 0, "", fmt.Errorf("user not found")
	}

	// Compare the stored password hash with the provided password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return 0, "", fmt.Errorf("invalid password")
	}

	return user.ID, user.Role, nil
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Authenticate user
	userID, role, err := authenticate(username, password)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Generate token
	accessToken, refreshToken, err := generateToken(userID, username, role)
	if err != nil {
		http.Error(w, "Token generation failed", http.StatusInternalServerError)
		return
	}

	// Respond with the access token
	fmt.Fprintf(w, "Login Successful!\n")
	fmt.Fprintf(w, "Access Token: %s\n", accessToken)
	fmt.Fprintf(w, "Refresh Token: %s\n", refreshToken)
}

func tokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from the request (assuming it's sent in the Authorization header)
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		// Extract the token part (assuming a format like "Bearer <token>")
		tokenString := authorizationHeader[len("Bearer "):]

		// Parse and verify the token
		token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})

		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Check if the token is valid
		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Access the custom claims
		claims, ok := token.Claims.(*CustomClaims)
		if !ok {
			http.Error(w, "Failed to extract claims", http.StatusInternalServerError)
			return
		}

		// Add the decoded claims to the request context for use in handlers
		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Authorizer(e *casbin.Enforcer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			userClaim, ok := r.Context().Value("claims").(*CustomClaims)
			if !ok {
				http.Error(w, http.StatusText(403), 403)
				return
			}

			method := r.Method
			path := r.URL.Path
			has, err := e.Enforce(userClaim.Username, path, method)
			if err != nil {
				fmt.Println(err)
			}

			if has {
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, http.StatusText(403), 403)
			}
		}

		return http.HandlerFunc(fn)
	}
}
