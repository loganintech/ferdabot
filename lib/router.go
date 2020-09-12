package ferdabot

import (
	"github.com/bwmarrin/discordgo"
)

type MessageCreateRouter struct {
	routes map[string]func(s *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) FerdaAction
}

func NewMessageCreateRouter() MessageCreateRouter {
	return MessageCreateRouter{
		routes: make(map[string]func(s *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) FerdaAction),
	}
}

func (r *MessageCreateRouter) AddRoute(command string, f func(*discordgo.Session, *discordgo.MessageCreate, string) FerdaAction) FerdaAction {
	_, ok := r.routes[command]
	if ok {
		return AddRouteFailed.RenderLogText(command).Finalize()
	}
	r.routes[command] = f

	return AddRouteSuccess.RenderLogText(command).Finalize()
}

func (r *MessageCreateRouter) AddRouteWithAliases(commands []string, f func(*discordgo.Session, *discordgo.MessageCreate, string) FerdaAction) FerdaAction {
	for _, command := range commands {
		_, ok := r.routes[command]
		if ok {
			return AddRouteFailed.RenderLogText(command).Finalize()
		}
		r.routes[command] = f
	}

	return AddRouteSuccess.RenderLogText(commands[0]).Finalize()
}

func (r *MessageCreateRouter) ExecuteRoute(command string, s *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) FerdaAction {
	f, ok := r.routes[command]
	if !ok {
		return RouteNotFound.RenderLogText(command).Finalize()
	}

	return f(s, m, trimmedText)
}
