package main

import (
	"log"
	"net/http"
	"os"

	"github.com/stuart-warren/serveit/access"
	"github.com/stuart-warren/serveit/router"
	"github.com/stuart-warren/serveit/rules"
)

func main() {
	wd, _ := os.Getwd()
	static := http.FileServer(http.Dir(wd))
	mux := router.NewRouter(static, rules.CheckUser)
	mux.Handle(router.NewPrefixRoute("/access/").Permit(access.BlankPermit().MethodRW().AllowUsers("some.admin")))
	mux.Handle(router.NewPrefixRoute("/").Permit(access.BlankPermit().MethodRO().AllowUsers("ALL")))
	srv := &http.Server{Addr: ":1234", Handler: mux}
	log.Println("starting")
	log.Fatal(srv.ListenAndServe())
}
