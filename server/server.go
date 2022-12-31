package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/driver/sqlite" // Sqlite driver based on GGO
	// "github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"gorm.io/gorm"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Cotacao struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type CotacaoResponse struct {
	Moeda float64 `json:"Dolar"`
}

type CotacaoDB struct {
	Code       string  `json:"code"`
	Codein     string  `json:"codein"`
	Name       string  `json:"name"`
	Bid        float64 `json:"bid"`
	CreateDate string  `json:"create_date"`
	gorm.Model
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", HomeHandler)
	mux.HandleFunc("/cotacao", BuscaCotacaoHandler)
	http.ListenAndServe(":8080", mux)

}

func HomeHandler(writer http.ResponseWriter, r *http.Request) {
	writer.Write([]byte("Desafio Full Cycle!!"))
}
func BuscaCotacaoHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	cotacao, error := BuscarCotacao()
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*300)
	defer cancel()
	InserirCotacao(ctx, cotacao)
	//ctx.with
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	cotacaoResponse := &CotacaoResponse{Moeda: parseFloat(cotacao.Usdbrl.Bid)}

	json.NewEncoder(w).Encode(cotacaoResponse)
}
func InserirCotacao(ctx context.Context, cotacao *Cotacao) {
	db, err := gorm.Open(sqlite.Open("cotacao.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&CotacaoDB{})
	select {
	case <-ctx.Done():
		fmt.Println("Insert data. Timeout reached")
		return
	case <-time.After(10 * time.Millisecond):
		var cotacaodb CotacaoDB
		cotacaodb.Code = cotacao.Usdbrl.Code
		cotacaodb.Codein = cotacao.Usdbrl.Codein
		cotacaodb.Name = cotacao.Usdbrl.Name
		cotacaodb.Bid = parseFloat(cotacao.Usdbrl.Bid)
		cotacaodb.CreateDate = cotacao.Usdbrl.CreateDate
		db.Create(&cotacaodb)
	}

}

func parseFloat(string2 string) float64 {
	parseFloat, err := strconv.ParseFloat(string2, 64)
	if err != nil {
		panic(err)
	}
	return parseFloat
}

func BuscarCotacao() (*Cotacao, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, error := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	req.Header.Set("Accept", "application/json")
	if error != nil {
		panic(error)
	}
	resp, error := http.DefaultClient.Do(req)
	if error != nil {
		panic(error)
	}
	defer resp.Body.Close()
	body, error := io.ReadAll(resp.Body)
	if error != nil {
		panic(error)
	}
	var c Cotacao

	error = json.Unmarshal(body, &c)

	if error != nil {
		return nil, error
	}
	return &c, nil
}
