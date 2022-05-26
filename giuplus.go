package giuplus

import (
	"strings"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/imgui-go"
)

type TextEditor struct {
	multiline  bool
	wordwrap   bool
	autoSelect bool
	width      float32
	height     float32
	text       string
	selStart   int
	selEnd     int
	onActivate func(e *TextEditor)
}

// NewTextEditor creates a new text editor with given height and width. The onActivate callback
// is called when the editor is activated.
func NewTextEditor(width, height float32, onActivate func(e *TextEditor)) *TextEditor {
	return &TextEditor{
		multiline:  false,
		wordwrap:   false,
		width:      width,
		height:     height,
		autoSelect: true,
		onActivate: onActivate,
	}
}

// NewTextEditorMultiline creates a new multiline text editor with given width and height.
// Word-wrapping is ON by default.
func NewTextEditorMultiline(width, height float32) *TextEditor {
	return &TextEditor{
		multiline:  true,
		wordwrap:   true,
		width:      width,
		height:     height,
		autoSelect: true,
	}
}

// SetOnActivate sets the onActivate callback which is called when the editor is activated.
func (e *TextEditor) SetOnActivate(callback func(e *TextEditor)) {
	e.onActivate = callback
}

// Widget returns the undlerying GUI widget, which is either a InputText or InputTextMultiline.
func (e *TextEditor) Widget() g.Widget {
	if e.multiline {
		widget := g.InputTextMultiline(&e.text).Size(-1, 100).
			Flags(imgui.InputTextFlagsCallbackAlways | imgui.InputTextFlagsCallbackCharFilter)
		cbwidget := func(data imgui.InputTextCallbackData) int32 {
			return WrapInputtextMultiline(e, data)
		}
		fullwidget := widget.Callback(cbwidget)
		return fullwidget
	}
	widget := g.InputText(&e.text).Size(-1).
		Callback(func(data imgui.InputTextCallbackData) int32 {
			switch data.EventFlag() {
			case imgui.InputTextFlagsCallbackAlways:
				e.selStart = data.SelectionStart()
				e.selEnd = data.SelectionEnd()
				e.onActivate(e)
			}
			return 0
		})
	return widget
}

// Build builds the widget's layout. This is to satisfy Giu's custom widget interface.
func (e *TextEditor) Build() {
	e.Widget().Build()
}

// SetText sets the editor text to the given UTF-8 string.
func (e *TextEditor) SetText(s string) {
	e.text = s
}

// Text returns the text of the editor.
func (e *TextEditor) Text() string {
	return e.text
}

// SetAutoSelect sets the automatic selection on focus feature.
func (e *TextEditor) SetAutoSelect(on bool) {
	e.autoSelect = on
}

// AutoSelect returns true if text is automatically selected on focus, false otherwise.
func (e *TextEditor) AutoSelect() bool {
	return e.autoSelect
}

// SetWordwrap sets the automatic word wrapping on or off.
func (e *TextEditor) SetWordwrap(on bool) {
	e.wordwrap = on
}

// Wordwrap returns true if the automatic word wrapping is on, false otherwise.
func (e *TextEditor) Wordrap() bool {
	return e.wordwrap
}

// Size returns the width and height of the editor. Widths and heights of -1 indicate maximum stretch.
func (e *TextEditor) Size() (float32, float32) {
	return e.width, e.height
}

// SetSize sets the editor width and height. Widths and heights of -1 indicate maximum stretch.
func (e *TextEditor) SetSize(width, height float32) {
	e.width = width
	e.height = height
}

// WrapInputTextMultiline is a callback to wrap an input text multiline.
func WrapInputtextMultiline(e *TextEditor, data imgui.InputTextCallbackData) int32 {
	switch data.EventFlag() {
	case imgui.InputTextFlagsCallbackCharFilter:
		c := data.EventChar()
		if c == '\n' {
			data.SetEventChar('\u07FF') // pivot character 2-bytes in UTF-8
		}

	case imgui.InputTextFlagsCallbackAlways:
		// 0. turn every pivot byte sequence into \r\n
		buff := data.Buffer()
		buff2 := []byte(strings.ReplaceAll(string(buff), "\u07FF", "\r\n"))
		for i := range buff {
			buff[i] = buff2[i]
		}
		data.MarkBufferModified()

		// 1. zap all newlines that are not preceeded by a CR (which was manually entered like above)
		cr := false
		for i, c := range buff {
			if c == 10 && !cr {
				buff[i] = 32
				data.MarkBufferModified()
			} else {
				if c == 13 {
					cr = true
				} else {
					cr = false
				}
			}
		}
		// 2. word break the whole buffer with the standard greedy algorithm
		nl := 0
		spc := 0
		w := g.GetWidgetWidth(e.Widget())
		for i, c := range buff {
			if c == 10 {
				nl = i
			}
			if c == 32 {
				spc = i
			}
			if TextWidth(string(buff[nl:i])) > w && spc > 0 {
				buff[spc] = 10
				data.MarkBufferModified()
			}
		}
	}
	return 0
}

// TextWidth returns the width of the given text.
func TextWidth(s string) float32 {
	w, _ := g.CalcTextSize(s)
	return w
}

var _ g.Widget = &TextEditor{}
