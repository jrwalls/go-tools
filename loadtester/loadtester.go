package loadtester

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

type (
	Result[R any] struct {
		Err             error
		Status          int
		RawResponseBody []byte
		Data            *R
	}

	LoadTester[I, O any] struct {
		url       string
		num       int
		headersFn func(i int) map[string]string
		inputFn   func(i int) I
		stopCh    chan struct{}
		resCh     chan *Result[O]
		wg        sync.WaitGroup
	}
)

func NewLoadTester[I, O any](url string, numRequests int, headersFn func(i int) map[string]string, inputFn func(i int) I) *LoadTester[I, O] {
	return &LoadTester[I, O]{
		url:       url,
		num:       numRequests,
		headersFn: headersFn,
		inputFn:   inputFn,
		stopCh:    make(chan struct{}, 1),
		resCh:     make(chan *Result[O]),
	}
}

func (lt *LoadTester[I, O]) Start(withDelay time.Duration) <-chan *Result[O] {
	lt.wg.Add(lt.num)
	go func() {
	outer:
		for i := 0; i < lt.num; i++ {
			select {
			case <-lt.stopCh:
				lt.wg.Add(-(lt.num - i))
				break outer
			default:
				go lt.worker(i)
				if withDelay > 0 {
					time.Sleep(withDelay)
				}
			}
		}
		lt.wg.Wait()
		close(lt.resCh)
	}()

	return lt.resCh
}

func (lt *LoadTester[I, O]) worker(i int) {
	defer lt.wg.Done()

	select {
	case <-lt.stopCh:
		return
	default:
		lt.resCh <- doRequest[I, O](lt.url, lt.inputFn(i), lt.headersFn(i))
	}
}

func (lt *LoadTester[I, O]) Stop() {
	close(lt.stopCh)
}

func doRequest[I, O any](url string, in I, headers map[string]string) *Result[O] {
	b, err := json.Marshal(in)
	if err != nil {
		return &Result[O]{Err: err}
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return &Result[O]{Err: err}
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &Result[O]{Err: err}
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Result[O]{Err: err}
	}

	var out O
	if err := json.Unmarshal(body, &out); err != nil {
		return &Result[O]{Err: err, Status: resp.StatusCode, RawResponseBody: body}
	}

	return &Result[O]{Data: &out, Status: resp.StatusCode, RawResponseBody: body}
}
