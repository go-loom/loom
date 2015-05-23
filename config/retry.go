package config

import (
	"time"
)

type Retry struct {
	Number    int    `json:"number,omitempty"`
	Timeout   string `json:"timeout,omitempty"`
	DelayTime string `json:"delay,omitempty"`
	delayTime *time.Duration
	timeout   *time.Duration
	NumRetry  int `json:"-"`
}

func (r *Retry) IncrRetry() bool {
	r.NumRetry++
	if r.NumRetry >= r.Number {
		return false
	}
	return true
}

func (r *Retry) GetTimeout() (*time.Duration, error) {
	if r.timeout != nil {
		return r.timeout, nil
	}
	if r.Timeout != "" {
		d, err := time.ParseDuration(r.Timeout)
		if err != nil {
			return nil, err
		}
		r.timeout = &d
		return r.timeout, err
	}

	return nil, nil
}

func (r *Retry) GetDelayTime() (time.Duration, error) {
	if r.delayTime != nil {
		return *r.delayTime, nil
	}
	if r.DelayTime != "" {
		d, err := time.ParseDuration(r.DelayTime)
		if err != nil {
			return time.Duration(0), err
		}
		r.delayTime = &d
		return d, err
	}

	d := time.Duration(0 * time.Second)
	r.delayTime = &d
	return *r.delayTime, nil
}
