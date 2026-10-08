package main

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost/server/public/model"
)

func (p *ForwardPrivatePlugin) handleForwardStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if !p.enabled() {
		http.Error(w, "disabled", http.StatusForbidden)
		return
	}
	userID, appErr := p.userFromRequest(r)
	if appErr != nil {
		http.Error(w, appErr.Error(), http.StatusUnauthorized)
		return
	}

	var body struct {
		PostID string `json:"post_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.PostID == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	post, channel, appErr := p.canAccessSourcePost(userID, body.PostID)
	if appErr != nil {
		http.Error(w, appErr.Message, http.StatusForbidden)
		return
	}

	ephemeral := &model.Post{
		UserId:    userID,
		ChannelId: post.ChannelId,
		Message:   "To forward this message to other users (separate DMs), click **Choose recipients**.",
		Props: model.StringInterface{
			"attachments": []*model.SlackAttachment{{
				Actions: []*model.PostAction{{
					Id:   "forward_private_open",
					Name: "Choose recipients",
					Type: model.PostActionTypeButton,
					Integration: &model.PostActionIntegration{
						URL: p.pluginRelativeURL("/api/v1/action/open-dialog"),
						Context: model.StringInterface{
							"source_post_id": post.Id,
							"channel_id":     channel.Id,
						},
					},
				}},
			}},
		},
	}
	p.API.SendEphemeralPost(userID, ephemeral)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{}"))
}

func (p *ForwardPrivatePlugin) handleForwardOpenDialog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if !p.enabled() {
		p.writePostActionEphemeral(w, "", "", "Forward Private is disabled.")
		return
	}

	var req model.PostActionIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	sourcePostID := contextString(req.Context, "source_post_id")
	if sourcePostID == "" {
		sourcePostID = req.PostId
	}
	if _, _, appErr := p.canAccessSourcePost(req.UserId, sourcePostID); appErr != nil {
		p.writePostActionEphemeral(w, req.UserId, req.ChannelId, appErr.Message)
		return
	}
	if req.TriggerId == "" {
		p.writePostActionEphemeral(w, req.UserId, req.ChannelId, "Could not open the form. Try again.")
		return
	}

	dialog := p.buildForwardDialog(sourcePostID)
	open := model.OpenDialogRequest{
		TriggerId: req.TriggerId,
		URL:       p.pluginURL("/api/v1/dialog/submit"),
		Dialog:    dialog,
	}
	if appErr := p.API.OpenInteractiveDialog(open); appErr != nil {
		p.writePostActionEphemeral(w, req.UserId, req.ChannelId, "Could not open dialog: "+appErr.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(model.PostActionIntegrationResponse{})
}

func (p *ForwardPrivatePlugin) writePostActionEphemeral(w http.ResponseWriter, userID, channelID, msg string) {
	if userID != "" && channelID != "" && msg != "" {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    userID,
			ChannelId: channelID,
			Message:   msg,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(model.PostActionIntegrationResponse{})
}
