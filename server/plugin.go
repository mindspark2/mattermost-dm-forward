package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/mattermost/mattermost/server/public/plugin"
)

const pluginID = "com.mst.forward-private"

type ForwardPrivatePlugin struct {
	plugin.MattermostPlugin

	configurationLock sync.RWMutex
	configuration     *configuration
}

func (p *ForwardPrivatePlugin) OnActivate() error {
	if err := p.OnConfigurationChange(); err != nil {
		return err
	}
	if err := p.registerSlashCommand(); err != nil {
		return fmt.Errorf("register slash command: %w", err)
	}
	return nil
}

func (p *ForwardPrivatePlugin) OnDeactivate() error {
	_ = p.API.UnregisterCommand("", "forwardprivate")
	return nil
}

func (p *ForwardPrivatePlugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, r *http.Request) {
	path := normalizePluginHTTPPath(r.URL.Path)
	switch {
	case path == "/api/v1/health":
		p.writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": "0.1.2"})
	case path == "/api/v1/action/start":
		p.handleForwardStart(w, r)
	case path == "/api/v1/action/open-dialog":
		p.handleForwardOpenDialog(w, r)
	case path == "/api/v1/dialog/submit":
		p.handleForwardDialogSubmit(w, r)
	case path == "/api/v1/lookup/users":
		p.handleUserLookup(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (p *ForwardPrivatePlugin) pluginURL(path string) string {
	cfg := p.API.GetConfig()
	site := ""
	if cfg.ServiceSettings.SiteURL != nil {
		site = strings.TrimRight(*cfg.ServiceSettings.SiteURL, "/")
	}
	return site + "/plugins/" + pluginID + path
}

func (p *ForwardPrivatePlugin) pluginRelativeURL(path string) string {
	return "/plugins/" + pluginID + path
}

func normalizePluginHTTPPath(path string) string {
	for _, prefix := range []string{"/plugins/" + pluginID, pluginID} {
		if strings.HasPrefix(path, prefix) {
			path = strings.TrimPrefix(path, prefix)
			break
		}
	}
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func (p *ForwardPrivatePlugin) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (p *ForwardPrivatePlugin) enabled() bool {
	return p.getConfiguration().Enabled
}
