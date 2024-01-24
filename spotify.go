package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/zmb3/spotify"
)

const (
	redirectURI  = "http://localhost:8080/callback"
	clientID     = "598654e5de9140c4897b8ad235595cc5"
	clientSecret = "8f16007838a742dc878ddf6678c017d8"
	playlistID   = "1G9sc90TdXmgtP3aQfR7UD"
)

var (
	auth   = spotify.NewAuthenticator(redirectURI, spotify.ScopePlaylistModifyPublic, spotify.ScopePlaylistModifyPrivate, spotify.ScopeUserReadPrivate)
	state  = "abc123" // some random string to protect against CSRF attacks
	ch     = make(chan *spotify.Client)
	client *spotify.Client
	srv    http.Server
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	url := auth.AuthURL(state)
	fmt.Fprintf(w, "Please log in to Spotify by visiting the following page in your browser: %s\n", url)
}

func init() {
	auth.SetAuthInfo(clientID, clientSecret)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	// the code will be in the "code" query parameter
	token, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}

	// create a client using the specified token
	newClient := auth.NewClient(token)
	client = &newClient

	// use client to make calls that require authorization
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)

	// Trigger server shutdown
	// shutdown <- true
}

func importHandler(w http.ResponseWriter, r *http.Request) {
	songs := processRecords()

	var trackIDs []spotify.ID

	for _, item := range songs {
		song := item[0]
		artist := item[1]
		fmt.Println("Processing song ", song, " - ", artist)
		trackID, err := SearchTrack(client, song, artist)
		if err != nil {
			fmt.Println("Error processing song ", song, " - ", artist)
			continue
		}

		trackIDs = append(trackIDs, trackID)
	}

	fmt.Println(trackIDs)

	// Adding tracks to the playlist
	chunks := chunkSlice(removeDuplicates(trackIDs), 100)

	for _, chunk := range chunks {
		snapshotID, err := client.AddTracksToPlaylist(playlistID, chunk...)
		if err != nil {
			log.Fatalf("Couldn't add tracks to playlist: %v", err)
		}

		fmt.Println("Added tracks to playlist. Snapshot ID:", snapshotID)
	}

}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	err := ClearPlaylist(client, playlistID)
	if err != nil {
		fmt.Println("Error clearing playlist", err)
	}
}

// SearchTrack searches for a track by its name and artist and returns the Spotify URI
func SearchTrack(client *spotify.Client, trackName, artistName string) (spotify.ID, error) {
	// Define the search query combining track name and artist
	searchQuery := trackName + " artist:" + artistName

	// Perform the search
	results, err := client.Search(searchQuery, spotify.SearchTypeTrack)
	if err != nil {
		return "", err // Return an empty ID and the error
	}

	// Check if there are any tracks in the search results
	if results.Tracks != nil && len(results.Tracks.Tracks) > 0 {
		// Return the ID of the first track found
		track := results.Tracks.Tracks[0]
		return track.ID, nil
	}

	return "", fmt.Errorf("no track found", searchQuery, client)
}

func ClearPlaylist(client *spotify.Client, playlistID spotify.ID) error {
	// Step 1: Get all track IDs from the playlist
	trackIDs, err := GetAllTrackIDsFromPlaylist(client, playlistID)
	if err != nil {
		return err // return any errors encountered
	}

	// Step 2: Remove all tracks from the playlist
	// Spotify's API requires us to provide the tracks as a slice of spotify.TrackToRemove
	var tracksToRemove []spotify.ID
	for _, id := range trackIDs {
		tracksToRemove = append(tracksToRemove, id)
	}

	// Remove the tracks in chunks if necessary (Spotify may have limits per request)
	chunks := chunkSlice(trackIDs, 100)

	for _, chunk := range chunks {
		_, err := client.RemoveTracksFromPlaylist(playlistID, chunk...)
		if err != nil {
			return err // return any errors encountered
		}
	}
	return nil
}

func GetAllTrackIDsFromPlaylist(client *spotify.Client, playlistID spotify.ID) ([]spotify.ID, error) {
	var trackIDs []spotify.ID

	// Spotify pagination starts at offset 0
	offset := 0
	for {
		// Get a page of playlist tracks
		tracks, err := client.GetPlaylistTracksOpt(playlistID, &spotify.Options{Offset: &offset}, "")
		if err != nil {
			return nil, err
		}

		// Add track IDs to the slice
		for _, item := range tracks.Tracks {
			trackIDs = append(trackIDs, item.Track.ID)
		}

		// Break the loop if we've retrieved all tracks
		if len(tracks.Tracks) < 100 {
			break
		}

		// Move to the next page
		offset += len(tracks.Tracks)
	}

	return trackIDs, nil
}

func chunkSlice(slice []spotify.ID, chunkSize int) [][]spotify.ID {
	var chunks [][]spotify.ID
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

func removeDuplicates(elements []spotify.ID) []spotify.ID {
	seen := make(map[spotify.ID]bool)
	unique := []spotify.ID{}

	for _, value := range elements {
		if _, ok := seen[value]; !ok {
			seen[value] = true
			unique = append(unique, value)
		}
	}
	return unique
}
