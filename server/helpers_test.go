package main

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestForwardSourceChannelGate(t *testing.T) {
	p := &ForwardPrivatePlugin{}
	cases := []struct {
		name string
		ch   *model.Channel
		ok   bool
	}{
		{name: "nil", ch: nil, ok: false},
		{name: "direct", ch: &model.Channel{Type: model.ChannelTypeDirect}, ok: true},
		{name: "group", ch: &model.Channel{Type: model.ChannelTypeGroup}, ok: true},
		{name: "open", ch: &model.Channel{Type: model.ChannelTypeOpen}, ok: true},
		{name: "private", ch: &model.Channel{Type: model.ChannelTypePrivate}, ok: true},
		{name: "empty", ch: &model.Channel{Type: ""}, ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := p.isForwardSourceChannel(tc.ch); got != tc.ok {
				t.Fatalf("isForwardSourceChannel(%s) = %v, want %v", tc.name, got, tc.ok)
			}
		})
	}
}
