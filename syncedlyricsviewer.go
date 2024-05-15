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

	scroll          *NoScroll
	vbox            *fyne.Container
	anim            *fyne.Animation
	animStartOffset float32
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
	if s.vbox == nil {
		return // no renderer yet
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.currentLine == len(s.Lines) {
		return // already at last line
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

	s.setupScrollAnimation(prevLine, nextLine)
	s.anim.Start()
}

func (s *SyncedLyricsViewer) Refresh() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.updateContent()
	s.BaseWidget.Refresh()
}

func (s *SyncedLyricsViewer) MinSize() fyne.Size {
	// overridden because NoScroll will have minSize encompass the full lyrics
	minHeight := s.singleLineLyricHeight*3 + theme.Padding()*2
	return fyne.NewSize(s.BaseWidget.MinSize().Width, minHeight)
}

func (s *SyncedLyricsViewer) Resize(size fyne.Size) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.updateSpacerSize(size)
	s.BaseWidget.Resize(size)
	if s.anim == nil {
		s.scroll.Offset = fyne.NewPos(0, s.offsetForLine(s.currentLine))
		s.scroll.Refresh()
	} else {
		// animation is running - update its reference scroll pos
		s.animStartOffset = s.offsetForLine(s.currentLine - 1)
	}
}

func (s *SyncedLyricsViewer) updateSpacerSize(size fyne.Size) {
	if s.vbox == nil {
		return // renderer not created yet
	}

	topSpaceHeight := theme.Padding() + (size.Height-s.singleLineLyricHeight)/2
	s.vbox.Objects[0].(*vSpace).Height = topSpaceHeight
}

func (s *SyncedLyricsViewer) updateContent() {
	if s.vbox == nil {
		return // renderer not created yet
	}

	l := len(s.vbox.Objects)
	if l == 0 {
		s.vbox.Objects = append(s.vbox.Objects, NewVSpace(0))
		l = 1
	}
	s.updateSpacerSize(s.Size())
	//endSpacer := s.vbox.Objects[l-1]
	for i, line := range s.Lines {
		if (i + 1) < l {
			s.setLineText(s.vbox.Objects[i+1].(*widget.RichText), line)
		} else {
			s.vbox.Objects = append(s.vbox.Objects, s.newLyricLine(line))
		}
	}
	for i := len(s.Lines) + 1; i < l; i++ {
		s.vbox.Objects[i] = nil
	}
	s.vbox.Objects = s.vbox.Objects[:len(s.Lines)+1]
	s.vbox.Refresh()
}

func (s *SyncedLyricsViewer) setupScrollAnimation(currentLine, nextLine *widget.RichText) {
	// calculate total scroll distance for the animation
	scrollDist := theme.Padding()
	if currentLine != nil {
		scrollDist += currentLine.Size().Height / 2
	} else {
		scrollDist += s.singleLineLyricHeight / 2
	}
	if nextLine != nil {
		scrollDist += nextLine.Size().Height / 2
	} else {
		scrollDist += s.singleLineLyricHeight / 2
	}

	s.animStartOffset = s.scroll.Offset.Y
	var alreadyUpdated bool
	s.anim = fyne.NewAnimation(100*time.Millisecond, func(f float32) {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		s.scroll.Offset.Y = s.animStartOffset + f*scrollDist
		s.scroll.Refresh()
		if !alreadyUpdated && f >= 0.5 {
			if nextLine != nil {
				s.setLineColor(nextLine, theme.ColorNameForeground)
			}
			if currentLine != nil {
				s.setLineColor(currentLine, theme.ColorNameDisabled)
			}
			alreadyUpdated = true
		}
		if f == 1 /*end of animation*/ {
			s.anim = nil
		}
	})
	s.anim.Curve = fyne.AnimationEaseInOut
}

func (s *SyncedLyricsViewer) offsetForLine(lineNum int /*one-indexed*/) float32 {
	if lineNum == 0 {
		return 0
	}
	pad := theme.Padding()
	offset := pad + s.singleLineLyricHeight/2
	for i := 1; i <= lineNum; i++ {
		if i > 1 {
			offset += s.vbox.Objects[i-1].MinSize().Height/2 + pad
		}
		offset += s.vbox.Objects[i].MinSize().Height / 2
	}
	return offset
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

func (s *SyncedLyricsViewer) setLineText(line *widget.RichText, text string) {
	line.Segments[0].(*widget.TextSegment).Text = text
	line.Refresh()
}

func (s *SyncedLyricsViewer) setLineColor(rt *widget.RichText, colorName fyne.ThemeColorName) {
	rt.Segments[0].(*widget.TextSegment).Style.ColorName = colorName
	rt.Refresh()
}

func (s *SyncedLyricsViewer) checkStopAnimation() bool {
	if s.anim != nil {
		s.anim.Stop()
		s.anim = nil
		return true
	}
	return false
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
	widget.BaseWidget

	Height float32
}

func NewVSpace(height float32) *vSpace {
	v := &vSpace{Height: height}
	v.ExtendBaseWidget(v)
	return v
}

func (v *vSpace) MinSize() fyne.Size {
	return fyne.NewSize(0, v.Height)
}

func (v *vSpace) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(layout.NewSpacer())
}
