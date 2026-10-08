package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
)

func (p *ForwardPrivatePlugin) buildForwardBody(sourcePost *model.Post, sourceChannel *model.Channel, comment string) string {
	var b strings.Builder
	if c := strings.TrimSpace(comment); c != "" {
		b.WriteString(c)
		b.WriteString("\n\n")
	}
	if sourcePost != nil && strings.TrimSpace(sourcePost.Message) != "" {
		b.WriteString(sourcePost.Message)
	}
	if p.getConfiguration().ShowAttribution {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		label := originLabel(sourceChannel)
		b.WriteString(fmt.Sprintf("_Forwarded from a private conversation (%s)._", label))
	}
	return strings.TrimSpace(b.String())
}

func originLabel(ch *model.Channel) string {
	if ch == nil {
		return "private chat"
	}
	if ch.DisplayName != "" {
		return ch.DisplayName
	}
	return "private chat"
}

type forwardResult struct {
	succeeded []*model.User
	failed    []string
}

func (p *ForwardPrivatePlugin) forwardToRecipients(actorID string, sourcePost *model.Post, sourceChannel *model.Channel, recipientIDs []string, comment string) forwardResult {
	body := p.buildForwardBody(sourcePost, sourceChannel, comment)
	var res forwardResult
	seen := map[string]bool{actorID: true}

	for _, rid := range recipientIDs {
		rid = strings.TrimSpace(rid)
		if rid == "" || seen[rid] {
			continue
		}
		seen[rid] = true

		u, appErr := p.API.GetUser(rid)
		if appErr != nil || u == nil || u.DeleteAt != 0 {
			res.failed = append(res.failed, fmt.Sprintf("unknown user %s", rid))
			continue
		}

		dm, appErr := p.API.GetDirectChannel(actorID, rid)
		if appErr != nil {
			res.failed = append(res.failed, userAt(u)+": "+appErr.Error())
			continue
		}
		if dm == nil {
			res.failed = append(res.failed, userAt(u)+": could not open DM")
			continue
		}

		fileIDs, err := p.copyPostFilesToChannel(sourcePost, dm.Id)
		if err != nil {
			res.failed = append(res.failed, userAt(u)+": "+err.Error())
			continue
		}

		newPost := &model.Post{
			UserId:    actorID,
			ChannelId: dm.Id,
			Message:   body,
			FileIds:   fileIDs,
		}
		if _, appErr := p.API.CreatePost(newPost); appErr != nil {
			res.failed = append(res.failed, userAt(u)+": "+appErr.Error())
			continue
		}
		res.succeeded = append(res.succeeded, u)
	}
	return res
}

func (p *ForwardPrivatePlugin) postSourceReceipt(actorID string, sourcePost *model.Post, succeeded []*model.User, failed []string) {
	if !p.getConfiguration().PostReceiptInSourceChannel || sourcePost == nil || len(succeeded) == 0 {
		return
	}
	actor, _ := p.API.GetUser(actorID)
	actorName := userAt(actor)
	dest := joinUserMentions(succeeded)
	msg := fmt.Sprintf("%s forwarded this message to %s.", actorName, dest)
	if len(failed) > 0 {
		msg += fmt.Sprintf(" _(Some deliveries failed: %s)_", strings.Join(failed, "; "))
	}

	receipt := &model.Post{
		UserId:    actorID,
		ChannelId: sourcePost.ChannelId,
		RootId:    sourcePost.Id,
		Message:   msg,
	}
	if _, appErr := p.API.CreatePost(receipt); appErr != nil {
		p.API.LogError("forward-private: receipt post failed", "err", appErr.Error())
	}
}
