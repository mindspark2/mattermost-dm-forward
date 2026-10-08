package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
)

func (p *ForwardPrivatePlugin) buildForwardDialog(sourcePostID string) model.Dialog {
	max := p.maxRecipients()
	elements := []model.DialogElement{
		{
			DisplayName: "Add a comment (optional)",
			Name:        "comment",
			Type:        "textarea",
			Optional:    true,
		},
	}
	for i := 0; i < max; i++ {
		help := "Choose teammates to receive separate DMs (same message and attachments)."
		if i == 0 {
			help = "Required. " + help
		}
		elements = append(elements, model.DialogElement{
			DisplayName: fmt.Sprintf("Recipient %d", i+1),
			Name:        fmt.Sprintf("recipient_%d", i),
			Type:        "select",
			DataSource:  "users",
			Optional:    i > 0,
			HelpText:    help,
		})
	}
	return model.Dialog{
		CallbackId:  "forward:" + sourcePostID,
		Title:       "Forward to…",
		SubmitLabel: "Send",
		Elements:    elements,
	}
}

func (p *ForwardPrivatePlugin) handleForwardDialogSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if !p.enabled() {
		p.writeDialogError(w, "Forward Private is disabled.")
		return
	}

	var req model.SubmitDialogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		p.writeDialogError(w, "Invalid submission.")
		return
	}
	if req.Cancelled {
		w.WriteHeader(http.StatusOK)
		return
	}

	sourcePostID := strings.TrimPrefix(req.CallbackId, "forward:")
	if sourcePostID == req.CallbackId || sourcePostID == "" {
		p.writeDialogError(w, "Missing source message.")
		return
	}

	sourcePost, sourceChannel, appErr := p.canAccessSourcePost(req.UserId, sourcePostID)
	if appErr != nil {
		p.writeDialogError(w, appErr.Message)
		return
	}

	comment := stringField(req.Submission, "comment")
	var recipientIDs []string
	max := p.maxRecipients()
	for i := 0; i < max; i++ {
		id := stringField(req.Submission, fmt.Sprintf("recipient_%d", i))
		if id != "" {
			recipientIDs = append(recipientIDs, id)
		}
	}
	if len(recipientIDs) == 0 {
		p.writeDialogError(w, "Choose at least one recipient.")
		return
	}

	result := p.forwardToRecipients(req.UserId, sourcePost, sourceChannel, recipientIDs, comment)
	if len(result.succeeded) == 0 {
		p.writeDialogError(w, "Could not forward to anyone. "+strings.Join(result.failed, "; "))
		return
	}

	p.postSourceReceipt(req.UserId, sourcePost, result.succeeded, result.failed)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model.SubmitDialogResponse{})
}

func (p *ForwardPrivatePlugin) handleUserLookup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	userID, appErr := p.userFromRequest(r)
	if appErr != nil {
		http.Error(w, appErr.Error(), http.StatusUnauthorized)
		return
	}

	var req struct {
		UserInput string `json:"user_input"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	term := strings.TrimSpace(req.UserInput)

	teams, appErr := p.API.GetTeamsForUser(userID)
	if appErr != nil || len(teams) == 0 {
		p.writeJSON(w, http.StatusOK, map[string]any{"items": []any{}})
		return
	}
	teamID := teams[0].Id

	limit := 50
	users, appErr := p.API.SearchUsers(&model.UserSearch{
		Term:             term,
		TeamId:           teamID,
		NotInChannelId:   "",
		AllowInactive:    false,
		GroupConstrained: false,
		Limit:            limit,
	})
	if appErr != nil {
		p.writeJSON(w, http.StatusOK, map[string]any{"items": []any{}})
		return
	}

	type item struct {
		Text  string `json:"text"`
		Value string `json:"value"`
	}
	var items []item
	for _, u := range users {
		if u == nil || u.Id == userID || u.DeleteAt != 0 {
			continue
		}
		label := "@" + u.Username
		display := strings.TrimSpace(u.FirstName + " " + u.LastName)
		if display != "" {
			label = display + " (@" + u.Username + ")"
		}
		items = append(items, item{Text: label, Value: u.Id})
		if len(items) >= 100 {
			break
		}
	}
	p.writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
