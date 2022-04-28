package front

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

//var GUploadRequestUnsignedContract []http_utils.Contract

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	//muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	////muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	//muxRouter.HandleFunc("/contract/upload-request", handleUploadrequest).Methods("POST")

	for _, route := range ALLRoutes {
		muxRouter.HandleFunc(route.routeName, route.HandlerFunc).Methods(route.Method)
	}
	return muxRouter
}

func HttpApiRun(httpPort string) error {

	mux := makeMuxRouter()
	//httpPort := os.Getenv("PORT")
	//httpPort := "8080"
	log.Println("HTTP Server Listening on port :", httpPort)
	s := &http.Server{
		Addr:           ":" + httpPort,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}


