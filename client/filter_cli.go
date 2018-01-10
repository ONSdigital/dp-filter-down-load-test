package client

import (
	"fmt"
	"bytes"
	"io/ioutil"
)

import (
	"github.com/ONSdigital/go-ns/log"

	"github.com/ONSdigital/go-ns/clients/filter"
	"net/http"
	"github.com/pkg/errors"
	"encoding/json"
)

type Filter struct {
	HttpClient http.Client
	Host       string
}

func (f Filter) UpdateBlueprint(m filter.Model) (mdl filter.Model, err error) {
	b, err := json.Marshal(m)
	if err != nil {
		return
	}

	uri := fmt.Sprintf("%s/filters/%s?submitted=true", f.Host, m.FilterID)

	req, err := http.NewRequest("PUT", uri, bytes.NewBuffer(b))
	if err != nil {
		return
	}

	resp, err := f.HttpClient.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		return m, errors.Errorf("incorrect status: expected 200, actual %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if err = json.Unmarshal(b, &m); err != nil {
		return
	}

	return m, nil
}

func (f Filter) CreateBlueprint(instanceID string, names []string) (string, error) {
	fj := filter.Model{InstanceID: instanceID}

	var dimensions []filter.ModelDimension
	for _, name := range names {
		dimensions = append(dimensions, filter.ModelDimension{Name: name})
	}

	fj.Dimensions = dimensions

	b, err := json.Marshal(fj)
	if err != nil {
		return "", err
	}

	uri := f.Host + "/filters"

	resp, err := f.HttpClient.Post(uri, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated {
		return "", errors.New("invalid status from filter api")
	}
	defer resp.Body.Close()

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err = json.Unmarshal(b, &fj); err != nil {
		return "", err
	}

	return fj.FilterID, nil
}

func (f Filter) GetOutput(outputID string) (m filter.Model, err error) {
	uri := fmt.Sprintf("%s/filter-outputs/%s", f.Host, outputID)

	resp, err := f.HttpClient.Get(uri)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Error(errors.Errorf("incorrect status code: ", resp.StatusCode), nil)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = json.Unmarshal(b, &m)
	return
}
