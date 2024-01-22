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

type Dollar struct {
	Bid float64 `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	dollar, err := getDollar(ctx)
	if err != nil {
		panic(err)
	}

	err = SaveDollarPriceTxt(dollar)
	if err != nil {
		panic(err)
	}
}

func getDollar(ctx context.Context) (Dollar, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return Dollar{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Dollar{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Dollar{}, err
	}

	var dollar Dollar
	err = json.Unmarshal(body, &dollar)
	if err != nil {
		return Dollar{}, err
	}

	return dollar, nil
}

func SaveDollarPriceTxt(dollar Dollar) error {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "Dólar: %.2f", dollar.Bid)
	if err != nil {
		return err
	}

	return nil
}
