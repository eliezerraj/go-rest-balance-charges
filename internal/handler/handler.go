package handler

import (
	"strconv"
	"net/http"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/gorilla/mux"

	"github.com/go-rest-balance-charges/internal/service"
	"github.com/go-rest-balance-charges/internal/core"
	"github.com/go-rest-balance-charges/internal/erro"
	
)

var childLogger = log.With().Str("handler", "handler").Logger()

type HttpWorkerAdapter struct {
	workerService 	*service.WorkerService
}

func NewHttpWorkerAdapter(workerService *service.WorkerService) *HttpWorkerAdapter {
	childLogger.Debug().Msg("NewHttpWorkerAdapter")
	return &HttpWorkerAdapter{
		workerService: workerService,
	}
}

func MiddleWareHandlerHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		childLogger.Debug().Msg("-------------- MiddleWareHandlerHeader (INICIO) --------------")
	
		if reqHeadersBytes, err := json.Marshal(r.Header); err != nil {
			childLogger.Error().Err(err).Msg("Could not Marshal http headers !!!")
		} else {
			childLogger.Debug().Str("Headers : ", string(reqHeadersBytes) ).Msg("")
		}

		childLogger.Debug().Str("Method : ", r.Method ).Msg("")
		childLogger.Debug().Str("URL : ", r.URL.Path ).Msg("")

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers","Content-Type,access-control-allow-origin, access-control-allow-headers")
	
		//log.Println(r.Header.Get("Host"))
		//log.Println(r.Header.Get("User-Agent"))
		//log.Println(r.Header.Get("X-Forwarded-For"))

		childLogger.Debug().Msg("-------------- MiddleWareHandlerHeader (FIM) ----------------")

		next.ServeHTTP(w, r)
	})
}

func (h *HttpWorkerAdapter) Health(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Health")

	health := true
	json.NewEncoder(rw).Encode(health)
	return
}

func (h *HttpWorkerAdapter) Live(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Live")

	live := true
	json.NewEncoder(rw).Encode(live)
	return
}

func (h *HttpWorkerAdapter) Header(rw http.ResponseWriter, req *http.Request) {
	log.Printf("/header")
	
	json.NewEncoder(rw).Encode(req.Header)
	return
}

func (h *HttpWorkerAdapter) Add(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Add")

	balanceCharge := core.BalanceCharge{}
	err := json.NewDecoder(req.Body).Decode(&balanceCharge)
    if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(erro.ErrUnmarshal.Error())
        return
    }
	
	res, err := h.workerService.Add(balanceCharge)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err.Error())
		return
	}

	json.NewEncoder(rw).Encode(res)
	return
}

func (h *HttpWorkerAdapter) Get(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Get")

	vars := mux.Vars(req)
	varID, err := strconv.Atoi(vars["id"])
	if err != nil{
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(erro.ErrConvertion.Error())
		return
	}

	balanceCharge := core.BalanceCharge{}
	balanceCharge.ID = varID
	
	res, err := h.workerService.Get(balanceCharge)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err.Error())
		return
	}

	json.NewEncoder(rw).Encode(res)
	return
}

func (h *HttpWorkerAdapter) List(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("List")

	vars := mux.Vars(req)
	varID := (vars["id"])

	balanceCharge := core.BalanceCharge{}
	balanceCharge.AccountID = varID
	
	res, err := h.workerService.List(balanceCharge)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err.Error())
		return
	}

	json.NewEncoder(rw).Encode(res)
	return
}