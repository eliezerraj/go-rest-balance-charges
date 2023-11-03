package restapi

import(
	"errors"
	"net/http"
	"time"
	"encoding/json"
	"bytes"

	"github.com/rs/zerolog/log"
	"github.com/go-rest-balance-charges/internal/erro"
)

var childLogger = log.With().Str("adapter/restapi", "restapi").Logger()

type RestApiSConfig struct {
	serverUrlDomain			string
	path					string
}

func NewRestApi(serverUrlDomain string, path string) (*RestApiSConfig){
	childLogger.Debug().Msg("*** NewRestApi")
	return &RestApiSConfig {
		serverUrlDomain: 	serverUrlDomain,
		path: 	path,
	}
}

func (r *RestApiSConfig) GetData(id string) (interface{}, error) {
	childLogger.Debug().Msg("GetData")

	domain := r.serverUrlDomain + r.path +"/" + id

	childLogger.Debug().Str("domain : ", domain).Msg("GetData")

	data_interface, err := makeGet(domain, id)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return nil, errors.New(err.Error())
	}
    
	return data_interface, nil
}

func (r *RestApiSConfig) PostData(id string, data interface{}) (interface{}, error) {
	childLogger.Debug().Msg("PostData")

	domain := r.serverUrlDomain + "/update" +"/" + id

	childLogger.Debug().Str("domain : ", domain).Msg("PostData")

	data_interface, err := makePost(domain, id ,data)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return nil, errors.New(err.Error())
	}
    
	return data_interface, nil
}

func makeGet(url string, id interface{}) (interface{}, error) {
	childLogger.Debug().Msg("makeGet")

	client := &http.Client{Timeout: time.Second * 29}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return false, errors.New(err.Error())
	}

	req.Header.Add("Content-Type", "application/json;charset=UTF-8");

	resp, err := client.Do(req)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Do Request")
		return false, errors.New(err.Error())
	}

	childLogger.Debug().Int("StatusCode :", resp.StatusCode).Msg("")
	switch (resp.StatusCode) {
		case 401:
			return false, erro.ErrHTTPForbiden
		case 403:
			return false, erro.ErrHTTPForbiden
		case 200:
		case 400:
			return false, erro.ErrNotFound
		case 404:
			return false, erro.ErrNotFound
		default:
			return false, erro.ErrHTTPForbiden
	}

	result := id
	err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
		childLogger.Error().Err(err).Msg("error no ErrUnmarshal")
		return false, errors.New(err.Error())
    }

	return result, nil
}

func makePost(url string, inter interface{}, data interface{}) (interface{}, error) {
	childLogger.Debug().Msg("makePost")

	client := &http.Client{Timeout: time.Second * 29}
	
	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(data)

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return false, errors.New(err.Error())
	}

	req.Header.Add("Content-Type", "application/json;charset=UTF-8");

	resp, err := client.Do(req)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Do Request")
		return false, errors.New(err.Error())
	}

	childLogger.Debug().Int("StatusCode :", resp.StatusCode).Msg("")
	switch (resp.StatusCode) {
		case 401:
			return false, erro.ErrHTTPForbiden
		case 403:
			return false, erro.ErrHTTPForbiden
		case 200:
		case 400:
			return false, erro.ErrNotFound
		case 404:
			return false, erro.ErrNotFound
		default:
			return false, erro.ErrHTTPForbiden
	}

	result := inter
	err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
		childLogger.Error().Err(err).Msg("error no ErrUnmarshal")
		return false, errors.New(err.Error())
    }

	return result, nil
}
