package main

import (
	"fmt"
	"github.com/casbin/casbin/v2"
	authz "github.com/casbin/chi-authz"
	"github.com/go-chi/chi"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
)

func finalizer(db *sqlx.DB) {
	err := db.Close()
	if err != nil {
		panic(err)
	}
}

func main() {
	//db, err := sqlx.Connect("mysql", "root:pass@tcp(127.0.0.1:3396)/casbin")
	//if err != nil {
	//	panic(err)
	//}

	//db.SetMaxOpenConns(20)
	//db.SetMaxIdleConns(10)
	//db.SetConnMaxLifetime(time.Minute * 10)
	//runtime.SetFinalizer(db, finalizer)

	//a, err := sqlxadapter.NewAdapter(db, "casbin_rule_test")
	//if err != nil {
	//	panic(err)
	//}

	router := chi.NewRouter()

	// load the casbin model and policy from files, database is also supported.
	e, err := casbin.NewEnforcer("authz_model.conf", "authz_policy.csv")
	//e, err := casbin.NewEnforcer("authz_model.conf", a)
	if err != nil {
		fmt.Println(err)
	}

	// Load the policy from DB.
	if err = e.LoadPolicy(); err != nil {
		log.Println("LoadPolicy failed, err: ", err)
	}

	router.Group(func(r chi.Router) {
		r.Route("/api/v1", func(r chi.Router) {
			r.Use(authz.Authorizer(e))
			r.Get("/data1", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("bisa akses get endpoint data1"))
			})
			r.Post("/data1", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("bisa akses post endpoint data1"))
			})

			r.Get("/data2", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("bisa akses endpoint data2"))
			})
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("bisa akses endpoint index/root"))
			})
		})
	})

	// define your handler, this is just an example to return HTTP 200 for any requests.
	// the access that is denied by authz will return HTTP 403 error.
	router.Get("/api/v2/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("GET api v2"))
	})

	router.Post("/api/v2/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("POST api v2"))
	})

	router.Post("/api/v2/updatepolicy", func(w http.ResponseWriter, r *http.Request) {
		res, err := e.AddPolicy("alice", "/api/v1/data1", "GET")
		if err != nil {
			panic(err)
		}

		fmt.Println(res)
		err = e.SavePolicy()

		if err != nil {
			panic(err)
		}

		w.WriteHeader(200)
		w.Write([]byte("POST api update policy"))
	})

	http.ListenAndServe(":8081", router)
}
