package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type cotacao struct {
	Bid string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	log.Println("Requisição iniciada.")
	defer log.Println("Requisição finalizada.")

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		select {
		case <-req.Context().Done():
			if req.Context().Err() == context.DeadlineExceeded {
				log.Println("Solicitação cancelada: excedeu o tempo limite de 300ms. Error: ", err.Error())
			} else if req.Context().Err() == context.Canceled {
				log.Println("Solicitação cancelada pelo cliente. Error : ", err.Error())
			}
		default:
			log.Println("Erro ao fazer a solicitação: ", err.Error())
		}
		return
	}
	defer res.Body.Close()
	log.Println("Solicitação processada com sucesso.")

	data, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var cotacao cotacao
	if err = json.Unmarshal(data, &cotacao); err != nil {
		panic(err)
	}

	f, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err := f.WriteString("Dólar: {" + cotacao.Bid + "}\n"); err != nil {
		panic(err)
	}

	log.Println("Valor salvo com sucesso no arquivo cotacao.txt")

}
