module github.com/JBoudou/Itero

go 1.15

require (
	github.com/felixge/httpsnoop v1.0.1
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.1
	github.com/justinas/alice v1.2.0
	github.com/tevino/abool v1.2.0
	golang.org/x/crypto v0.0.0-20201117144127-c1f2f97bffc9
	golang.org/x/sys v0.0.0-20210503173754-0981d6026fa6 // indirect
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56 // indirect
)

replace github.com/go-sql-driver/mysql v1.6.0 => github.com/JBoudou/mysql v1.6.1-0.20210507083111-2eaa51c65ad1
