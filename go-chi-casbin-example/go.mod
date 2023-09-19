module github.com/kitabisa/casbin-example

go 1.18

require (
	github.com/Blank-Xu/sqlx-adapter 0a30309eefa6
	github.com/casbin/chi-authz f77abe171fc6
	github.com/go-chi/chi v1.5.4
	github.com/go-sql-driver/mysql v1.6.0
	github.com/jmoiron/sqlx v1.3.5
)

require (
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible // indirect
	github.com/casbin/casbin/v2 v2.55.1 // indirect
)

replace github.com/casbin/chi-authz => ../chi-authz
