package etcd

import (
	"fmt"
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	c := NewClient(nil)

	go setHelper("bar", c)

	result, err := c.WatchDir("watch_foo", nil, nil)

	if err != nil || result.Key != "/watch_foo/foo" || result.Value != "bar" {
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("Watch failed with %s %s %v %v", result.Key, result.Value, result.TTL, result.Index)
	}

	go setHelper("bar", c)

	result, err = c.WatchDirIndex("watch_foo", result.Index, nil, nil)

	if err != nil || result.Key != "/watch_foo/foo" || result.Value != "bar" {
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("Watch with Index failed with %s %s %v %v", result.Key, result.Value, result.TTL, result.Index)
	}

	ch := make(chan *Response, 10)
	stop := make(chan bool, 1)

	go setLoop("bar", c)

	go receiver(ch, stop)

	_, err = c.WatchDir("watch_foo", ch, stop)
	if err != ErrWatchStoppedByUser {
		t.Fatalf("Watch returned a non-user stop error")
	}
}

func setHelper(value string, c *Client) {
	time.Sleep(time.Second)
	c.Set("watch_foo/foo", value, 100)
}

func setLoop(value string, c *Client) {
	time.Sleep(time.Second)
	for i := 0; i < 10; i++ {
		newValue := fmt.Sprintf("%s_%v", value, i)
		c.Set("watch_foo/foo", newValue, 100)
		time.Sleep(time.Second / 10)
	}
}

func receiver(c chan *Response, stop chan bool) {
	for i := 0; i < 10; i++ {
		<-c
	}
	stop <- true
}
