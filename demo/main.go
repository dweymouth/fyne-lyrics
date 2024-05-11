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
		},
	}

	/*
				scroll := container.NewVScroll(widget.NewRichTextFromMarkdown(`# Heading
			* Bullet point 1
			* Bullet point 2
			* Bullet point 3

			## Subheading

			* More content 1
			* More content 2`))


		win.SetContent(container.NewBorder(widget.NewButton("Scroll", func() {
			scroll.Offset = fyne.NewPos(0, 20)
			scroll.Refresh()
		}), nil, nil, nil, scroll))

	*/
	win.SetContent(l)
	win.Resize(fyne.NewSize(200, 300))

	go func() {
		time.Sleep(1 * time.Second)
		l.NextLine()
	}()

	win.ShowAndRun()
}
