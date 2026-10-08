package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
)

func stringField(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strings.TrimSpace(fmt.Sprintf("%.0f", t))
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", t))
	}
}

func contextString(ctx map[string]interface{}, key string) string {
	if ctx == nil {
		return ""
	}
	return stringField(ctx, key)
}

func (p *ForwardPrivatePlugin) userFromRequest(r *http.Request) (string, *model.AppError) {
	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		return "", model.NewAppError("ForwardPrivate", "api.context.session_expired.app_error", nil, "missing Mattermost-User-Id", http.StatusUnauthorized)
	}
	return userID, nil
}

func (p *ForwardPrivatePlugin) writeDialogError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(model.SubmitDialogResponse{Error: msg})
}

func (p *ForwardPrivatePlugin) isPrivateConversationChannel(channel *model.Channel) bool {
	if channel == nil {
		return false
	}
	return channel.Type == model.ChannelTypeDirect || channel.Type == model.ChannelTypeGroup
}

func (p *ForwardPrivatePlugin) canAccessSourcePost(userID, postID string) (*model.Post, *model.Channel, *model.AppError) {
	post, appErr := p.API.GetPost(postID)
	if appErr != nil || post == nil {
		return nil, nil, model.NewAppError("ForwardPrivate", "app.post.get.app_error", nil, "post not found", http.StatusNotFound)
	}
	if post.DeleteAt != 0 {
		return nil, nil, model.NewAppError("ForwardPrivate", "app.post.get.app_error", nil, "post deleted", http.StatusNotFound)
	}
	ch, appErr := p.API.GetChannel(post.ChannelId)
	if appErr != nil || ch == nil {
		return nil, nil, appErr
	}
	if !p.isPrivateConversationChannel(ch) {
		return nil, nil, model.NewAppError("ForwardPrivate", "plugin.forward_private.channel_type", nil, "only DMs and group messages", http.StatusForbidden)
	}
	if _, appErr := p.API.GetChannelMember(ch.Id, userID); appErr != nil {
		return nil, nil, model.NewAppError("ForwardPrivate", "app.channel.get_member.missing.app_error", nil, "", http.StatusForbidden)
	}
	return post, ch, nil
}

func userAt(u *model.User) string {
	if u == nil {
		return "@unknown"
	}
	return "@" + u.Username
}

func joinUserMentions(users []*model.User) string {
	if len(users) == 0 {
		return ""
	}
	parts := make([]string, 0, len(users))
	for _, u := range users {
		parts = append(parts, userAt(u))
	}
	if len(parts) == 1 {
		return parts[0]
	}
	if len(parts) == 2 {
		return parts[0] + " and " + parts[1]
	}
	return strings.Join(parts[:len(parts)-1], ", ") + ", and " + parts[len(parts)-1]
}
