package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	URL_CEP_PATTERN_BRASIL  = "https://brasilapi.com.br/api/cep/v1/{cep}"
	URL_CEP_PATTERN_VIA_CEP = "http://viacep.com.br/ws/{cep}/json/"
	URL_CEP_REPLACE_PATTERN = "{cep}"
)

type respCepApi struct {
	Origin       string `json:"origin"`
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
}

type respViaCepApi struct {
	Cep         string `json:"cep"`
	Uf          string `json:"uf"`
	Localidade  string `json:"localidade"`
	Bairro      string `json:"bairro"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Unidade     string `json:"unidade"`
	Estado      string `json:"estado"`
	Regiao      string `json:"regiao"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type respBrasilApi struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

func (rvcapi *respViaCepApi) paserToRespCepApi() respCepApi {
	return respCepApi{
		Origin:       "Via CEP API",
		Cep:          rvcapi.Cep,
		State:        rvcapi.Uf,
		City:         rvcapi.Localidade,
		Neighborhood: rvcapi.Bairro,
		Street:       rvcapi.Logradouro,
	}
}

func (rbapi *respBrasilApi) paserToRespCepApi() respCepApi {
	return respCepApi{
		Origin:       "Brasil API",
		Cep:          rbapi.Cep,
		State:        rbapi.State,
		City:         rbapi.City,
		Neighborhood: rbapi.Neighborhood,
		Street:       rbapi.Street,
	}
}

func main() {
	arg_cep := flag.String("cep", "", "cep (formato: 00000-000)")
	flag.Parse()

	if !isCepValid(*arg_cep) {
		println(fmt.Sprintf("CEP inválido: %s", *arg_cep))
		return
	}

	chanCep := make(chan respCepApi, 1)
	chanErr := make(chan error, 1)

	go getCepInfoByBrasilAPi(*arg_cep, chanCep, chanErr)
	go getCepInfoByViaCepAPi(*arg_cep, chanCep, chanErr)

	select {
	case resp := <-chanCep:
		jsonData, err := json.Marshal(resp)
		if err != nil {
			print("TEMOS UM ERRO, APÓS A CHAMADA DE API", err.Error())
		}

		println(string(jsonData))
		return
	case err := <-chanErr:
		print("TEMOS UM ERRO", err.Error())
		return
	case <-time.After(1 * time.Second):
		fmt.Println("TIMEOUT 1s")
		return
	}
}

func isCepValid(arg_cep string) bool {
	return regexp.MustCompile(`^\d{5}-\d{3}$`).MatchString(arg_cep)
}

func getCepInfoByBrasilAPi(cep string, chanCep chan respCepApi, chanErr chan error) {
	time.Sleep(1200 * time.Millisecond)
	url, err := applyCepToUrlApi(cep, URL_CEP_PATTERN_BRASIL)
	if err != nil {
		chanErr <- err
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		chanErr <- err
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		chanErr <- err
		return
	}

	respCepApi := respBrasilApi{}
	err = json.Unmarshal([]byte(body), &respCepApi)
	if err != nil {
		chanErr <- err
		return
	}

	chanCep <- respCepApi.paserToRespCepApi()
}

func getCepInfoByViaCepAPi(cep string, chanCep chan respCepApi, chanErr chan error) {
	//time.Sleep(1500 * time.Millisecond)
	url, err := applyCepToUrlApi(cep, URL_CEP_PATTERN_VIA_CEP)
	if err != nil {
		chanErr <- err
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		chanErr <- err
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		chanErr <- err
		return
	}

	respCepApi := respViaCepApi{}
	err = json.Unmarshal([]byte(body), &respCepApi)
	if err != nil {
		chanErr <- err
		return
	}

	chanCep <- respCepApi.paserToRespCepApi()
}

func applyCepToUrlApi(cep, urlPattern string) (string, error) {
	url := strings.Replace(urlPattern, URL_CEP_REPLACE_PATTERN, cep, -1)
	return url, nil
}
