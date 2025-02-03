package rod

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/go-rod/rod/lib/proto"
	"github.com/ysmood/gson"
	"net/http"
	"runtime"
	"strings"
	"sync"
)

type TransportRod struct {
	browser        *rod.Browser
	browserBuilder sync.Once
}

func (t TransportRod) Close() error {
	if t.browser == nil {
		return nil
	}
	return t.browser.Close()
}

func NewTransport() *TransportRod {
	return &TransportRod{}
}

// SetExtraHeaders whether to always send extra HTTP headers with the requests from this page.
func setExtraHeaders(p *rod.Page, headers http.Header) (func(), error) {
	gsonHeaders := proto.NetworkHeaders{}

	for key, values := range headers {
		gsonHeaders[key] = gson.New(strings.Join(values, ";"))
	}

	return p.EnableDomain(&proto.NetworkEnable{}), proto.NetworkSetExtraHTTPHeaders{Headers: gsonHeaders}.Call(p)
}

func (t TransportRod) RoundTrip(request *http.Request) (*http.Response, error) {
	t.browserBuilder.Do(func() {
		u := launcher.New().Leakless(runtime.GOOS == "linux").Set(flags.Headless, "new").MustLaunch()
		t.browser = rod.New().ControlURL(u).MustConnect()
	})
	defer t.browser.MustClose();

	page, err := t.browser.Context(request.Context()).Page(proto.TargetCreateTarget{URL: ""})

	if err != nil {
		return nil, err
	}

	_, err = page.SetExtraHeaders([]string{"Referer", request.Header.Get("Referer")})
	if err != nil {
		return nil, err
	}

	err = page.Navigate(request.URL.String())
	if err != nil {
		return nil, err
	}
	page.WaitNavigation(proto.PageLifecycleEventNameNetworkAlmostIdle)()

	if err != nil {
		return nil, err
	}
	return &http.Response{Body: newPageReader(page),
		StatusCode: 200,
		Header:     map[string][]string{"Content-Type": {"text/html"}},
	}, nil
}
