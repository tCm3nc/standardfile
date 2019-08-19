package main

import (
	"net"

	"github.com/go-playground/pure"
	mw "github.com/go-playground/pure/_examples/middleware/logging-recovery"

	// "github.com/go-playground/pure/middleware"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/tectiv3/standardfile/db"
)

func worker() {
	db.Init(cfg.DB)
	log.Println("Started StandardFile Server", Version)
	log.Println("Loaded config:", loadedConfig)

	if cfg.Debug {
		log.Println("Debug on")
	}

	r := pure.New()
	if cfg.UseCORS {
    log.Println("Using CORS");
		r.Use(mw.LoggingAndRecovery(true), cors)
		r.RegisterAutomaticOPTIONS(cors)
	} else {
		r.Use(mw.LoggingAndRecovery(true), cors)
    r.RegisterAutomaticOPTIONS(cors)
	}
	//r.Handle("/", cors(http.HandlerFunc(Dashboard));
	r.Get("/", cors(http.HandlerFunc(Dashboard)))
	r.Post("/api/items/sync", SyncItems)
	r.Post("/api/items/backup", BackupItems)
	// r.DELETE("/api/items", DeleteItems)
	if !cfg.NoReg {
		r.Post("/api/auth", Registration)
	}
	r.Patch("/api/auth", ChangePassword)
	r.Post("/api/auth/update", UpdateUser)
	r.Post("/api/auth/change_pw", ChangePassword)
	r.Post("/api/auth/sign_in", Login)
	r.Post("/api/auth/sign_in.json", Login)
	r.Get("/api/auth/params", GetParams)

	defer removeSock()
	go listen(r)
	<-run
	log.Println("Server stopped")
	os.Exit(0)
}

func listen(r *pure.Mux) {
	if len(cfg.Socket) != 0 {
		os.Remove(cfg.Socket)
		unixListener, err := net.Listen("unix", cfg.Socket)
		if err != nil {
			panic(err)
		}
		server := http.Server{
			Handler: r.Serve(),
		}
		log.Println("Listening on socket " + cfg.Socket)
		server.Serve(unixListener)
	} else {
		log.Println("Listening on port " + strconv.Itoa(cfg.Port))
		err := http.ListenAndServe(":"+strconv.Itoa(cfg.Port), r.Serve())
		if err != nil {
			log.Println(err)
		}
	}
}

func cors(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Got a request : %v\n", r);
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "authorization,content-type")
		}
		//next(w, r)
		next.ServeHTTP(w, r)
	})
}

func removeSock() {
	if len(cfg.Socket) != 0 {
		os.Remove(cfg.Socket)
	}
}
