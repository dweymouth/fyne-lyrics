# fyne-lyrics
A lyrics widget for Fyne supporting synced and unsynced mode.

## Usage (synced lyrics)

```go
l := fynelyics.NewLyricsViewer()
l.SetLyrics([]string{"My song lyrics", "And some more", "And another"}, true /*synced*/)

myMediaPlayer.OnPositionUpdate(func(...) {
    // When updating playback position in your gui, check if the playback
    // time has advanced to the next lyric's start time, if so...
    l.NextLine() // animates the scroll. No-op in unsynced mode.
})

myMediaPlayer.OnSeeked(func(...) {
    // When user seeks the song, check the new position against the lyric line
    // start times, and if needed...
    l.SetCurrentLine(n /*line number to display*/) // scroll not animated. No-op in unsynced mode.
})

myMediaPlayer.OnNextSong(func(...) {
    // Reset the widget with the next song's lyrics
    // This automatically resets the scroll position to before the first line
    l.SetLyrics(lyrics, true /*synced*/)
})

```
