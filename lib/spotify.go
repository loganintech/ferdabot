package ferdabot

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/zmb3/spotify"
)

var whitelistedMusicPlaylists = []string{
	"Bot Testin",
	"Cousin Music",
	"Music",
}

var getLimit = 50

var spotifyPlaylistRegex = regexp.MustCompile("track[:/]([0-9a-zA-Z]{22})")

// processDice processes the dice roll command
func (b *Bot) processSpotifyLink(s *discordgo.Session, m *discordgo.MessageCreate) *FerdaAction {
	fromChannel, channelErr := s.Channel(m.ChannelID)
	if channelErr != nil {
		return &FromChannelNotFound
	}

	isOk := false
	for _, okPlaylist := range whitelistedMusicPlaylists {
		if okPlaylist == channelNameToPlaylistName(fromChannel.Name) {
			isOk = true
		}
	}
	if !isOk {
		return nil
	}

	text := m.Content
	matches := spotifyPlaylistRegex.FindAllStringSubmatch(text, -1)
	var songIDs []spotify.ID
	for _, submatch := range matches {
		if len(submatch) > 1 {
			match := submatch[1]
			songIDs = append(songIDs, spotify.ID(match))
			b.ProcessFerdaAction(SpotifySongAdded.RenderLogText(match, "playlist").Finalize(), s, m)
		}
	}

	currentUser, userErr := b.spotifyConnection.CurrentUser()
	if userErr != nil {
		userAction := SpotifyUserNotFound.RenderLogText(userErr).Finalize()
		return &userAction
	}

	listSearch, playlistErr := b.spotifyConnection.GetPlaylistsForUserOpt(currentUser.ID, &spotify.Options{
		Limit: &getLimit,
	})
	if playlistErr != nil {
		return &SpotifyPlaylistNotFound
	}

	var playlistID spotify.ID
	for _, list := range listSearch.Playlists {
		if list.Name == channelNameToPlaylistName(fromChannel.Name) {
			playlistID = list.ID
		}
	}

	if playlistID.String() == "" {
		playlistName := channelNameToPlaylistName(fromChannel.Name)
		newPlaylist, playlistErr := b.spotifyConnection.CreatePlaylistForUser(currentUser.ID, playlistName, fmt.Sprintf("%s created by Ferdabot", playlistName), true)
		if playlistErr != nil {
			playlistAction := SpotifyCreatePlaylistError.RenderLogText(playlistErr).Finalize()
			return &playlistAction
		}
		playlistID = newPlaylist.ID
	}

	_, addErr := b.spotifyConnection.AddTracksToPlaylist(playlistID, songIDs...)
	if addErr != nil {
		playlistAddErr := SpotifyAddToPlaylistError.RenderLogText(addErr).Finalize()
		return &playlistAddErr
	}

	return &DontLog
}

func channelNameToPlaylistName(channelName string) string {
	parts := strings.Split(channelName, "-")
	for i := 0; i < len(parts); i++ {
		parts[i] = strings.ToUpper(string(parts[i][0])) + parts[i][1:]
	}
	return strings.Join(parts, " ")
}
