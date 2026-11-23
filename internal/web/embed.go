package web

import "embed"

// ContentFS embeds the static web assets and HTML templates for the wbweb application.
//
//go:embed static/* templates/*
var ContentFS embed.FS
