package tars

import "embed"

//go:embed migrations/*.sql
var MigrationsFS embed.FS

//go:embed web/*
var WebFS embed.FS
