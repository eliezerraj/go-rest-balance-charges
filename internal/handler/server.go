package handler

import (
	"time"
	"encoding/json"
	"net/http"
	"strconv"
	"os"
	"os/signal"
	"syscall"
	"context"

	"github.com/gorilla/mux"

	"github.com/go-rest-balance-charges/internal/core"
	"github.com/aws/aws-xray-sdk-go/xray"

)

type HttpServer struct {
	start 			time.Time
	httpAppServer 	core.HttpAppServer
}

func NewHttpAppServer( httpAppServer core.HttpAppServer) HttpServer {
	childLogger.Debug().Msg("NewHttpAppServer")

	return HttpServer{	start: time.Now(), 
						httpAppServer: httpAppServer,
					}
}

func (h HttpServer) StartHttpAppServer(ctx context.Context, httpWorkerAdapter *HttpWorkerAdapter) {
	childLogger.Info().Msg("StartHttpAppServer")

	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		json.NewEncoder(rw).Encode(h.httpAppServer)
	})

	myRouter.HandleFunc("/info", func(rw http.ResponseWriter, req *http.Request) {
		json.NewEncoder(rw).Encode(h.httpAppServer)
	})

	health := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    health.HandleFunc("/health", httpWorkerAdapter.Health)
	health.Use(MiddleWareHandlerHeader)

	live := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    live.HandleFunc("/live", httpWorkerAdapter.Live)
	live.Use(MiddleWareHandlerHeader)

	header := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    header.HandleFunc("/header", httpWorkerAdapter.Header)
	header.Use(MiddleWareHandlerHeader)

	addBalance := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
    //addBalance.HandleFunc("/add", httpWorkerAdapter.Add)
	addBalance.Handle("/add", 
		xray.Handler(xray.NewFixedSegmentNamer("go-rest-balance-charges.add"), 
		http.HandlerFunc(httpWorkerAdapter.Add),
		),
	)
	addBalance.Use(MiddleWareHandlerHeader)

	getBalance := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    //getBalance.HandleFunc("/get/{id}", httpWorkerAdapter.Get)
	getBalance.Handle("/get/{id}", 
		xray.Handler(xray.NewFixedSegmentNamer("go-rest-balance-charges.getId"), 
		http.HandlerFunc(httpWorkerAdapter.Get),
		),
	)
	getBalance.Use(MiddleWareHandlerHeader)

	listBalance := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    //listBalance.HandleFunc("/list/{id}", httpWorkerAdapter.List)
	listBalance.Handle("/list/{id}", 
		xray.Handler(xray.NewFixedSegmentNamer("go-rest-balance-charges.listId"), 
		http.HandlerFunc(httpWorkerAdapter.List),
		),
	)
	listBalance.Use(MiddleWareHandlerHeader)

	srv := http.Server{
		Addr:         ":" +  strconv.Itoa(h.httpAppServer.Server.Port),      	
		Handler:      myRouter,                	          
		ReadTimeout:  time.Duration(h.httpAppServer.Server.ReadTimeout) * time.Second,   
		WriteTimeout: time.Duration(h.httpAppServer.Server.WriteTimeout) * time.Second,  
		IdleTimeout:  time.Duration(h.httpAppServer.Server.IdleTimeout) * time.Second, 
	}

	childLogger.Info().Str("Service Port : ", strconv.Itoa(h.httpAppServer.Server.Port)).Msg("Service Port")

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			childLogger.Error().Err(err).Msg("Cancel http mux server !!!")
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch

	ctx , cancel := context.WithTimeout(context.Background(), time.Duration(h.httpAppServer.Server.CtxTimeout) * time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		childLogger.Error().Err(err).Msg("WARNING Dirty Shutdown !!!")
		return
	}
	childLogger.Info().Msg("Stop Done !!!!")
}