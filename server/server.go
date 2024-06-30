package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Usdbrl struct {
	Id         uint64 `gorm:"primaryKey;autoIncrement"`
	Code       string `json:"code" gorm:"primaryKey"`
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
}

func main() {

	//json.NewEncoder(os.Stdout).Encode(cotacao())
	http.HandleFunc("/cotacao", cotacaoHandler)
	http.ListenAndServe(":8080", nil)

}

func cotacaoHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	log.Println("Requisição iniciada.")
	defer cancel()

	defer log.Println("Requisição finalizada.")

	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	r, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(r)

	if err != nil {
		select {
		case <-r.Context().Done():
			if r.Context().Err() == context.DeadlineExceeded {
				log.Println("Solicitação cancelada: excedeu o tempo limite de 200ms")
				http.Error(w, "TimeOut da solicitação", http.StatusRequestTimeout)

			} else if r.Context().Err() == context.Canceled {
				log.Println("Solicitação cancelada pelo cliente")
				http.Error(w, "Solicitação cancelada pelo cliente", 499)

			}
		default:
			log.Println("Erro ao fazer a solicitação: ", err)
			http.Error(w, "Erro ao fazer a solicitação", http.StatusInternalServerError)

		}
		return
	}
	defer res.Body.Close()

	log.Println("Solicitação processada com sucesso.")

	data, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var dataBody map[string]json.RawMessage
	if err = json.Unmarshal(data, &dataBody); err != nil {
		panic(err)
	}

	var cota Usdbrl
	err = json.Unmarshal(dataBody["USDBRL"], &cota)
	if err != nil {
		panic(err)
	}

	saveCotation(mysqlCotation(), &cota)

	jsonBid := map[string]string{"Bid": cota.Bid}
	json.NewEncoder(os.Stdout).Encode(&jsonBid)
	err = json.NewEncoder(w).Encode(&jsonBid)
	if err != nil {
		panic(err)
	}

}

func mysqlCotation() *gorm.DB {
	dsn := "root:root@tcp(localhost:3306)/goexpert"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&Usdbrl{})

	return db

}

func saveCotation(db *gorm.DB, cotacao *Usdbrl) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := db.WithContext(ctx).Create(&cotacao).Error

	if err != nil {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				log.Println("Erro ao gravar cotação: excedeu o tempo limite de 10ms")
				return
			} else if ctx.Err() == context.Canceled {
				log.Println("Solicitação cancelada pelo cliente")
				return
			}
		default:
			log.Printf("Erro ao gravar cotação: %s", err)

		}
		return
	}

}
