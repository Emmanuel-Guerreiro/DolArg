package main

import (
	"fmt"
	"net/http"

	"github.com/Jeffail/gabs/v2"
	xj "github.com/basgys/goxml2json"
)

//Any new path from dolarSi must be added here.
//The map is: {request path: DolarSi xml parsed path}
var DolarSiPaths map[string]string = map[string]string{
	"oficial":         "cotiza.Dolar.casa344",
	"blue":            "Dolar.casa380",
	"BNABillete":      "Dolar.casa47",
	"BCRAReferencia":  "Dolar.casa49",
	"MayoristaBancos": "Dolar.casa44",
}

type NonValidPath struct{}

func (m *NonValidPath) Error() string {
	return "Non valid path"
}

func fetchDolarSi() (*http.Response, error) {
	resp, err := http.Get("https://www.dolarsi.com/api/dolarSiInfo.xml")

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func parseDolarSi() (*gabs.Container, error) {

	resp, err := fetchDolarSi()

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	j, err := xj.Convert(resp.Body)
	if err != nil {
		return nil, err
	}
	parsed, err := gabs.ParseJSON(j.Bytes())

	if err != nil {
		return nil, err
	}

	return parsed, nil
}

//This can access any value inside the xml parsed
//Isnt bounded to dolar value
func DolarSiValue(path string) (string, error) {
	var value string
	var ok bool

	parsed, err := parseDolarSi()

	if err != nil {
		return "", err
	}

	value, ok = parsed.Path(path).Data().(string)

	if !ok {
		return "", &NonValidPath{}
	}

	return value, nil
}

//This will bound DolarSiValue to numeric and returns
//buy and sell value at the same time. Making clearer
//error handling. [Buy, Sell]
func DolarSiBuySell(path string) ([]string, error) {
	buyPath := fmt.Sprintf("%s%s", path, ".compra")
	sellPath := fmt.Sprintf("%s%s", path, ".venta")
	buy, err := DolarSiValue(buyPath)

	if err != nil {
		return []string{""}, err
	}

	sell, err := DolarSiValue(sellPath)

	if err != nil {
		return []string{""}, err
	}

	return []string{buy, sell}, nil
}
