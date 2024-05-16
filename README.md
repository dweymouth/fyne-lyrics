# fyne-lyrics
A lyrics widget for Fyne supporting synced and unsynced mode.

## Usage (synced lyrics)

```go
l := fynelyics.NewLyricsViewer()
l.SetLyrics([]string{"My song lyrics", "And some more", "And another"}, true /*synced*/)

myMediaPlayer.OnPositionUpdate(func(...) {
    // When updating playback position in your gui, check if the playback
    // time has advanced to the next lyric's start time, if so...
    // No-op if the widget is in unsynced mode
    l.NextLine() // animates the scroll
})

myMediaPlayer.OnSeeked(func(...) {
    // When user seeks the song, check the new position against the lyric line
    // start times, and if needed...
    // No-op if the widget is in unsynced mode
    l.SetCurrentLine(n /*line number to display*/) // scroll not animated
})

myMediaPlayer.OnNextSong(func(...) {
    // Reset the widget with the next song's lyrics
    // This automatically resets the scroll position to before the first line
    l.SetLyrics(lyrics, true /*synced*/)
})

```
