package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	appTitle = "Oracle Report Group Finder v4"

	WS_OVERLAPPED  = 0x00000000
	WS_CAPTION     = 0x00C00000
	WS_SYSMENU     = 0x00080000
	WS_THICKFRAME  = 0x00040000
	WS_MINIMIZEBOX = 0x00020000
	WS_MAXIMIZEBOX = 0x00010000
	WS_VISIBLE     = 0x10000000
	WS_CHILD       = 0x40000000
	WS_TABSTOP     = 0x00010000
	WS_VSCROLL     = 0x00200000
	WS_HSCROLL     = 0x00100000
	WS_BORDER      = 0x00800000

	WS_EX_CLIENTEDGE = 0x00000200

	ES_AUTOHSCROLL  = 0x0080
	ES_MULTILINE    = 0x0004
	ES_AUTOVSCROLL  = 0x0040
	ES_AUTOHSCROLL2 = 0x0080
	ES_READONLY     = 0x0800

	BS_PUSHBUTTON = 0x00000000

	CBS_DROPDOWN    = 0x0002
	CBS_AUTOHSCROLL = 0x0040

	SW_SHOW       = 5
	SW_SHOWNORMAL = 1

	CW_USEDEFAULT = ^uintptr(0x7fffffff)

	WM_CREATE    = 0x0001
	WM_DESTROY   = 0x0002
	WM_SIZE      = 0x0005
	WM_COMMAND   = 0x0111
	WM_CLOSE     = 0x0010
	WM_SETFONT   = 0x0030
	WM_APP       = 0x8000
	WM_LOAD_DONE = WM_APP + 1

	BN_CLICKED = 0

	EM_SETREADONLY = 0x00CF
	EM_SETSEL      = 0x00B1
	EM_REPLACESEL  = 0x00C2

	CB_ADDSTRING    = 0x0143
	CB_RESETCONTENT = 0x014B
	CB_SETCURSEL    = 0x014E

	SWP_NOZORDER = 0x0004

	MB_OK              = 0x00000000
	MB_ICONERROR       = 0x00000010
	MB_ICONINFORMATION = 0x00000040

	OFN_FILEMUSTEXIST = 0x00001000
	OFN_PATHMUSTEXIST = 0x00000800
	OFN_EXPLORER      = 0x00080000

	ID_CONVERTER_BROWSE = 1001
	ID_REPORT_BROWSE    = 1002
	ID_LOAD             = 1003
	ID_SEARCH           = 1004
	ID_OPEN_FOLDER      = 1005

	COLOR_WINDOW     = 5
	IDC_ARROW        = 32512
	IDI_APPLICATION  = 32512
	DEFAULT_GUI_FONT = 17
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")

	procRegisterClassExW   = user32.NewProc("RegisterClassExW")
	procCreateWindowExW    = user32.NewProc("CreateWindowExW")
	procDefWindowProcW     = user32.NewProc("DefWindowProcW")
	procShowWindow         = user32.NewProc("ShowWindow")
	procUpdateWindow       = user32.NewProc("UpdateWindow")
	procGetMessageW        = user32.NewProc("GetMessageW")
	procTranslateMessage   = user32.NewProc("TranslateMessage")
	procDispatchMessageW   = user32.NewProc("DispatchMessageW")
	procPostQuitMessage    = user32.NewProc("PostQuitMessage")
	procSendMessageW       = user32.NewProc("SendMessageW")
	procPostMessageW       = user32.NewProc("PostMessageW")
	procSetWindowTextW     = user32.NewProc("SetWindowTextW")
	procGetWindowTextW     = user32.NewProc("GetWindowTextW")
	procGetWindowTextLenW  = user32.NewProc("GetWindowTextLengthW")
	procMoveWindow         = user32.NewProc("MoveWindow")
	procEnableWindow       = user32.NewProc("EnableWindow")
	procMessageBoxW        = user32.NewProc("MessageBoxW")
	procLoadCursorW        = user32.NewProc("LoadCursorW")
	procLoadIconW          = user32.NewProc("LoadIconW")
	procGetClientRect      = user32.NewProc("GetClientRect")
	procSetFocus           = user32.NewProc("SetFocus")
	procSetProcessDPIAware = user32.NewProc("SetProcessDPIAware")

	procGetModuleHandleW   = kernel32.NewProc("GetModuleHandleW")
	procGetModuleFileNameW = kernel32.NewProc("GetModuleFileNameW")
	procGetStockObject     = gdi32.NewProc("GetStockObject")
	procGetOpenFileNameW   = comdlg32.NewProc("GetOpenFileNameW")
	procShellExecuteW      = shell32.NewProc("ShellExecuteW")
)

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     syscall.Handle
	HIcon         syscall.Handle
	HCursor       syscall.Handle
	HbrBackground syscall.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       syscall.Handle
}

type POINT struct{ X, Y int32 }
type MSG struct {
	Hwnd    syscall.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}
type RECT struct{ Left, Top, Right, Bottom int32 }

type OPENFILENAME struct {
	LStructSize       uint32
	HwndOwner         syscall.Handle
	HInstance         syscall.Handle
	LpstrFilter       *uint16
	LpstrCustomFilter *uint16
	NMaxCustFilter    uint32
	NFilterIndex      uint32
	LpstrFile         *uint16
	NMaxFile          uint32
	LpstrFileTitle    *uint16
	NMaxFileTitle     uint32
	LpstrInitialDir   *uint16
	LpstrTitle        *uint16
	Flags             uint32
	NFileOffset       uint16
	NFileExtension    uint16
	LpstrDefExt       *uint16
	LCustData         uintptr
	LpfnHook          uintptr
	LpTemplateName    *uint16
	PvReserved        unsafe.Pointer
	DwReserved        uint32
	FlagsEx           uint32
}

type Node struct {
	Name     string
	Attr     map[string]string
	Parent   *Node
	Children []*Node
}

type FrameRef struct {
	Name  string
	Depth int
}

type GroupUsage struct {
	GroupName    string
	ObjectName   string
	ObjectType   string
	Section      string
	ParentFrames []FrameRef
	ChildFrames  []FrameRef
}

type ParsedReport struct {
	Groups       []string
	SourceFrames []*Node
}

type loadResult struct {
	parsed  *ParsedReport
	xmlPath string
	err     error
}

var app struct {
	hwnd            syscall.Handle
	converterLabel  syscall.Handle
	converterEdit   syscall.Handle
	converterBrowse syscall.Handle
	reportLabel     syscall.Handle
	reportEdit      syscall.Handle
	reportBrowse    syscall.Handle
	userIDLabel     syscall.Handle
	userIDEdit      syscall.Handle
	groupLabel      syscall.Handle
	groupCombo      syscall.Handle
	loadBtn         syscall.Handle
	searchBtn       syscall.Handle
	resultsEdit     syscall.Handle
	statusStatic    syscall.Handle
	openFolderBtn   syscall.Handle
	parsed          *ParsedReport
	loadedXML       string
	mu              sync.Mutex
	pending         *loadResult
}

func main() {
	runtime.LockOSThread()
	procSetProcessDPIAware.Call()

	hInstance, _, _ := procGetModuleHandleW.Call(0)
	className := utf16Ptr("OracleReportGroupFinderWindow")
	title := utf16Ptr(appTitle)
	cursor, _, _ := procLoadCursorW.Call(0, uintptr(IDC_ARROW))
	icon, _, _ := procLoadIconW.Call(0, uintptr(IDI_APPLICATION))

	wc := WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEX{})),
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     syscall.Handle(hInstance),
		HIcon:         syscall.Handle(icon),
		HCursor:       syscall.Handle(cursor),
		HbrBackground: syscall.Handle(COLOR_WINDOW + 1),
		LpszClassName: className,
		HIconSm:       syscall.Handle(icon),
	}
	atom, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 {
		fatalBox(fmt.Sprintf("Could not register window class: %v", err))
		return
	}

	hwnd, _, err := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU|WS_THICKFRAME|WS_MINIMIZEBOX|WS_MAXIMIZEBOX|WS_VISIBLE,
		CW_USEDEFAULT, CW_USEDEFAULT, 1060, 720,
		0, 0, hInstance, 0,
	)
	if hwnd == 0 {
		fatalBox(fmt.Sprintf("Could not create window: %v", err))
		return
	}
	app.hwnd = syscall.Handle(hwnd)
	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)

	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_CREATE:
		app.hwnd = syscall.Handle(hwnd)
		createControls(syscall.Handle(hwnd))
		loadConfig()
		if strings.TrimSpace(getText(app.converterEdit)) == "" {
			if found := findReportsConverter(""); found != "" {
				setText(app.converterEdit, found)
			}
		}
		layoutControls(syscall.Handle(hwnd))
		return 0
	case WM_SIZE:
		layoutControls(syscall.Handle(hwnd))
		return 0
	case WM_COMMAND:
		id := int(wParam & 0xffff)
		code := int((wParam >> 16) & 0xffff)
		if code == BN_CLICKED || id == ID_SEARCH {
			switch id {
			case ID_CONVERTER_BROWSE:
				if p := openFileDialog(syscall.Handle(hwnd), "Select Oracle Reports converter", "Oracle Reports Converter\x00RWCON60.EXE;rwconverter.exe;rwconverter.bat;rwconverter.cmd\x00All files\x00*.*\x00\x00"); p != "" {
					setText(app.converterEdit, p)
					saveConfig(p)
				}
			case ID_REPORT_BROWSE:
				if p := openFileDialog(syscall.Handle(hwnd), "Select Oracle Report", "Oracle Reports\x00*.rdf;*.rex;*.xml\x00RDF files\x00*.rdf\x00REX files\x00*.rex\x00XML files\x00*.xml\x00All files\x00*.*\x00\x00"); p != "" {
					setText(app.reportEdit, p)
				}
			case ID_LOAD:
				beginLoad()
			case ID_SEARCH:
				searchAndRender()
			case ID_OPEN_FOLDER:
				openXMLFolder()
			}
		}
		return 0
	case WM_LOAD_DONE:
		finishLoad()
		return 0
	case WM_CLOSE:
		procDefWindowProcW.Call(hwnd, WM_CLOSE, wParam, lParam)
		return 0
	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func createControls(hwnd syscall.Handle) {
	font, _, _ := procGetStockObject.Call(DEFAULT_GUI_FONT)

	createStatic := func(text string) syscall.Handle {
		h := createWindow("STATIC", text, WS_CHILD|WS_VISIBLE, 0, 0)
		procSendMessageW.Call(uintptr(h), WM_SETFONT, font, 1)
		return h
	}
	createEdit := func(readOnly bool, multiline bool) syscall.Handle {
		style := uint32(WS_CHILD | WS_VISIBLE | WS_TABSTOP | ES_AUTOHSCROLL)
		ex := uint32(WS_EX_CLIENTEDGE)
		if multiline {
			style = WS_CHILD | WS_VISIBLE | WS_TABSTOP | WS_VSCROLL | WS_HSCROLL | ES_MULTILINE | ES_AUTOVSCROLL | ES_AUTOHSCROLL2
		}
		if readOnly {
			style |= ES_READONLY
		}
		h := createWindowEx(ex, "EDIT", "", style, 0, 0)
		procSendMessageW.Call(uintptr(h), WM_SETFONT, font, 1)
		return h
	}
	createButton := func(text string, id int) syscall.Handle {
		h := createWindow("BUTTON", text, WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, id, 0)
		procSendMessageW.Call(uintptr(h), WM_SETFONT, font, 1)
		return h
	}

	app.converterLabel = createStatic("Oracle Reports converter:")
	app.converterEdit = createEdit(false, false)
	app.converterBrowse = createButton("Browse...", ID_CONVERTER_BROWSE)

	app.reportLabel = createStatic("Report (.rdf/.rex/.xml):")
	app.reportEdit = createEdit(false, false)
	app.reportBrowse = createButton("Browse...", ID_REPORT_BROWSE)

	app.userIDLabel = createStatic("USERID (optional):")
	app.userIDEdit = createEdit(false, false)
	app.loadBtn = createButton("Convert && Load", ID_LOAD)

	app.groupLabel = createStatic("Group:")
	app.groupCombo = createWindow("COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_VSCROLL|CBS_DROPDOWN|CBS_AUTOHSCROLL, 0, 0)
	procSendMessageW.Call(uintptr(app.groupCombo), WM_SETFONT, font, 1)
	app.searchBtn = createButton("Search Group", ID_SEARCH)
	procEnableWindow.Call(uintptr(app.groupCombo), 0)
	procEnableWindow.Call(uintptr(app.searchBtn), 0)

	app.resultsEdit = createEdit(true, true)
	app.statusStatic = createStatic("Select an .rdf, .rex or .xml report.")
	app.openFolderBtn = createButton("Open Converted Folder", ID_OPEN_FOLDER)
	procEnableWindow.Call(uintptr(app.openFolderBtn), 0)
}

func createWindow(class, text string, style uint32, id int, ex uint32) syscall.Handle {
	return createWindowEx(ex, class, text, style, id, 0)
}
func createWindowEx(ex uint32, class, text string, style uint32, id int, unused uintptr) syscall.Handle {
	h, _, _ := procCreateWindowExW.Call(
		uintptr(ex), uintptr(unsafe.Pointer(utf16Ptr(class))), uintptr(unsafe.Pointer(utf16Ptr(text))), uintptr(style),
		0, 0, 100, 25, uintptr(app.hwnd), uintptr(id), 0, 0,
	)
	return syscall.Handle(h)
}

func layoutControls(hwnd syscall.Handle) {
	var rc RECT
	procGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rc)))
	w := int(rc.Right - rc.Left)
	h := int(rc.Bottom - rc.Top)
	if w < 700 {
		w = 700
	}
	if h < 500 {
		h = 500
	}

	margin, gap, labelW, btnW, rowH := 12, 8, 165, 112, 26
	xField := margin + labelW + gap
	fieldW := w - xField - btnW - gap - margin
	y := margin

	move(app.converterLabel, margin, y+4, labelW, rowH)
	move(app.converterEdit, xField, y, fieldW, rowH)
	move(app.converterBrowse, xField+fieldW+gap, y, btnW, rowH)
	y += rowH + 8
	move(app.reportLabel, margin, y+4, labelW, rowH)
	move(app.reportEdit, xField, y, fieldW, rowH)
	move(app.reportBrowse, xField+fieldW+gap, y, btnW, rowH)
	y += rowH + 8
	move(app.userIDLabel, margin, y+4, labelW, rowH)
	move(app.userIDEdit, xField, y, fieldW, rowH)
	move(app.loadBtn, xField+fieldW+gap, y, btnW, rowH)
	y += rowH + 14
	move(app.groupLabel, margin, y+4, labelW, rowH)
	move(app.groupCombo, xField, y, fieldW, 240)
	move(app.searchBtn, xField+fieldW+gap, y, btnW, rowH)
	y += rowH + 12

	bottomH := 34
	resultsH := h - y - bottomH - margin
	if resultsH < 120 {
		resultsH = 120
	}
	move(app.resultsEdit, margin, y, w-2*margin, resultsH)
	y += resultsH + 6
	move(app.statusStatic, margin, y+5, w-margin*2-btnW-gap, rowH)
	move(app.openFolderBtn, w-margin-btnW, y, btnW, rowH)
}

func move(h syscall.Handle, x, y, w, hh int) {
	if h != 0 {
		procMoveWindow.Call(uintptr(h), uintptr(x), uintptr(y), uintptr(w), uintptr(hh), 1)
	}
}

func beginLoad() {
	report := strings.TrimSpace(getText(app.reportEdit))
	if report == "" {
		errorBox("Select an RDF, REX or XML report first.")
		return
	}
	if _, err := os.Stat(report); err != nil {
		errorBox("Report file was not found.")
		return
	}

	setBusy(true, "Loading report...")
	converter := strings.TrimSpace(getText(app.converterEdit))
	userID := strings.TrimSpace(getText(app.userIDEdit))

	go func() {
		var parsedPath string
		var err error
		ext := strings.ToLower(filepath.Ext(report))
		switch ext {
		case ".xml", ".rex":
			parsedPath = report
		case ".rdf":
			if converter == "" || !fileExists(converter) {
				converter = findReportsConverter(converter)
			}
			if converter == "" || !fileExists(converter) {
				err = errors.New("Oracle Reports converter was not found. Select RWCON60.EXE / rwconverter.exe using Browse.")
			} else {
				saveConfig(converter)
				parsedPath, err = convertRDFToREX(converter, report, userID)
			}
		default:
			err = errors.New("Supported report files are .rdf, .rex and .xml.")
		}

		var parsed *ParsedReport
		if err == nil {
			parsed, err = parseReport(parsedPath)
		}
		app.mu.Lock()
		app.pending = &loadResult{parsed: parsed, xmlPath: parsedPath, err: err}
		app.mu.Unlock()
		procPostMessageW.Call(uintptr(app.hwnd), WM_LOAD_DONE, 0, 0)
	}()
}

func finishLoad() {
	app.mu.Lock()
	res := app.pending
	app.pending = nil
	app.mu.Unlock()
	if res == nil {
		setBusy(false, "Ready.")
		return
	}
	if res.err != nil {
		setBusy(false, "Load failed.")
		errorBox(res.err.Error())
		return
	}
	app.parsed = res.parsed
	app.loadedXML = res.xmlPath
	setText(app.converterEdit, strings.TrimSpace(getText(app.converterEdit)))
	procSendMessageW.Call(uintptr(app.groupCombo), CB_RESETCONTENT, 0, 0)
	for _, g := range app.parsed.Groups {
		p := utf16Ptr(g)
		procSendMessageW.Call(uintptr(app.groupCombo), CB_ADDSTRING, 0, uintptr(unsafe.Pointer(p)))
	}
	setText(app.groupCombo, "")
	procEnableWindow.Call(uintptr(app.groupCombo), 1)
	procEnableWindow.Call(uintptr(app.searchBtn), 1)
	procEnableWindow.Call(uintptr(app.openFolderBtn), 1)
	text := fmt.Sprintf("Loaded definition: %s\r\nGroups found: %d\r\nFrames with Source found: %d\r\n\r\nChoose or type a Group and click Search Group.", res.xmlPath, len(app.parsed.Groups), len(app.parsed.SourceFrames))
	setText(app.resultsEdit, text)
	setBusy(false, fmt.Sprintf("Loaded: %s | Groups: %d", filepath.Base(res.xmlPath), len(app.parsed.Groups)))
	procSetFocus.Call(uintptr(app.groupCombo))
}

func searchAndRender() {
	if app.parsed == nil {
		return
	}
	group := strings.TrimSpace(getText(app.groupCombo))
	if group == "" {
		errorBox("Type or choose a Group.")
		return
	}
	usages := findUsages(app.parsed, group)
	var b strings.Builder
	fmt.Fprintf(&b, "GROUP: %s\r\nFrames with Source: %d\r\n\r\n", group, len(usages))
	if len(usages) == 0 {
		fmt.Fprintf(&b, "No Frame or Repeating Frame found with source = %s\r\n", group)
		sugg := suggestGroups(app.parsed, group)
		if len(sugg) > 0 {
			b.WriteString("\r\nPossible Groups:\r\n")
			for i, g := range sugg {
				if i >= 15 {
					break
				}
				fmt.Fprintf(&b, "  - %s\r\n", g)
			}
		}
	} else {
		for _, u := range usages {
			fmt.Fprintf(&b, "%s: %s\r\n", u.ObjectType, u.ObjectName)
			if u.Section != "" {
				fmt.Fprintf(&b, "  Section: %s\r\n", u.Section)
			}
			b.WriteString("  Parent Frames:\r\n")
			if len(u.ParentFrames) == 0 {
				b.WriteString("    <none>\r\n")
			} else {
				for _, f := range u.ParentFrames {
					fmt.Fprintf(&b, "    - %s\r\n", f.Name)
				}
			}
			b.WriteString("  Child Frames:\r\n")
			if len(u.ChildFrames) == 0 {
				b.WriteString("    <none>\r\n")
			} else {
				for _, f := range u.ChildFrames {
					fmt.Fprintf(&b, "    %s- %s\r\n", strings.Repeat("  ", max(0, f.Depth-1)), f.Name)
				}
			}
			b.WriteString("\r\n")
		}
	}
	setText(app.resultsEdit, b.String())
	setBusy(false, fmt.Sprintf("Group %s: %d frame(s) with matching Source", group, len(usages)))
}

func setBusy(busy bool, status string) {
	en := uintptr(1)
	if busy {
		en = 0
	}
	procEnableWindow.Call(uintptr(app.loadBtn), en)
	if app.parsed != nil {
		procEnableWindow.Call(uintptr(app.searchBtn), en)
		procEnableWindow.Call(uintptr(app.groupCombo), en)
	}
	setText(app.statusStatic, status)
}

func openXMLFolder() {
	if app.loadedXML == "" {
		errorBox("No converted/loaded report has been loaded yet.")
		return
	}
	folder := filepath.Dir(app.loadedXML)
	verb := utf16Ptr("open")
	target := utf16Ptr(folder)
	r, _, _ := procShellExecuteW.Call(uintptr(app.hwnd), uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(target)), 0, 0, SW_SHOWNORMAL)
	if r <= 32 {
		errorBox("Could not open converted report folder.")
	}
}

func openFileDialog(owner syscall.Handle, title, filter string) string {
	buf := make([]uint16, 32768)
	titlep := utf16Ptr(title)
	// filter uses escaped literal \x00 sequences in Go string above; convert them to NULs.
	filter = strings.ReplaceAll(filter, "\\x00", "\x00")
	fp := utf16FromStringWithNuls(filter)
	ofn := OPENFILENAME{LStructSize: uint32(unsafe.Sizeof(OPENFILENAME{})), HwndOwner: owner, LpstrFilter: &fp[0], LpstrFile: &buf[0], NMaxFile: uint32(len(buf)), LpstrTitle: titlep, Flags: OFN_FILEMUSTEXIST | OFN_PATHMUSTEXIST | OFN_EXPLORER}
	r, _, _ := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

func parseReport(path string) (*ParsedReport, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".xml":
		return parseXMLReport(path)
	case ".rex":
		return parseREXReport(path)
	default:
		return nil, fmt.Errorf("unsupported parsed report format: %s", ext)
	}
}

func parseXMLReport(path string) (*ParsedReport, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := xml.NewDecoder(bufio.NewReaderSize(f, 128*1024))
	dec.Strict = false
	root := &Node{Name: "root", Attr: map[string]string{}}
	stack := []*Node{root}
	var sourceFrames []*Node
	groupSet := map[string]string{}
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("XML parse error: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			n := &Node{Name: strings.ToLower(t.Name.Local), Attr: map[string]string{}, Parent: stack[len(stack)-1]}
			for _, a := range t.Attr {
				n.Attr[strings.ToLower(a.Name.Local)] = strings.TrimSpace(a.Value)
			}
			n.Parent.Children = append(n.Parent.Children, n)
			stack = append(stack, n)
			if n.Name == "group" {
				if name := attr(n, "name"); name != "" {
					groupSet[strings.ToLower(name)] = name
				}
			}
			if n.Name == "frame" || n.Name == "repeatingframe" {
				if src := attr(n, "source"); src != "" {
					sourceFrames = append(sourceFrames, n)
					groupSet[strings.ToLower(src)] = src
				}
			}
		case xml.EndElement:
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	groups := sortedValues(groupSet)
	return &ParsedReport{Groups: groups, SourceFrames: sourceFrames}, nil
}

type rexObject struct {
	Type  string
	Props map[string]string
}

type rexFrame struct {
	itemID      string
	name        string
	groupID     string
	sourceName  string
	isRepeating bool
	parentID    string
	section     string
	x           float64
	y           float64
	w           float64
	h           float64
	hasRect     bool
	node        *Node
}

var rexAssignRE = regexp.MustCompile(`^\s*([A-Za-z0-9_]+)\s*=\s*(.*)$`)
var rexDefineRE = regexp.MustCompile(`(?i)^\s*DEFINE\s+([A-Za-z0-9_]+)\s*$`)
var rexTypePrefixRE = regexp.MustCompile(`^\([A-Za-z0-9_]+\)\s*`)

func parseREXReport(path string) (*ParsedReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	objects := parseREXObjects(string(data))
	if len(objects) == 0 {
		return nil, errors.New("REX file was created, but no DEFINE blocks were recognized.")
	}

	groupByID := map[string]string{}
	groupSet := map[string]string{}
	for _, o := range objects {
		if !isREXGroupType(o.Type) {
			continue
		}
		name := firstProp(o.Props, "NAME", "GROUP_NAME")
		id := normalizeID(firstProp(o.Props, "ITEMID", "OBJECT_ID", "ID"))
		if name != "" {
			groupSet[strings.ToLower(name)] = name
			if id != "" {
				groupByID[id] = name
			}
		}
	}

	var frames []*rexFrame
	frameByID := map[string]*rexFrame{}
	for _, o := range objects {
		if !isREXFrameType(o.Type) {
			continue
		}
		f := &rexFrame{}
		f.isRepeating = isREXRepeatingFrameType(o.Type)
		f.itemID = normalizeID(firstProp(o.Props, "ITEMID", "OBJECT_ID", "ID"))
		f.name = firstProp(o.Props, "NAME", "FRAME_NAME")
		if f.name == "" {
			f.name = "<unnamed frame>"
		}
		f.groupID = normalizeID(firstProp(o.Props, "GROUP_ID", "GROUPID", "SOURCE_ID"))
		f.sourceName = firstProp(o.Props, "SOURCE_NAME", "GROUP_NAME", "SOURCE")
		if !looksLikeGroupName(f.sourceName) {
			f.sourceName = ""
		}
		if f.sourceName == "" && f.groupID != "" {
			f.sourceName = groupByID[f.groupID]
		}
		f.parentID = normalizeID(firstProp(o.Props,
			"PARENT_FRAME_ID", "PARENT_ITEM_ID", "PARENT_ITEMID", "ENCLOSING_FRAME_ID", "PARENT_ID"))
		f.section = firstProp(o.Props, "SECTION", "SECTION_NAME", "REGION", "REGION_NAME", "LAYOUT_SECTION")
		f.x, f.y, f.w, f.h, f.hasRect = rexRect(o.Props)
		frames = append(frames, f)
		if f.itemID != "" {
			frameByID[f.itemID] = f
		}
		if f.sourceName != "" {
			groupSet[strings.ToLower(f.sourceName)] = f.sourceName
		}
	}

	if len(frames) == 0 {
		return nil, fmt.Errorf("REX was created successfully, but no SRW2_FRAME/frame objects were recognized. Converted file: %s", path)
	}

	root := &Node{Name: "root", Attr: map[string]string{}}
	var sourceFrames []*Node
	for _, f := range frames {
		tag := "frame"
		if f.isRepeating {
			tag = "repeatingframe"
		}
		n := &Node{Name: tag, Attr: map[string]string{"name": f.name}}
		if f.sourceName != "" {
			n.Attr["source"] = f.sourceName
			sourceFrames = append(sourceFrames, n)
		}
		if f.section != "" {
			n.Attr["section"] = f.section
		}
		f.node = n
	}

	// First prefer an explicit parent relationship when the REX contains one.
	for _, f := range frames {
		if f.parentID == "" {
			continue
		}
		if p := frameByID[f.parentID]; p != nil && p != f {
			attachNode(p.node, f.node)
		}
	}

	// Oracle 6i REX does not have a documented public grammar. If no explicit
	// enclosing-frame id is present, infer ownership from the layout rectangles.
	for _, child := range frames {
		if child.node.Parent != nil || !child.hasRect {
			continue
		}
		var best *rexFrame
		bestArea := 0.0
		for _, candidate := range frames {
			if candidate == child || !candidate.hasRect {
				continue
			}
			if child.section != "" && candidate.section != "" && !strings.EqualFold(child.section, candidate.section) {
				continue
			}
			if !rectContains(candidate, child) {
				continue
			}
			area := candidate.w * candidate.h
			if best == nil || area < bestArea {
				best = candidate
				bestArea = area
			}
		}
		if best != nil {
			attachNode(best.node, child.node)
		}
	}
	for _, f := range frames {
		if f.node.Parent == nil {
			attachNode(root, f.node)
		}
	}

	groups := sortedValues(groupSet)
	if len(groups) == 0 || len(sourceFrames) == 0 {
		return nil, fmt.Errorf("REX was created, but no Frame/Repeating Frame with a resolvable Source Group was found. Converted file: %s", path)
	}
	return &ParsedReport{Groups: groups, SourceFrames: sourceFrames}, nil
}

func parseREXObjects(text string) []rexObject {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	var out []rexObject
	var cur *rexObject
	var multiKey string
	var multi strings.Builder
	flush := func() {
		if cur != nil {
			out = append(out, *cur)
		}
		cur = nil
		multiKey = ""
		multi.Reset()
	}
	for _, line := range lines {
		if m := rexDefineRE.FindStringSubmatch(line); m != nil {
			flush()
			cur = &rexObject{Type: strings.ToUpper(strings.TrimSpace(m[1])), Props: map[string]string{}}
			continue
		}
		if cur == nil {
			continue
		}
		if multiKey != "" {
			if multi.Len() > 0 {
				multi.WriteString("\n")
			}
			multi.WriteString(line)
			if strings.Contains(line, ">>") {
				cur.Props[multiKey] = normalizeREXValue(multi.String())
				multiKey = ""
				multi.Reset()
			}
			continue
		}
		m := rexAssignRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		key := strings.ToUpper(strings.TrimSpace(m[1]))
		val := strings.TrimSpace(m[2])
		if strings.HasPrefix(val, "<<") && !strings.Contains(val, ">>") {
			multiKey = key
			multi.WriteString(val)
			continue
		}
		cur.Props[key] = normalizeREXValue(val)
	}
	flush()
	return out
}

func normalizeREXValue(v string) string {
	v = strings.TrimSpace(v)
	v = rexTypePrefixRE.ReplaceAllString(v, "")
	v = strings.TrimSpace(v)
	upper := strings.ToUpper(v)
	if v == "" || upper == "NULL" || upper == "NULLP" || strings.HasSuffix(upper, " NULLP") {
		return ""
	}
	if strings.HasPrefix(v, "<<") && strings.HasSuffix(v, ">>") {
		v = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(v, "<<"), ">>"))
	}
	v = strings.TrimSpace(v)
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		v = v[1 : len(v)-1]
	}
	return strings.TrimSpace(v)
}

func isREXGroupType(t string) bool {
	t = strings.ToUpper(strings.TrimSpace(t))
	return t == "SRW2_GROUP" || t == "GROUP" || strings.HasSuffix(t, "_GROUP")
}

func isREXFrameType(t string) bool {
	t = strings.ToUpper(strings.TrimSpace(t))
	return t == "SRW2_FRAME" || t == "FRAME" || t == "REPEATING_FRAME" || t == "REPEATINGFRAME"
}

func isREXRepeatingFrameType(t string) bool {
	t = strings.ToUpper(strings.TrimSpace(t))
	return t == "REPEATING_FRAME" || t == "REPEATINGFRAME"
}

func firstProp(m map[string]string, keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(m[strings.ToUpper(k)]); v != "" {
			return v
		}
	}
	return ""
}

func normalizeID(v string) string {
	v = normalizeREXValue(v)
	if v == "" || v == "0" || strings.EqualFold(v, "NULLP") {
		return ""
	}
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		if f == float64(int64(f)) {
			return strconv.FormatInt(int64(f), 10)
		}
	}
	return strings.TrimSpace(v)
}

func looksLikeGroupName(v string) bool {
	v = strings.TrimSpace(v)
	return len(v) >= 2 && (strings.HasPrefix(strings.ToUpper(v), "G_") || strings.HasPrefix(strings.ToUpper(v), "GROUP"))
}

func rexNumber(m map[string]string, keys ...string) (float64, bool) {
	v := firstProp(m, keys...)
	if v == "" {
		return 0, false
	}
	v = strings.TrimSpace(rexTypePrefixRE.ReplaceAllString(v, ""))
	f, err := strconv.ParseFloat(v, 64)
	return f, err == nil
}

func rexRect(m map[string]string) (x, y, w, h float64, ok bool) {
	var ox, oy, ow, oh bool
	x, ox = rexNumber(m, "X", "X_POS", "XPOS")
	y, oy = rexNumber(m, "Y", "Y_POS", "YPOS")
	w, ow = rexNumber(m, "WD", "WIDTH", "WID")
	h, oh = rexNumber(m, "HT", "HEIGHT", "HGT")
	ok = ox && oy && ow && oh && w > 0 && h > 0
	return
}

func rectContains(parent, child *rexFrame) bool {
	if parent == nil || child == nil || !parent.hasRect || !child.hasRect {
		return false
	}
	eps := 0.0001
	contains := child.x+eps >= parent.x && child.y+eps >= parent.y &&
		child.x+child.w <= parent.x+parent.w+eps && child.y+child.h <= parent.y+parent.h+eps
	if !contains {
		return false
	}
	// Equal rectangles are not treated as ownership.
	return parent.w*parent.h > child.w*child.h+eps
}

func attachNode(parent, child *Node) {
	if parent == nil || child == nil || parent == child || child.Parent != nil {
		return
	}
	child.Parent = parent
	parent.Children = append(parent.Children, child)
}

func sortedValues(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i]) < strings.ToLower(out[j]) })
	return out
}

func findUsages(r *ParsedReport, group string) []GroupUsage {
	target := strings.TrimSpace(group)
	var out []GroupUsage
	for _, frame := range r.SourceFrames {
		src := attr(frame, "source")
		if !strings.EqualFold(src, target) {
			continue
		}
		section := attr(frame, "section")
		if section == "" {
			section = findAncestorName(frame, "section")
		}
		objectType := "Frame"
		fallback := "<unnamed frame>"
		if frame.Name == "repeatingframe" {
			objectType = "Repeating Frame"
			fallback = "<unnamed repeatingFrame>"
		}
		out = append(out, GroupUsage{
			GroupName:    src,
			ObjectName:   fallbackName(frame, fallback),
			ObjectType:   objectType,
			Section:      section,
			ParentFrames: collectAncestorFrames(frame),
			ChildFrames:  collectDescendantFrames(frame),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ObjectType != out[j].ObjectType {
			return out[i].ObjectType < out[j].ObjectType
		}
		return strings.ToLower(out[i].ObjectName) < strings.ToLower(out[j].ObjectName)
	})
	return out
}
func suggestGroups(r *ParsedReport, text string) []string {
	n := strings.ToLower(strings.TrimSpace(text))
	if n == "" {
		return r.Groups
	}
	var o []string
	for _, g := range r.Groups {
		if strings.Contains(strings.ToLower(g), n) {
			o = append(o, g)
		}
	}
	return o
}
func collectAncestorFrames(n *Node) []FrameRef {
	var o []FrameRef
	level := 1
	for p := n.Parent; p != nil; p = p.Parent {
		if p.Name == "frame" || p.Name == "repeatingframe" {
			fallback := "<unnamed frame>"
			if p.Name == "repeatingframe" {
				fallback = "<unnamed repeatingFrame>"
			}
			o = append(o, FrameRef{fallbackName(p, fallback), level})
			level++
		}
	}
	return o
}
func collectDescendantFrames(n *Node) []FrameRef {
	var o []FrameRef
	var walk func(*Node, int)
	walk = func(p *Node, depth int) {
		for _, c := range p.Children {
			if c.Name == "frame" || c.Name == "repeatingframe" {
				fallback := "<unnamed frame>"
				if c.Name == "repeatingframe" {
					fallback = "<unnamed repeatingFrame>"
				}
				o = append(o, FrameRef{fallbackName(c, fallback), depth})
				walk(c, depth+1)
			} else {
				walk(c, depth)
			}
		}
	}
	walk(n, 1)
	return o
}
func findAncestorName(n *Node, tag string) string {
	for p := n.Parent; p != nil; p = p.Parent {
		if p.Name == tag {
			return fallbackName(p, tag)
		}
	}
	return ""
}
func fallbackName(n *Node, f string) string {
	if v := attr(n, "name"); v != "" {
		return v
	}
	return f
}
func attr(n *Node, k string) string {
	if n == nil {
		return ""
	}
	return strings.TrimSpace(n.Attr[strings.ToLower(k)])
}

func convertRDFToREX(converter, rdf, userID string) (string, error) {
	tempRoot := chooseASCIITempRoot()
	workDir, err := os.MkdirTemp(tempRoot, "ORGF_")
	if err != nil {
		return "", fmt.Errorf("could not create temporary conversion folder: %w", err)
	}
	keepTemp := false
	defer func() {
		if !keepTemp {
			_ = os.RemoveAll(workDir)
		}
	}()

	tempRDF := filepath.Join(workDir, "input.rdf")
	tempREX := filepath.Join(workDir, "output.rex")
	logPath := filepath.Join(workDir, "rwcon.log")
	if err := copyFile(rdf, tempRDF); err != nil {
		return "", fmt.Errorf("could not copy RDF to temporary ASCII path: %w", err)
	}

	args := []string{}
	if strings.TrimSpace(userID) != "" {
		args = append(args, "userid="+strings.TrimSpace(userID))
	}
	args = append(args,
		"stype=rdffile",
		"source="+tempRDF,
		"dtype=rexfile",
		"dest="+tempREX,
		"overwrite=yes",
		"batch=yes",
		"logfile="+logPath,
	)

	var cmd *exec.Cmd
	ext := strings.ToLower(filepath.Ext(converter))
	if ext == ".bat" || ext == ".cmd" {
		a := append([]string{"/d", "/c", "call", converter}, args...)
		cmd = exec.Command("cmd.exe", a...)
	} else {
		cmd = exec.Command(converter, args...)
	}
	cmd.Dir = workDir
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("could not start Oracle Reports converter: %w", err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	deadline := time.Now().Add(120 * time.Second)
	lastSize := int64(-1)
	stableCount := 0
	processDone := false
	var runErr error

	for time.Now().Before(deadline) {
		if !processDone {
			select {
			case runErr = <-done:
				processDone = true
			default:
			}
		}
		if st, statErr := os.Stat(tempREX); statErr == nil && st.Size() > 0 {
			if st.Size() == lastSize {
				stableCount++
			} else {
				lastSize = st.Size()
				stableCount = 0
			}
			if stableCount >= 3 {
				if !processDone && cmd.Process != nil {
					_ = cmd.Process.Kill()
					select {
					case <-done:
					case <-time.After(2 * time.Second):
					}
				}
				break
			}
		}
		if processDone && !fileExistsNonEmpty(tempREX) {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if !fileExistsNonEmpty(tempREX) {
		if !processDone && cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		logText, _ := os.ReadFile(logPath)
		return "", fmt.Errorf("RDF → REX conversion did not create a REX file.\r\nConverter: %s\r\nError: %v\r\n\r\nOracle output:\r\n%s\r\n\r\nOracle log:\r\n%s",
			filepath.Base(converter), runErr, output.String(), string(logText))
	}

	base := strings.TrimSuffix(filepath.Base(rdf), filepath.Ext(rdf))
	finalREX := filepath.Join(filepath.Dir(rdf), base+"_groupfinder.rex")
	_ = os.Remove(finalREX)
	if err := copyFile(tempREX, finalREX); err != nil {
		keepTemp = true
		return tempREX, nil
	}
	return finalREX, nil
}

func chooseASCIITempRoot() string {
	var candidates []string
	if t := os.TempDir(); t != "" {
		candidates = append(candidates, t)
	}
	if drive := strings.TrimSpace(os.Getenv("SystemDrive")); drive != "" {
		candidates = append(candidates, filepath.Join(drive+string(os.PathSeparator), "Temp"))
	}
	for _, root := range candidates {
		if root == "" || !isASCII(root) {
			continue
		}
		if err := os.MkdirAll(root, 0755); err != nil {
			continue
		}
		test := filepath.Join(root, "orgf_write_test.tmp")
		if err := os.WriteFile(test, []byte("x"), 0600); err == nil {
			_ = os.Remove(test)
			return root
		}
	}
	return os.TempDir()
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func fileExists(p string) bool { st, e := os.Stat(p); return e == nil && !st.IsDir() }
func fileExistsNonEmpty(p string) bool {
	st, e := os.Stat(p)
	return e == nil && !st.IsDir() && st.Size() > 0
}

func findReportsConverter(preferred string) string {
	var candidates []string
	add := func(p string) {
		if p != "" {
			candidates = append(candidates, p)
		}
	}
	if preferred != "" {
		add(preferred)
	}
	if oh := os.Getenv("ORACLE_HOME"); oh != "" {
		for _, n := range []string{"RWCON60.EXE", "rwcon60.exe", "rwconverter.exe", "rwconverter.bat", "rwconverter.cmd"} {
			add(filepath.Join(oh, "bin", n))
		}
	}
	for _, n := range []string{"RWCON60.EXE", "rwcon60.exe", "rwconverter.exe", "rwconverter.bat"} {
		if p, err := exec.LookPath(n); err == nil {
			add(p)
		}
	}
	if rb, err := exec.LookPath("rwbuilder.exe"); err == nil {
		for _, n := range []string{"RWCON60.EXE", "rwcon60.exe", "rwconverter.exe", "rwconverter.bat", "rwconverter.cmd"} {
			add(filepath.Join(filepath.Dir(rb), n))
		}
	}
	if rb, err := exec.LookPath("rwbl60.exe"); err == nil {
		for _, n := range []string{"RWCON60.EXE", "rwcon60.exe"} {
			add(filepath.Join(filepath.Dir(rb), n))
		}
	}
	for _, p := range candidates {
		if fileExists(p) {
			return p
		}
	}
	return ""
}

func configPath() string {
	exe := make([]uint16, 32768)
	r, _, _ := procGetModuleFileNameW.Call(0, uintptr(unsafe.Pointer(&exe[0])), uintptr(len(exe)))
	if r == 0 {
		return "OracleReportGroupFinder.ini"
	}
	return filepath.Join(filepath.Dir(syscall.UTF16ToString(exe[:r])), "OracleReportGroupFinder.ini")
}
func loadConfig() {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "rwconverter=") {
			p := strings.TrimSpace(line[len("rwconverter="):])
			if p != "" {
				setText(app.converterEdit, p)
			}
		}
	}
}
func saveConfig(p string) {
	if strings.TrimSpace(p) == "" {
		return
	}
	_ = os.WriteFile(configPath(), []byte("rwconverter="+p+"\r\n"), 0644)
}

func getText(h syscall.Handle) string {
	if h == 0 {
		return ""
	}
	n, _, _ := procGetWindowTextLenW.Call(uintptr(h))
	buf := make([]uint16, n+1)
	procGetWindowTextW.Call(uintptr(h), uintptr(unsafe.Pointer(&buf[0])), n+1)
	return syscall.UTF16ToString(buf)
}
func setText(h syscall.Handle, s string) {
	if h == 0 {
		return
	}
	p := utf16Ptr(s)
	procSetWindowTextW.Call(uintptr(h), uintptr(unsafe.Pointer(p)))
}
func errorBox(s string) { messageBox(s, MB_OK|MB_ICONERROR) }
func infoBox(s string)  { messageBox(s, MB_OK|MB_ICONINFORMATION) }
func fatalBox(s string) { messageBox(s, MB_OK|MB_ICONERROR) }
func messageBox(s string, flags uintptr) {
	title := utf16Ptr(appTitle)
	msg := utf16Ptr(s)
	procMessageBoxW.Call(uintptr(app.hwnd), uintptr(unsafe.Pointer(msg)), uintptr(unsafe.Pointer(title)), flags)
}
func utf16Ptr(s string) *uint16 { return syscall.StringToUTF16Ptr(s) }
func utf16FromStringWithNuls(s string) []uint16 {
	r := []rune(s)
	out := make([]uint16, 0, len(r)+2)
	for _, rr := range r {
		if rr == 0 {
			out = append(out, 0)
		} else {
			u := syscall.StringToUTF16(string(rr))
			out = append(out, u[:len(u)-1]...)
		}
	}
	if len(out) < 2 || out[len(out)-1] != 0 || out[len(out)-2] != 0 {
		out = append(out, 0, 0)
	}
	return out
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
