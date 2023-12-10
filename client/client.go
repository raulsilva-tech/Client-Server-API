package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Usdbrl struct {
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
}

type CotacaoAPI struct {
	Usdbrl `json:"USDBRL"`
}

func main() {

	//criando o contexto com def de timeout de 300ms
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	//se timeout da requisição for excedido: registre log
	select {
	case <-ctx.Done():
		log.Println("Request cancelada por timeout do contexto.")
	default:
	}

	//criando requisição
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}
	//executando requisição
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	//obtendo corpo da requisição
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		fmt.Println("StatusCode Error: ", resp.StatusCode, string(bodyBytes))
		return
	}
	defer resp.Body.Close()

	//exibindo retorno json da requisição
	fmt.Println(string(bodyBytes))

	//transformando corpo na struct Cotacao
	var data CotacaoAPI
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Não foi possivel converter o json recebido: %v\n", err)
		panic(err)
	}

	//obtendo campo Bid
	bid := data.Usdbrl.Bid

	//registrando cotação no arquivo texto
	err = logCotacao(bid)
	if err == nil {
		fmt.Println("Cotação registrada com sucesso no arquivo cotacao.txt")
	}
}

func logCotacao(bid string) error {

	file, err := os.Create("cotacao.txt")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Não foi possivel criar o arquivo: %v\n", err)
		return err
	}
	_, err = file.WriteString("Dólar: " + bid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Não foi possivel criar o arquivo: %v\n", err)
		return err
	}
	defer file.Close()
	return nil
}
