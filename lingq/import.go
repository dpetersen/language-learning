package lingq

import (
	"fmt"
)

/*
POST https://www.lingq.com/api/v3/es/lessons/import/

Response: 201 with no body

Params in Chrome:
audio: (binary)
image: (binary)
collection: 1467506
description: Una historia simple sobre una niña llamada Sofía que pasa un día en el parque y hace un nuevo amigo.
file: (binary)
hasPrice: false
isProtected: false
isHidden: true
language: es
status: private
tags: gpt
title: El día en el parque de Sofía
translations: []
notes:
save: true
*/

const importURL = "https://www.lingq.com/api/v3/es/lessons/import/"

func (c *Client) ImportLesson(textPath, audioPath, thumbnailPath, description, title string) error {
	response, err := c.newAPIRequest().
		SetFile("image", thumbnailPath).
		SetFile("audio", audioPath).
		SetFile("file", textPath).
		SetFormData(map[string]string{
			"description": description,
			"title":       title,
			// TODO: don't hardcode this?
			"collection":  "1467506",
			"hasPrice":    "false",
			"isProtected": "false",
			"isHidden":    "true",
			"language":    "es",
			"status":      "private",
			"tags":        "gpt",
			"save":        "true",
		}).
		Post(importURL)
	if err != nil {
		return fmt.Errorf("making API request: %v", err)
	}

	if response.StatusCode() != 201 {
		return fmt.Errorf("got unexpected status code: %d", response.StatusCode())
	}

	return nil
}
