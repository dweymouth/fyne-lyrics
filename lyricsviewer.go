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

// LyricsViewer is a widget for displaying song lyrics.
// It supports synced and unsynced mode. In synced mode, the active line
// is highlighted and the widget can advance to the next line
// with an animated scroll. In unsynced mode all lyrics are shown
// in the foreground color and the user is allowed to scroll freely.
type LyricsViewer struct {
	widget.BaseWidget
	mutex sync.Mutex

	lines  []string
	synced bool

	// one-indexed - 0 means before the first line
	// during an animation, currentLine is the line
	// that will be scrolled when the animation is finished
	currentLine int

	singleLineLyricHeight float32

	scroll *NoScroll
	vbox   *fyne.Container

	// nil when an animation is not currently running
	anim            *fyne.Animation
	animStartOffset float32
}

// NewSyncedLyricsViewer returns a new lyrics viewer.
func NewSyncedLyricsViewer() *LyricsViewer {
	s := &LyricsViewer{}
	s.ExtendBaseWidget(s)
	return s
}

// SetLyrics sets the lyrics and also resets the current line to 0 if synced.
func (l *LyricsViewer) SetLyrics(lines []string, synced bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.lines = lines
	l.synced = synced
	l.currentLine = 0
	if l.scroll != nil {
		direction := container.ScrollVerticalOnly
		if synced {
			direction = container.ScrollNone
		}
		l.scroll.Direction = direction
	}
	l.updateContent()
}

// SetCurrentLine sets the current line that the lyric viewer is scrolled to.
// Argument is *one-indexed* - SetCurrentLine(0) means setting the scroll to be
// before the first line. In unsynced mode this is a no-op. This function is
// typically called when the user has seeked the playing song to a new position.
func (l *LyricsViewer) SetCurrentLine(line int) {
	if l.vbox == nil || !l.synced {
		l.currentLine = line
		return // renderer not created yet or unsynced mode
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.checkStopAnimation() && l.currentLine > 1 {
		// we were in the middle of animation
		// make sure prev line is right color
		l.setLineColor(l.vbox.Objects[l.currentLine-1].(*widget.RichText), theme.ColorNameDisabled, true)
	}
	l.setLineColor(l.vbox.Objects[l.currentLine].(*widget.RichText), theme.ColorNameDisabled, true)
	l.currentLine = line
	l.setLineColor(l.vbox.Objects[l.currentLine].(*widget.RichText), theme.ColorNameForeground, true)
	l.scroll.Offset.Y = l.offsetForLine(l.currentLine)
	l.scroll.Refresh()
}

// NextLine advances the lyric viewer to the next line with an animated scroll.
// In unsynced mode this is a no-op.
func (l *LyricsViewer) NextLine() {
	if l.vbox == nil || !l.synced {
		return // no renderer yet, or unsynced lyrics (no-op)
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.currentLine == len(l.lines) {
		return // already at last line
	}
	if l.checkStopAnimation() {
		// we were in the middle of animation - short-circuit it to completed
		// make sure prev and current lines are right color and scrolled to the end
		if l.currentLine > 1 {
			l.setLineColor(l.vbox.Objects[l.currentLine-1].(*widget.RichText), theme.ColorNameDisabled, true)
		}
		l.setLineColor(l.vbox.Objects[l.currentLine].(*widget.RichText), theme.ColorNameForeground, true)
		l.scroll.Offset.Y = l.offsetForLine(l.currentLine)
	}
	l.currentLine++

	var prevLine, nextLine *widget.RichText
	if l.currentLine > 1 {
		prevLine = l.vbox.Objects[l.currentLine-1].(*widget.RichText)
	}
	if l.currentLine <= len(l.lines) {
		nextLine = l.vbox.Objects[l.currentLine].(*widget.RichText)
	}

	l.setupScrollAnimation(prevLine, nextLine)
	l.anim.Start()
}

func (l *LyricsViewer) Refresh() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.updateContent()
}

func (l *LyricsViewer) MinSize() fyne.Size {
	// overridden because NoScroll will have minSize encompass the full lyrics
	minHeight := l.singleLineLyricHeight*3 + theme.Padding()*2
	return fyne.NewSize(l.BaseWidget.MinSize().Width, minHeight)
}

func (l *LyricsViewer) Resize(size fyne.Size) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.updateSpacerSize(size)
	l.BaseWidget.Resize(size)
	if l.anim == nil {
		l.scroll.Offset = fyne.NewPos(0, l.offsetForLine(l.currentLine))
		l.scroll.Refresh()
	} else {
		// animation is running - update its reference scroll pos
		l.animStartOffset = l.offsetForLine(l.currentLine - 1)
	}
}

func (l *LyricsViewer) updateSpacerSize(size fyne.Size) {
	if l.vbox == nil {
		return // renderer not created yet
	}

	var topSpaceHeight, bottomSpaceHeight float32
	if l.synced {
		topSpaceHeight = theme.Padding() + (size.Height-l.singleLineLyricHeight)/2
		// end spacer only needs to be big enough - can't be too big
		// so use a very simple height calculation
		bottomSpaceHeight = size.Height
	}
	l.vbox.Objects[0].(*vSpace).Height = topSpaceHeight
	l.vbox.Objects[len(l.vbox.Objects)-1].(*vSpace).Height = bottomSpaceHeight
}

func (l *LyricsViewer) updateContent() {
	if l.vbox == nil {
		return // renderer not created yet
	}
	l.checkStopAnimation()

	lnObj := len(l.vbox.Objects)
	if lnObj == 0 {
		l.vbox.Objects = append(l.vbox.Objects, NewVSpace(0), NewVSpace(0))
		lnObj = 2
	}
	l.updateSpacerSize(l.Size())
	endSpacer := l.vbox.Objects[lnObj-1]
	for i, line := range l.lines {
		lineNum := i + 1 // one-indexed
		if lineNum < lnObj-1 {
			rt := l.vbox.Objects[lineNum].(*widget.RichText)
			color := theme.ColorNameDisabled
			if !l.synced || l.currentLine == lineNum {
				color = theme.ColorNameForeground
			}
			l.setLineColor(rt, color, false)
			l.setLineText(rt, line)
		} else if (i + 1) < lnObj {
			// replacing end spacer (last element in Objects) with a new richtext
			l.vbox.Objects[i+1] = l.newLyricLine(line)
		} else {
			// extending the Objects slice
			l.vbox.Objects = append(l.vbox.Objects, l.newLyricLine(line))
		}
	}
	for i := len(l.lines) + 1; i < lnObj; i++ {
		l.vbox.Objects[i] = nil
	}
	l.vbox.Objects = l.vbox.Objects[:len(l.lines)+1]
	l.vbox.Objects = append(l.vbox.Objects, endSpacer)
	l.vbox.Refresh()
	l.scroll.Offset.Y = l.offsetForLine(l.currentLine)
	l.scroll.Refresh()
}

func (l *LyricsViewer) setupScrollAnimation(currentLine, nextLine *widget.RichText) {
	// calculate total scroll distance for the animation
	scrollDist := theme.Padding()
	if currentLine != nil {
		scrollDist += currentLine.Size().Height / 2
	} else {
		scrollDist += l.singleLineLyricHeight / 2
	}
	if nextLine != nil {
		scrollDist += nextLine.Size().Height / 2
	} else {
		scrollDist += l.singleLineLyricHeight / 2
	}

	l.animStartOffset = l.scroll.Offset.Y
	var alreadyUpdated bool
	l.anim = fyne.NewAnimation(140*time.Millisecond, func(f float32) {
		l.mutex.Lock()
		defer l.mutex.Unlock()
		l.scroll.Offset.Y = l.animStartOffset + f*scrollDist
		l.scroll.Refresh()
		if !alreadyUpdated && f >= 0.5 {
			if nextLine != nil {
				l.setLineColor(nextLine, theme.ColorNameForeground, true)
			}
			if currentLine != nil {
				l.setLineColor(currentLine, theme.ColorNameDisabled, true)
			}
			alreadyUpdated = true
		}
		if f == 1 /*end of animation*/ {
			l.anim = nil
		}
	})
	l.anim.Curve = fyne.AnimationEaseInOut
}

func (l *LyricsViewer) offsetForLine(lineNum int /*one-indexed*/) float32 {
	if lineNum == 0 {
		return 0
	}
	pad := theme.Padding()
	offset := pad + l.singleLineLyricHeight/2
	for i := 1; i <= lineNum; i++ {
		if i > 1 {
			offset += l.vbox.Objects[i-1].MinSize().Height/2 + pad
		}
		offset += l.vbox.Objects[i].MinSize().Height / 2
	}
	return offset
}

func (l *LyricsViewer) newLyricLine(text string) *widget.RichText {
	ts := &widget.TextSegment{
		Text:  text,
		Style: widget.RichTextStyleSubHeading,
	}
	ts.Style.ColorName = theme.ColorNameDisabled
	rt := widget.NewRichText(ts)
	rt.Wrapping = fyne.TextWrapWord
	return rt
}

func (l *LyricsViewer) setLineText(line *widget.RichText, text string) {
	line.Segments[0].(*widget.TextSegment).Text = text
	line.Refresh()
}

func (l *LyricsViewer) setLineColor(rt *widget.RichText, colorName fyne.ThemeColorName, refresh bool) {
	rt.Segments[0].(*widget.TextSegment).Style.ColorName = colorName
	if refresh {
		rt.Refresh()
	}
}

func (l *LyricsViewer) checkStopAnimation() bool {
	if l.anim != nil {
		l.anim.Stop()
		l.anim = nil
		return true
	}
	return false
}

func (l *LyricsViewer) CreateRenderer() fyne.WidgetRenderer {
	l.singleLineLyricHeight = l.newLyricLine("W").MinSize().Height
	l.vbox = container.NewVBox()
	l.scroll = NewNoScroll(l.vbox)
	if !l.synced {
		l.scroll.Direction = container.ScrollVerticalOnly
	}
	l.updateContent()
	return widget.NewSimpleRenderer(l.scroll)
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

func (n *NoScroll) Scrolled(e *fyne.ScrollEvent) {
	if n.Direction != container.ScrollNone {
		n.Scroll.Scrolled(e)
	}
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
