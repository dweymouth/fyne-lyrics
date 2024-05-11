package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	fynesyncedlyrics "github.com/dweymouth/fyne-synced-lyrics"
)

func main() {
	app := app.New()
	win := app.NewWindow("Synced Lyrics Demo")

	l := &fynesyncedlyrics.SyncedLyricsViewer{
		Lines: []string{
			"Hello world",
			"These are my song lyrics",
			"Let's add some more",
			"And even some more",
			"And now yet another",
			"And how about even one more to see",
			"And just so we have a long enough song...",
			"Here we go to repeat.",
			"Hello world",
			"These are my song lyrics",
			"Let's add some more",
			"And even some more",
			"And now yet another",
			"And how about even one more to see",
			"And just so we have a long enough song...",
			"Here we go to repeat",
			"Just kidding we're done.",
		},
	}

	win.SetContent(l)
	win.Resize(fyne.NewSize(200, 300))

	tick := time.NewTicker(1 * time.Second)
	go func() {
		for {
			<-tick.C
			l.NextLine()
		}
	}()

	win.ShowAndRun()
}
