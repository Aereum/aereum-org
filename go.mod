module github.com/Aereum/aereum-org

go 1.18

replace github.com/Aereum/aereum/core => ../aereum/core

require (
	github.com/Aereum/aereum/core v0.0.0-00010101000000-000000000000
	github.com/gobwas/ws v1.1.0
	github.com/gorilla/mux v1.8.0
)

require (
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	golang.org/x/sys v0.0.0-20201207223542-d4d67f95c62d // indirect
)
