//go:build ignore
// +build ignore

package main

import (
	"os"

	"github.com/thehowl/conf"
	"golang.org/x/exp/slog"
)

type simplePage struct {
	Handler, Template, TitleBar, BannerContent string
	BannerType                                 int
	MinPrivilegesRaw                           uint64
}

type noTemplate struct {
	Handler, TitleBar, BannerContent string
	BannerType                       int
	MinPrivileges                    uint64
}

var simplePages = [...]simplePage{
	{"/", "homepage.html", "Home Page", "homepage2.jpg", 1, 0},
	{"/login", "login.html", "Log in", "login2.jpg", 1, 0},
	{"/settings/avatar", "settings/avatar.html", "Change avatar", "settings2.jpg", 1, 2},
	{"/dev/tokens", "dev/tokens.html", "Your API tokens", "dev.jpg", 1, 2},
	{"/donate", "support.html", "Support Ripple", "donate2.png", 1, 0},
	{"/doc", "doc.html", "Documentation", "documentation.jpg", 1, 0},
	{"/doc/:id", "doc_content.html", "View document", "documentation.jpg", 1, 0},
	{"/help", "help.html", "Contact support", "help.jpg", 1, 0},
	{"/leaderboard", "leaderboard.html", "Leaderboard", "leaderboard2.jpg", 1, 0},
	{"/friends", "friends.html", "Friends", "", 0, 2},
	{"/changelog", "changelog.html", "Changelog", "changelog.jpg", 1, 0},
	{"/team", "team.html", "Team", "", 0, 0},
	{"/pwreset", "pwreset.html", "Reset password", "", 0, 0},
	{"/about", "about.html", "About", "", 0, 0},
	{"/patcher", "patcherdl.html", "Akatsuki Patcher", "documentation.jpg", 1, 0},
	// TODO: should merge.html be here?
}

func main() {
	for _, p := range simplePages {
		slog.Info("Generating config for", "handler", p.Handler)
		noTemplateP := noTemplate{
			Handler:       p.Handler,
			TitleBar:      p.TitleBar,
			BannerContent: p.BannerContent,
			BannerType:    p.BannerType,
			MinPrivileges: p.MinPrivilegesRaw,
		}
		d := []byte("{{/*###\n")
		confData, err := conf.ExportRaw(&noTemplateP)
		if err != nil {
			slog.Error("Error exporting config", "error", err.Error())
			panic(err)
		}
		d = append(d, confData...)
		fileData, err := os.ReadFile("templates/" + p.Template)
		if err != nil {
			slog.Error("Error reading file", "error", err.Error())
			panic(err)
		}
		d = append(d, []byte("*/}}\n")...)
		d = append(d, fileData...)
		err = os.WriteFile("templates/"+p.Template, d, 0644)
		if err != nil {
			slog.Error("Error writing file", "error", err.Error())
			panic(err)
		}
		slog.Info("ok.")
	}
}
