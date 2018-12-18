package httpclient_test

import (
	"context"

	"inet.af/httpclient"
)

type resT struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}

func Example() {
	ctx := context.Background()
	resTi, err := httpclient.Fetch(new(resT))(ctx, "GET", "http://foo.com/")
	check(err)
	resT := resTi.(*resT) // temporary; won't be required once we have generics
	_ = resT
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
