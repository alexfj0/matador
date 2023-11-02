package handler

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
    "strconv"

	"github.com/kodylow/matador/pkg/auth"
	models "github.com/kodylow/matador/pkg/models"
	"github.com/kodylow/matador/pkg/service"
)

var APIKey string
var APIRoot string
var globalPrice uint64

// Init initializes data for the handler
func Init(price string, root string, lnAddress string) error {

    var err error
    globalPrice, err = strconv.ParseUint(price, 10, 64)
    if err != nil {
        return fmt.Errorf("error converting global price", err)
    }

	APIRoot = root
	service.LnAddr, err = service.GetCallback(lnAddress)
	if err != nil {
		return fmt.Errorf("error getting lnaddress callback: %w", err)
	}
	return nil
}

// PassthroughHandler forwards the request to the OpenAI API
func PassthroughHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("passthroughHandler started")

	// Read the body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading body:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset the body to its original state

	// Create a RequestInfo
	reqInfo := models.RequestInfo{
		AuthHeader: r.Header.Get("Authorization"),
		Method:     r.Method,
		Path:       r.URL.Path,
		Body:       body,
	}

	err = auth.CheckAuthorizationHeader(reqInfo)
	if err != nil {
		log.Println("Unauthorized, payment required")
		l402, err := auth.GetL402(globalPrice, reqInfo)
		if err != nil {
			log.Println("Error getting L402:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Send 402 Payment Required with the invoice
		w.Header().Set("WWW-Authenticate", l402)
		http.Error(w, "Payment Required", http.StatusPaymentRequired)
		return
	}

	// Create a new request to forward
	req, err := http.NewRequest(r.Method, APIRoot+reqInfo.Path, r.Body)
	if err != nil {
		log.Println("Error creating new forward request:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Forward the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error forwarding the request:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy headers from the response
	for name, values := range resp.Header {
		w.Header()[name] = values
	}

	// Set the status code on the response writer to the status code of the response
	w.WriteHeader(resp.StatusCode)

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Response behind proxy:", string(responseBody))

    // Write the response back to the client
	w.Write(responseBody)

	log.Println("passthroughHandler completed")
}

