package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

type CotacaoDB struct {
	ID         int `gorm:"primaryKey"`
	CotacaoAPI `gorm:"CotacaoAPI"`
}

func main() {

	fmt.Println("Iniciado ")
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", CotacaoHandler)
	http.ListenAndServe(":8080", mux)

}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {

	data, err := GetCotacao()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"Error":"` + err.Error() + `"}`))
		return
	}
	//criando header do tipo json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	//retornando json na resposta da requisição
	json.NewEncoder(w).Encode(data)

}

func GetCotacao() (*CotacaoAPI, error) {

	//criando o contexto com def de timeout de 200ms
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	//criando requisição
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		panic(err)
	}
	//executando requisição
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer requisição: %v\n", err)
		//se timeout excedido: registrar log e devolver erro
		select {
		case <-ctx.Done():
			log.Println("Request cancelada por timeout do contexto da requisição de https://economia.awesomeapi.com.br/json/last/USD-BRL")
		default:
		}
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New("Erro " + fmt.Sprint(resp.StatusCode))
		return nil, err
	}

	//obtendo corpo da requisição
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler body: %v\n", err)
		return nil, err
	}

	//transformando corpo na struct CotacaoAPI
	var data CotacaoAPI
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Não foi possivel converter o json recebido: %v\n", err)
		return nil, err
	}

	//registrando cotação no banco de dados
	err = saveData(data)

	return &data, err

}

func saveData(data CotacaoAPI) error {

	//fazendo conexão com banco de dados
	db, err := gorm.Open(sqlite.Open("cotacoes.db"), &gorm.Config{})
	if err != nil {
		return err
	}
	//criando tabela
	db.AutoMigrate(&CotacaoDB{})

	//contexto para definir timeout de 10ms
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*10)
	defer cancel()
	select {
	case <-ctx.Done():
		log.Println("Conexão com banco de dados cancelada por timeout!")
		return errors.New("conexão com banco de dados cancelada por timeout")
	default:
	}
	//inserindo nova cotação
	db.WithContext(ctx).Create(&CotacaoDB{
		CotacaoAPI: CotacaoAPI{
			Usdbrl: Usdbrl{
				Code:       data.Usdbrl.Code,
				Codein:     data.Usdbrl.Codein,
				Name:       data.Usdbrl.Name,
				High:       data.Usdbrl.High,
				Low:        data.Usdbrl.Low,
				VarBid:     data.Usdbrl.VarBid,
				PctChange:  data.Usdbrl.PctChange,
				Bid:        data.Usdbrl.Bid,
				Ask:        data.Usdbrl.Ask,
				Timestamp:  data.Usdbrl.Timestamp,
				CreateDate: data.Usdbrl.CreateDate,
			},
		}})

	return nil
}
