package ferdabot

import "github.com/bwmarrin/discordgo"

type MessageCreateRoute struct {
	f    func(s *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) FerdaAction
	desc string
	key  string
}

type CommandMatcher struct {
	root CommandNode
}

func (m *CommandMatcher) AddCommand(key string, command *MessageCreateRoute) FerdaAction {
	runeList := []rune(key)
	return m.root.AddCommand(runeList, command)
}

func (m *CommandMatcher) ExecuteCommand(key string, s *discordgo.Session, msg *discordgo.MessageCreate, trimmedText string) FerdaAction {
	runeList := []rune(key)
	return m.root.ExecuteCommand(runeList, s, msg, trimmedText)
}

type CommandNode struct {
	nodes   map[rune]CommandNode
	command *MessageCreateRoute
}

func NewCommandMatcher() CommandMatcher {
	root := CommandNode{
		nodes:   make(map[rune]CommandNode),
		command: nil,
	}
	return CommandMatcher{root: root}
}

func (n *CommandNode) AddCommand(key []rune, command *MessageCreateRoute) FerdaAction {
	// If we've reached the end of our line,
	if len(key) == 1 {
		if n.command != nil {
			return AddRouteFailed.Finalize()
		}
		n.nodes[key[0]] = CommandNode{
			nodes:   make(map[rune]CommandNode),
			command: command,
		}
		return AddRouteSuccess.Finalize()
	}

	if next, ok := n.nodes[key[0]]; ok {
		return next.AddCommand(key[1:], command)
	}

	newNode := CommandNode{
		nodes:   make(map[rune]CommandNode),
		command: nil,
	}
	n.nodes[key[0]] = newNode

	return newNode.AddCommand(key[1:], command)

}

func (n *CommandNode) ExecuteCommand(key []rune, s *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) FerdaAction {
	if len(key) == 0 {
		if n.command == nil {
			return RouteNotFound.Finalize()
		}
		return n.command.f(s, m, trimmedText)
	}

	next, ok := n.nodes[key[0]]
	if !ok {
		return RouteNotFound.Finalize()
	}

	return next.ExecuteCommand(key[1:], s, m, trimmedText)
}
func (m CommandMatcher) GetHelpActions() *FerdaAction {
	return m.root.GetHelpActions()
}

func (n CommandNode) GetHelpActions() *FerdaAction {
	var theseFerdaActions *FerdaAction = nil
	for _, node := range n.nodes {
		// If we found a command at this level
		if node.command != nil {
			// Render it into our ferda help message
			nextAction := HelpBody.RenderDiscordText(node.command.key, node.command.desc).Finalize()
			if theseFerdaActions == nil {
				theseFerdaActions = &nextAction
			} else {
				combined := theseFerdaActions.CombineActions(nextAction)
				theseFerdaActions = &combined
			}

			// And check if it has even more descendents. Without this, we'd find `?ferda` but not `?ferdasearch`
			anotherAction := node.GetHelpActions()
			if anotherAction != nil {
				if theseFerdaActions == nil {
					theseFerdaActions = anotherAction
				} else {
					combined := theseFerdaActions.CombineActions(*anotherAction)
					theseFerdaActions = &combined
				}
			}
		} else {
			nextAction := node.GetHelpActions()
			if nextAction != nil {
				if theseFerdaActions == nil {
					theseFerdaActions = nextAction
				} else {
					combined := theseFerdaActions.CombineActions(*nextAction)
					theseFerdaActions = &combined
				}
			}
		}
	}

	return theseFerdaActions
}
