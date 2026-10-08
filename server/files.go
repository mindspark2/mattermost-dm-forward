package main

import (
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
)

func (p *ForwardPrivatePlugin) copyPostFilesToChannel(sourcePost *model.Post, targetChannelID string) ([]string, error) {
	if sourcePost == nil || len(sourcePost.FileIds) == 0 {
		return nil, nil
	}
	var out []string
	for _, fid := range sourcePost.FileIds {
		info, appErr := p.API.GetFileInfo(fid)
		if appErr != nil || info == nil {
			return out, fmt.Errorf("file info: %v", appErr)
		}
		data, appErr := p.API.GetFile(fid)
		if appErr != nil {
			return out, fmt.Errorf("read file: %v", appErr)
		}
		name := info.Name
		if name == "" {
			name = "attachment"
		}
		uploaded, appErr := p.API.UploadFile(data, targetChannelID, name)
		if appErr != nil {
			return out, fmt.Errorf("upload %s: %v", name, appErr)
		}
		if uploaded == nil {
			return out, fmt.Errorf("upload returned nil for %s", name)
		}
		out = append(out, uploaded.Id)
	}
	return out, nil
}
