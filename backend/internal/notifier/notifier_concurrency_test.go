package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

type collectingTransport struct {
	want     int
	arrived  int
	mu       sync.Mutex
	release  chan struct{}
	problems chan<- string
}

func (t *collectingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	t.mu.Lock()
	t.arrived++
	if t.arrived == t.want {
		close(t.release)
	}
	t.mu.Unlock()
	<-t.release

	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.problems <- fmt.Sprintf("read body: %v", err)
	} else {
		var got struct {
			Title string `json:"title"`
			Body  string `json:"body"`
		}
		if err := json.Unmarshal(body, &got); err != nil {
			t.problems <- fmt.Sprintf("invalid JSON: %v", err)
		} else if !strings.HasPrefix(got.Body, got.Title+":") {
			t.problems <- fmt.Sprintf("mixed message: title %q, body prefix %q", got.Title, got.Body[:min(len(got.Body), 32)])
		}
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("")),
		Request:    r,
	}, nil
}

func TestHTTPSenderConcurrentSendKeepsRequestBodiesIsolated(t *testing.T) {
	const sends = 48

	problems := make(chan string, sends)
	transport := &collectingTransport{want: sends, release: make(chan struct{}), problems: problems}
	oldClient := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: transport}
	defer func() { http.DefaultClient = oldClient }()

	sender := &HTTPSender{Name: "WEBHOOK", URL: "http://webhook.test/notify", Mode: "real"}
	start := make(chan struct{})
	errCh := make(chan error, sends)
	var wg sync.WaitGroup
	for n := 0; n < sends; n++ {
		n := n
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			title := fmt.Sprintf("notice-%02d", n)
			errCh <- sender.Send(context.Background(), Message{
				Title: title,
				Body:  title + ":" + strings.Repeat(string(rune('a'+n%26)), 64*1024),
			})
		}()
	}
	close(start)
	wg.Wait()
	close(errCh)
	close(problems)

	for err := range errCh {
		if err != nil {
			t.Fatalf("Send returned an error: %v", err)
		}
	}
	for problem := range problems {
		t.Error(problem)
	}
}
