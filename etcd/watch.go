package etcd

import (
	"errors"
)

// Errors introduced by the Watch command.
var (
	ErrWatchStoppedByUser = errors.New("Watch stopped by the user via stop channel")
)

// Watch any change under the given prefix.
// When a sinceIndex is given, watch will try to scan from that index to the last index
// and will return any changes under the given prefix during the history
// If a receiver channel is given, it will be a long-term watch. Watch will block at the
// channel. And after someone receive the channel, it will go on to watch that prefix.
// If a stop channel is given, client can close long-term watch using the stop channel

func (c *Client) WatchDir(prefix string, receiver chan *Response, stop chan bool) (*Response, error) {
	return c.watch(prefix, 0, true, receiver, stop)
}

func (c *Client) WatchDirIndex(prefix string, waitIndex uint64, receiver chan *Response, stop chan bool) (*Response, error) {
	return c.watch(prefix, waitIndex, true, receiver, stop)
}

func (c *Client) Watch(prefix string, receiver chan *Response, stop chan bool) (*Response, error) {
	return c.watch(prefix, 0, false, receiver, stop)
}

func (c *Client) WatchIndex(prefix string, waitIndex uint64, receiver chan *Response, stop chan bool) (*Response, error) {
	return c.watch(prefix, waitIndex, false, receiver, stop)
}

func (c *Client) watch(prefix string, waitIndex uint64, dir bool, receiver chan *Response, stop chan bool) (*Response, error) {
	logger.Debugf("watch %s [%s]", prefix, c.cluster.Leader)
	if receiver == nil {
		return c.watchOnce(prefix, waitIndex, dir, stop)
	} else {
		for {
			resp, err := c.watchOnce(prefix, waitIndex, dir, stop)
			if resp != nil {
				waitIndex = resp.Index + 1
				receiver <- resp
			} else {
				return nil, err
			}
		}
	}

	return nil, nil
}

// helper func
// return when there is change under the given prefix
func (c *Client) watchOnce(key string, waitIndex uint64, dir bool, stop chan bool) (*Response, error) {

	respChan := make(chan *Response)
	errChan := make(chan error)

	go func() {
		options := Options{
			"wait": true,
		}
		if waitIndex > 0 {
			options["waitIndex"] = waitIndex
		}
		if dir {
			options["recursive"] = true
		}

		resp, err := c.get(key, options)

		if err != nil {
			errChan <- err
		}

		respChan <- resp
	}()

	select {
	case resp := <-respChan:
		return resp, nil
	case err := <-errChan:
		return nil, err
	case <-stop:
		return nil, ErrWatchStoppedByUser
	}
}
