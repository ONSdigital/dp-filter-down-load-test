package main

import (
	"github.com/ONSdigital/go-ns/log"

	"github.com/ONSdigital/go-ns/clients/filter"
	"os"
	"time"
	"fmt"
	"net/http"
	"sync"
	"github.com/ONSdigital/down-load-test/client"
	"sort"
	"github.com/ONSdigital/down-load-test/config"
	"encoding/json"
	"io/ioutil"
)

var wg sync.WaitGroup

type FilterDownloadTask struct {
	RequestNo  int
	Host       string
	FilterCli  client.Filter
	Blueprint  filter.Model
	InstanceID string
}

type result struct {
	RequestNo int
	Time      float64
}

func main() {
	cfg := config.Load()

	startTime := time.Now()
	var totalTime float64

	results := make([]result, 0)

	allJobsCompleted := false
	filterCli := client.Filter{HttpClient: http.Client{}, Host: cfg.FilterAPIHost}

	for i, file := range cfg.Filters {
		b, _ := ioutil.ReadFile("filters/" + file)
		var f filter.ModelDimension
		json.Unmarshal(b, &f)

		task := FilterDownloadTask{
			InstanceID: cfg.InstanceID,
			RequestNo:  i,
			FilterCli:  filterCli,
			Blueprint: filter.Model{
				InstanceID: cfg.InstanceID,
				Dimensions: []filter.ModelDimension{f},
			},
		}

		wg.Add(1)
		go func() {
			res := task.filterDownload()
			results = append(results, res)
		}()
	}

	go func() {
		log.Info("starting filter download test", nil)
		for !allJobsCompleted {
			fmt.Print(".")
			<-time.After(time.Second * 1)
		}
	}()

	wg.Wait()
	totalTime = time.Since(startTime).Seconds()
	allJobsCompleted = true

	sort.Slice(results, func(i, j int) bool {
		return results[i].Time < results[j].Time
	})

	fmt.Println("")
	log.Info("total time to complete all downloads", log.Data{"totalTime": totalTime})
	for _, res := range results {
		fmt.Println(fmt.Sprintf("requestNo: %d, time: %f", res.RequestNo, res.Time))
	}
}

func (t *FilterDownloadTask) filterDownload() result {
	var err error
	var filterID string

	filterID, err = t.FilterCli.CreateBlueprint(t.InstanceID, []string{"geography", "age", "sex", "time"})
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	startTime := time.Now()
	var totalTime float64

	t.Blueprint.FilterID = filterID
	t.Blueprint, err = t.FilterCli.UpdateBlueprint(t.Blueprint)

	if err != nil {
		log.ErrorC("update blueprint error", err, nil)
		os.Exit(1)
	}

	outputID := t.Blueprint.Links.FilterOutputs.ID
	var output filter.Model

	done := false
	for !done {
		output, err = t.FilterCli.GetOutput(outputID)
		if err != nil {
			log.ErrorC("get output error", err, nil)
			os.Exit(1)
		}

		if output.State == "completed" {
			totalTime = time.Since(startTime).Seconds()
			wg.Done()
			done = true
		} else {
			// poll for a result every 1 second
			<-time.After(time.Second * 1)
		}
	}
	return result{RequestNo: t.RequestNo, Time: totalTime}
}
