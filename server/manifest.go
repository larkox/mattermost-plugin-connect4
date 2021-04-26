// This file is automatically generated. Do not modify it manually.

package main

import (
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
)

var manifest *model.Manifest

const manifestStr = `
{
  "id": "com.mattermost.plugin-starter-template",
  "name": "Plugin Starter Template",
  "description": "This plugin serves as a starting point for writing a Mattermost plugin.",
  "homepage_url": "https://github.com/larkox/mattermost-plugin-connect4",
  "support_url": "https://github.com/larkox/mattermost-plugin-connect4/issues",
  "release_notes_url": "https://github.com/larkox/mattermost-plugin-connect4/releases/tag/v0.1.0",
  "icon_path": "assets/starter-template-icon.svg",
  "version": "0.1.0",
  "min_server_version": "5.12.0",
  "server": {
    "executables": {
      "linux-amd64": "server/dist/plugin-linux-amd64",
      "darwin-amd64": "server/dist/plugin-darwin-amd64",
      "windows-amd64": "server/dist/plugin-windows-amd64.exe"
    },
    "executable": ""
  },
  "webapp": {
    "bundle_path": "webapp/dist/main.js"
  },
  "settings_schema": {
    "header": "",
    "footer": "",
    "settings": []
  }
}
`

func init() {
	manifest = model.ManifestFromJson(strings.NewReader(manifestStr))
}
