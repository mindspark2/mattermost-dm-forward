package main

import (
	"fmt"
	"reflect"
)

type configuration struct {
	Enabled                     bool
	MaxRecipients               float64
	ShowAttribution             bool
	PostReceiptInSourceChannel  bool
}

func (p *ForwardPrivatePlugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()
	if p.configuration == nil {
		return &configuration{
			Enabled:                    true,
			MaxRecipients:              8,
			ShowAttribution:            true,
			PostReceiptInSourceChannel: true,
		}
	}
	return p.configuration
}

func (p *ForwardPrivatePlugin) setConfiguration(c *configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()
	if c != nil && p.configuration == nil {
		p.configuration = c
		return
	}
	if c == nil || (p.configuration != nil && reflect.DeepEqual(p.configuration, c)) {
		return
	}
	p.configuration = c
}

func (p *ForwardPrivatePlugin) maxRecipients() int {
	cfg := p.getConfiguration()
	n := int(cfg.MaxRecipients)
	if n < 1 {
		n = 1
	}
	if n > 8 {
		n = 8
	}
	return n
}

func (p *ForwardPrivatePlugin) OnConfigurationChange() error {
	var cfg configuration
	if err := p.API.LoadPluginConfiguration(&cfg); err != nil {
		return fmt.Errorf("load plugin configuration: %w", err)
	}
	p.setConfiguration(&cfg)
	return nil
}
