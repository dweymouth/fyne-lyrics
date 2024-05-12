package fynesyncedlyrics

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SyncedLyricsViewer struct {
	widget.BaseWidget
	mutex sync.Mutex

	Lines []string

	// one-indexed - 0 means before the first line
	currentLine int

	singleLineLyricHeight float32

	scroll *NoScroll
	vbox   *fyne.Container
	anim   *fyne.Animation
}

func NewSyncedLyricsViewer() *SyncedLyricsViewer {
	s := &SyncedLyricsViewer{}
	s.ExtendBaseWidget(s)
	return s
}

func (s *SyncedLyricsViewer) SetCurrentLine(line int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.currentLine = line + 1
}

func (s *SyncedLyricsViewer) NextLine() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.vbox == nil /*no renderer yet*/ || s.currentLine == len(s.Lines)-1 {
		return
	}
	s.currentLine++
	s.checkStopAnimation()

	var prevLine, nextLine *widget.RichText
	if s.currentLine > 1 {
		prevLine = s.vbox.Objects[s.currentLine-1].(*widget.RichText)
	}
	if s.currentLine < len(s.Lines) {
		nextLine = s.vbox.Objects[s.currentLine].(*widget.RichText)
	}

	scrollDist := nextLine.Size().Height
	scrollDist += theme.Padding()
	origOffset := s.scroll.Offset.Y
	var alreadyUpdated bool
	s.anim = fyne.NewAnimation(100*time.Millisecond, func(f float32) {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		s.scroll.Offset.Y = origOffset + f*scrollDist
		s.scroll.Refresh()
		if !alreadyUpdated && f >= 0.5 {
			if nextLine != nil {
				s.setLineColor(nextLine, theme.ColorNameForeground)
			}
			if prevLine != nil {
				s.setLineColor(prevLine, theme.ColorNameDisabled)
			}
			alreadyUpdated = true
		}
	})
	s.anim.Curve = fyne.AnimationEaseInOut

	s.anim.Start()
}

func (s *SyncedLyricsViewer) Refresh() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.updateContent()
	s.BaseWidget.Refresh()
}

func (s *SyncedLyricsViewer) Resize(size fyne.Size) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.scroll != nil {
		s.scroll.Resize(size)
	}
	s.BaseWidget.Resize(size)
}

func (s *SyncedLyricsViewer) updateContent() {
	if s.vbox == nil {
		return // renderer not created yet
	}

	l := len(s.vbox.Objects)
	//endSpacer := s.vbox.Objects[l-1]
	for i, line := range s.Lines {
		if i < l {
			rt := s.vbox.Objects[i].(*widget.RichText)
			ts := rt.Segments[0].(*widget.TextSegment)
			ts.Text = line
			rt.Refresh()
		} else {
			s.vbox.Objects = append(s.vbox.Objects, s.newLyricLine(line))
		}
	}
	for i := len(s.Lines); i < l; i++ {
		s.vbox.Objects[i] = nil
	}
	s.vbox.Objects = s.vbox.Objects[:len(s.Lines)]
	s.vbox.Refresh()
}

func (s *SyncedLyricsViewer) newLyricLine(text string) *widget.RichText {
	ts := &widget.TextSegment{
		Text:  text,
		Style: widget.RichTextStyleSubHeading,
	}
	ts.Style.ColorName = theme.ColorNameDisabled
	rt := widget.NewRichText(ts)
	rt.Wrapping = fyne.TextWrapWord
	return rt
}

func (s *SyncedLyricsViewer) setLineColor(rt *widget.RichText, colorName fyne.ThemeColorName) {
	rt.Segments[0].(*widget.TextSegment).Style.ColorName = colorName
	rt.Refresh()
}

func (s *SyncedLyricsViewer) checkStopAnimation() {
	if s.anim != nil {
		s.anim.Stop()
		s.anim = nil
	}
}

func (s *SyncedLyricsViewer) CreateRenderer() fyne.WidgetRenderer {
	s.singleLineLyricHeight = s.newLyricLine("W").MinSize().Height
	s.vbox = container.NewVBox()
	s.scroll = NewNoScroll(s.vbox)
	s.updateContent()
	return widget.NewSimpleRenderer(s.scroll)
}

// overridden container.Scroll to not respond to mouse wheel/trackpad
type NoScroll struct {
	container.Scroll
}

func NewNoScroll(content fyne.CanvasObject) *NoScroll {
	n := &NoScroll{
		Scroll: container.Scroll{
			Content:   content,
			Direction: container.ScrollNone,
		},
	}
	n.ExtendBaseWidget(n)
	return n
}

func (n *NoScroll) Scrolled(_ *fyne.ScrollEvent) {
	// ignore scroll event
}

type vSpace struct {
	layout.Spacer

	Height float32
}

func NewVSpace(height float32) *vSpace {
	return &vSpace{Height: height}
}

func (v *vSpace) MinSize() fyne.Size {
	return fyne.NewSize(0, v.Height)
}
