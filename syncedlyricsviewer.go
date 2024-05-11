package fynesyncedlyrics

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SyncedLyricsViewer struct {
	widget.BaseWidget

	Lines []string

	// one-indexed - 0 means before the first line
	currentLine int

	scroll *NoScroll
	vbox   *fyne.Container
	anim   *fyne.Animation
}

func (s *SyncedLyricsViewer) SetCurrentLine(line int) {
	s.currentLine = line + 1
}

func (s *SyncedLyricsViewer) NextLine() {
	if s.vbox == nil {
		return // renderer not created yet
	}
	if s.currentLine == len(s.Lines)-1 {
		return // can't advance
	}

	if s.anim != nil {
		s.anim.Stop()
	}

	scrollDist := s.vbox.Objects[s.currentLine].(*widget.RichText).Size().Height
	scrollDist += theme.Padding()
	origOffset := s.scroll.Offset.Y
	s.anim = fyne.NewAnimation(100*time.Millisecond, func(f float32) {
		s.scroll.Offset.Y = origOffset + f*scrollDist
		s.scroll.Refresh()
		if f == 1.0 {
			s.currentLine++
		}
	})

	s.anim.Start()
}

func (s *SyncedLyricsViewer) Refresh() {
	s.updateTextSegments()
	s.BaseWidget.Refresh()
}

func (s *SyncedLyricsViewer) Resize(size fyne.Size) {
	if s.scroll != nil {
		s.scroll.Resize(size)
	}
	s.BaseWidget.Resize(size)
}

func (s *SyncedLyricsViewer) updateTextSegments() {
	if s.vbox == nil {
		return // renderer not created yet
	}

	l := len(s.vbox.Objects)
	for i, line := range s.Lines {
		if i < l {
			rt := s.vbox.Objects[i].(*widget.RichText)
			ts := rt.Segments[0].(*widget.TextSegment)
			ts.Text = line
			rt.Refresh()
		} else {
			ts := &widget.TextSegment{
				Text:  line,
				Style: widget.RichTextStyleSubHeading,
			}
			rt := widget.NewRichText(ts)
			rt.Wrapping = fyne.TextWrapWord
			s.vbox.Objects = append(s.vbox.Objects, rt)
		}
	}
	for i := len(s.Lines); i < l; i++ {
		s.vbox.Objects[i] = nil
	}
	s.vbox.Objects = s.vbox.Objects[:len(s.Lines)]
	s.vbox.Refresh()
}

func (s *SyncedLyricsViewer) CreateRenderer() fyne.WidgetRenderer {
	s.vbox = container.NewVBox()
	s.scroll = NewNoScroll(s.vbox)
	s.scroll.Direction = container.ScrollNone
	s.updateTextSegments()
	return widget.NewSimpleRenderer(s.scroll)
}

// overridden container.Scroll to not respond to mouse wheel/trackpad
type NoScroll struct {
	container.Scroll
}

func NewNoScroll(content fyne.CanvasObject) *NoScroll {
	n := &NoScroll{
		Scroll: container.Scroll{
			Content: content,
		},
	}
	n.ExtendBaseWidget(n)
	return n
}

func (n *NoScroll) Scrolled(_ *fyne.ScrollEvent) {
	// ignore scroll event
}
