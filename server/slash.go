package main

import (
	"strings"

	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/model"
)

func (p *ForwardPrivatePlugin) registerSlashCommand() error {
	return p.API.RegisterCommand(&model.Command{
		Trigger:          "forwardprivate",
		AutoComplete:     true,
		AutoCompleteDesc: "Forward a DM/GM message to other users (paste message link, or use ⋯ → Forward to…).",
		AutoCompleteHint: "[message permalink]",
	})
}

func (p *ForwardPrivatePlugin) ExecuteCommand(_ *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if !p.enabled() {
		return &model.CommandResponse{Text: "Forward Private is disabled.", ResponseType: model.CommandResponseTypeEphemeral}, nil
	}

	text := strings.TrimSpace(args.Command)
	text = strings.TrimPrefix(text, "/forwardprivate")
	text = strings.TrimSpace(text)

	postID := extractPostIDFromText(text)
	if postID == "" {
		return &model.CommandResponse{
			Text:         "Use **⋯ → Forward to…** on a message in a DM or group chat, or run `/forwardprivate` with a message permalink from that chat.",
			ResponseType: model.CommandResponseTypeEphemeral,
		}, nil
	}

	if _, _, appErr := p.canAccessSourcePost(args.UserId, postID); appErr != nil {
		return &model.CommandResponse{Text: appErr.Message, ResponseType: model.CommandResponseTypeEphemeral}, nil
	}
	if args.TriggerId == "" {
		return &model.CommandResponse{Text: "Could not open dialog.", ResponseType: model.CommandResponseTypeEphemeral}, nil
	}

	dialog := p.buildForwardDialog(postID)
	open := model.OpenDialogRequest{
		TriggerId: args.TriggerId,
		URL:       p.pluginURL("/api/v1/dialog/submit"),
		Dialog:    dialog,
	}
	if appErr := p.API.OpenInteractiveDialog(open); appErr != nil {
		return &model.CommandResponse{Text: "Could not open dialog: " + appErr.Error(), ResponseType: model.CommandResponseTypeEphemeral}, nil
	}
	return &model.CommandResponse{}, nil
}

func extractPostIDFromText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	// Permalink path segment .../pl/{postId}
	if i := strings.Index(text, "/pl/"); i >= 0 {
		rest := text[i+4:]
		for j, c := range rest {
			if c == '?' || c == '/' || c == '#' {
				return rest[:j]
			}
		}
		return rest
	}
	// Raw 26-char id
	if len(text) == 26 && isAlphaNum26(text) {
		return text
	}
	return ""
}

func isAlphaNum26(s string) bool {
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			continue
		}
		return false
	}
	return true
}
