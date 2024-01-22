package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Dollar struct {
	Bid float64 `json:"bid"`
}

type ApiResponse struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func NewDollar(bid float64) *Dollar {
	return &Dollar{
		Bid: bid,
	}
}

func main() {
	db, err := sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cotacoes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bid FLOAT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
		defer cancel()

		dollar, err := getDollar(ctx)
		if err != nil {
			http.Error(w, "Erro ao obter a cotação do dólar", http.StatusInternalServerError)
			return
		}

		ctx, cancel = context.WithTimeout(r.Context(), 10*time.Millisecond)
		defer cancel()
		err = insertDollar(ctx, db, dollar)
		if err != nil {
			panic(err)
		}

		json.NewEncoder(w).Encode(dollar)
	})

	fmt.Println("Servidor iniciado na porta 8080")
	http.ListenAndServe(":8080", nil)
}

func getDollar(ctx context.Context) (*Dollar, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse ApiResponse

	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		panic(err)
	}

	bid, err := strconv.ParseFloat(apiResponse.USDBRL.Bid, 64)

	if err != nil {
		panic(err)
	}

	var dollar = NewDollar(bid)

	return dollar, nil
}

func insertDollar(ctx context.Context, db *sql.DB, dollar *Dollar) error {
	stmt, err := db.Prepare("INSERT INTO cotacoes (bid) VALUES ($1)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, dollar.Bid)
	if err != nil {
		return err
	}

	return nil
}
