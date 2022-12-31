package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type CotacaoResponse struct {
	Dolar float64 `json:"Dolar"`
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	req.Header.Set("Accept", "application/json")
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	io.Copy(os.Stdout, res.Body)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	var c CotacaoResponse
	res1, err := http.DefaultClient.Do(req)
	body, error := io.ReadAll(res1.Body)
	defer res1.Body.Close()

	if error != nil {
		panic(error)
	}
	err = json.Unmarshal(body, &c)
	file, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	_, err = file.WriteString(fmt.Sprintf("Dolar: %v", c.Dolar))

}
