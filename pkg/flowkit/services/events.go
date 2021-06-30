/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package services

import (
	"fmt"
	"sync"

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/flow-go-sdk/client"

	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
)

// Events is a service that handles all event-related interactions.
type Events struct {
	gateway gateway.Gateway
	state   *flowkit.State
	logger  output.Logger
}

// NewEvents returns a new events service.
func NewEvents(
	gateway gateway.Gateway,
	state *flowkit.State,
	logger output.Logger,
) *Events {
	return &Events{
		gateway: gateway,
		state:   state,
		logger:  logger,
	}
}


func (e * Events) CalculateStartEnd(start uint64, end uint64, last uint64) (uint64, uint64, error) {

	if start == 0 {
		latestBlock, err := e.gateway.GetLatestBlock()
		if err != nil {
			return 0,0, err
		}
		return latestBlock.Height-last, latestBlock.Height, nil
	}

	if end == 0 {
		latestBlock, err := e.gateway.GetLatestBlock()
		if err != nil {
			return 0,0, err
		}
		return start, latestBlock.Height, nil
	}

	return start, end, nil
}

func makeEventQueries(events[] string, startHeight uint64, endHeight uint64, blockCount uint64) []client.EventRangeQuery {
	var queries []client.EventRangeQuery
	for startHeight <= endHeight {
		suggestedEndHeight := startHeight + blockCount -1 //since we are inclusive
		endHeight := endHeight
		if suggestedEndHeight < endHeight {
			endHeight = suggestedEndHeight
		}
		for _, event := range events {
			queries = append(queries, client.EventRangeQuery{
				Type:        event,
				StartHeight: startHeight,
				EndHeight:   endHeight,
			})
		}
		startHeight = suggestedEndHeight + 1
	}
	return queries

}
func (e *Events) Get(events []string, startHeight uint64, endHeight uint64, blockCount uint64, workerCount int) ([]client.BlockEvents, error) {
	e.logger.StartProgress("Fetching Events...")
	defer e.logger.StopProgress()

	//why do I need to substract one here, because block searches are inclusive.
	queries := makeEventQueries(events, startHeight, endHeight, blockCount)

	jobChan := make(chan client.EventRangeQuery, workerCount)
	results := make(chan EventWorkerResult)

	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e.eventWorker(jobChan, results)
		}()
	}

	// wait on the workers to finish and close the result channel
	// to signal downstream that all work is done
	go func() {
		defer close(results)
		wg.Wait()
	}()

	go func() {
		defer close(jobChan)
		for _, query := range queries {
			jobChan <- query
		}
	}()

	var resultEvents []client.BlockEvents
	for eventResult := range results {
		if eventResult.Error != nil {
			return nil, eventResult.Error
		}

		resultEvents = append(resultEvents, eventResult.Events...)
	}
	return resultEvents, nil

}

func (e *Events) eventWorker(jobChan <-chan client.EventRangeQuery, results chan<- EventWorkerResult) {
	for q := range jobChan {
		e.logger.Debug(fmt.Sprintf("Fetching events %v", q))
		blockEvents, err := e.gateway.GetEvents(q.Type, q.StartHeight, q.EndHeight)
		if err != nil {
			results <- EventWorkerResult{nil, err}
		}
		results <- EventWorkerResult{blockEvents, nil}
	}
}


type EventWorkerResult struct {
	Events []client.BlockEvents
	Error  error
}