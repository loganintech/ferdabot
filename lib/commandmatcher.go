package ferdabot

import "github.com/bwmarrin/discordgo"

// MessageCreateRoute is a route for the MessageCreate event
type MessageCreateRoute struct {
	f    func(s *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) FerdaAction
	desc string
	key  string
}

// CommandMatcher routes commands to their associated processors
type CommandMatcher struct {
	root CommandNode
}

// AddCommand adds a command to the router
func (m *CommandMatcher) AddCommand(key string, command *MessageCreateRoute) FerdaAction {
	runeList := []rune(key)
	return m.root.AddCommand(runeList, command)
}

// ExecuteCommand runs a command given a key
func (m *CommandMatcher) ExecuteCommand(key string, s *discordgo.Session, msg *discordgo.MessageCreate, trimmedText string) FerdaAction {
	runeList := []rune(key)
	return m.root.ExecuteCommand(runeList, s, msg, trimmedText)
}

// CommandNode is a single node in our command trie
type CommandNode struct {
	nodes   map[rune]CommandNode
	command *MessageCreateRoute
}

// NewCommandMatcher creates a new command matcher with an empty root node
func NewCommandMatcher() CommandMatcher {
	root := CommandNode{
		nodes:   make(map[rune]CommandNode),
		command: nil,
	}
	return CommandMatcher{root: root}
}

// AddCommand adds a command into the trie
func (n *CommandNode) AddCommand(key []rune, command *MessageCreateRoute) FerdaAction {
	// If we've reached the end of our line,
	if len(key) == 1 {
		// If the node already exists
		if cmd, ok := n.nodes[key[0]]; ok {
			// But the command is nil
			if cmd.command == nil {
				// Assign the command and return success
				cmd.command = command
				return AddRouteSuccess.Finalize()
			} else {
				// If it exists, we found a conflict
				return AddRouteFailed.Finalize()
			}
		}
		// Otherwise, connect the tail node to the location it belongs in the trie
		n.nodes[key[0]] = CommandNode{
			nodes:   make(map[rune]CommandNode),
			command: command,
		}
		return AddRouteSuccess.Finalize()
	}

	// If the node already exists, recurse
	if next, ok := n.nodes[key[0]]; ok {
		return next.AddCommand(key[1:], command)
	}

	// If the node doesn't exist, create a passthrough node
	newNode := CommandNode{
		nodes:   make(map[rune]CommandNode),
		command: nil,
	}
	// Link our new node into the trie
	n.nodes[key[0]] = newNode

	// And recurse
	return newNode.AddCommand(key[1:], command)
}

func (n *CommandNode) ExecuteCommand(key []rune, s *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) FerdaAction {
	// If we're at the end of our search
	if len(key) == 0 {
		// And the command is nil
		if n.command == nil {
			// Return the route isn't found
			return RouteNotFound.Finalize()
		}
		// Otherwise, return the result of the function
		return n.command.f(s, m, trimmedText)
	}

	// Get the next node
	next, ok := n.nodes[key[0]]
	// If the node is missing, return route not found
	if !ok {
		return RouteNotFound.Finalize()
	}

	// If it's there, recurse
	return next.ExecuteCommand(key[1:], s, m, trimmedText)
}

func (m CommandMatcher) GetHelpActions() *FerdaAction {
	return m.root.GetHelpActions()
}

func (n CommandNode) GetHelpActions() *FerdaAction {
	// Start our ferda out
	var theseFerdaActions *FerdaAction = nil
	for _, node := range n.nodes {
		// If we found a command at this level
		if node.command != nil {
			// Render it into our ferda help message
			nextAction := HelpBody.RenderDiscordText(node.command.key, node.command.desc).Finalize()
			// If the current help is nil
			if theseFerdaActions == nil {
				// Assign the new help
				theseFerdaActions = &nextAction
			} else {
				// If it's not nil, append the new one
				combined := theseFerdaActions.CombineActions(nextAction)
				theseFerdaActions = &combined
			}

			// And check if it has even more descendents. Without this, we'd find `?ferda` but not `?ferdasearch`
			anotherAction := node.GetHelpActions()
			if anotherAction != nil {
				// If the current help is nil
				if theseFerdaActions == nil {
					// Assign the recursed results
					theseFerdaActions = anotherAction
				} else {
					// If the current help isn't nil, append the new one
					combined := theseFerdaActions.CombineActions(*anotherAction)
					theseFerdaActions = &combined
				}
			}
		} else {
			// If we're at a passthrough node
			nextAction := node.GetHelpActions()
			if nextAction != nil {
				// And the current ferda is empty, assign the new one
				if theseFerdaActions == nil {
					theseFerdaActions = nextAction
				} else {
					// If the current ferda isn't empty, combine them
					combined := theseFerdaActions.CombineActions(*nextAction)
					theseFerdaActions = &combined
				}
			}
		}
	}

	// Return the ferdas found
	return theseFerdaActions
}
