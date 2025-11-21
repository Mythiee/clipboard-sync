package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                            = windows.NewLazySystemDLL("user32.dll")
	kernel32                          = windows.NewLazySystemDLL("kernel32.dll")
	procCreateWindowExW               = user32.NewProc("CreateWindowExW")
	procDefWindowProcW                = user32.NewProc("DefWindowProcW")
	procDispatchMessageW              = user32.NewProc("DispatchMessageW")
	procGetMessageW                   = user32.NewProc("GetMessageW")
	procTranslateMessage              = user32.NewProc("TranslateMessage")
	procRegisterClassExW              = user32.NewProc("RegisterClassExW")
	procAddClipboardFormatListener    = user32.NewProc("AddClipboardFormatListener")
	procRemoveClipboardFormatListener = user32.NewProc("RemoveClipboardFormatListener")
	procOpenClipboard                 = user32.NewProc("OpenClipboard")
	procCloseClipboard                = user32.NewProc("CloseClipboard")
	procGetClipboardData              = user32.NewProc("GetClipboardData")
	procGlobalLock                    = kernel32.NewProc("GlobalLock")
	procGlobalUnlock                  = kernel32.NewProc("GlobalUnlock")
	procGetModuleHandleW              = kernel32.NewProc("GetModuleHandleW")
)

const (
	wmClipboardUpdate = 0x031D // windows clipboard message
	cfUnicodeText     = 13     // The character encoding of the Windows clipboard is UTF-16
)

// WNDCLASSEX the window of windows
type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr // handle window message
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     windows.Handle
	HIcon         windows.Handle
	HCursor       windows.Handle
	HbrBackground windows.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       windows.Handle
}

type MSG struct {
	Hwnd    windows.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

func wndProc(hwnd windows.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	if msg == wmClipboardUpdate {
		text, err := getClipboardText()
		if err == nil {
			fmt.Println("clipboard updated:", text)
		} else {
			fmt.Println("read clipboard failed:", err)
		}
	}
	ret, _, _ := procDefWindowProcW.Call(
		uintptr(hwnd),
		uintptr(msg),
		wParam,
		lParam,
	)
	return ret
}

func getClipboardText() (string, error) {
	r, _, err := procOpenClipboard.Call(0)
	if r == 0 {
		return "", err
	}
	defer procCloseClipboard.Call()

	h, _, _ := procGetClipboardData.Call(uintptr(cfUnicodeText))
	if h == 0 {
		return "", fmt.Errorf("GetClipboardData return 0")
	}

	ptr, _, _ := procGlobalLock.Call(h)
	if ptr == 0 {
		return "", fmt.Errorf("GlobalLock failed")
	}
	defer procGlobalUnlock.Call(h)

	text := syscall.UTF16ToString((*[1 << 20]uint16)(unsafe.Pointer(ptr))[:])
	return text, nil
}

func main() {
	hInstance, _, _ := procGetModuleHandleW.Call(0)

	className, _ := windows.UTF16PtrFromString("MyHiddenWindowClass")
	windowName, _ := windows.UTF16PtrFromString("MyHiddenWindow")

	var wcex WNDCLASSEX
	wcex.CbSize = uint32(unsafe.Sizeof(wcex))
	wcex.LpfnWndProc = syscall.NewCallback(wndProc)
	wcex.HInstance = windows.Handle(hInstance)
	wcex.LpszClassName = className

	ret, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wcex)))
	if ret == 0 {
		fmt.Println("RegisterClassExW failed:", err)
		return
	}

	hwnd, _, err := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		0, // hide the window
		0, 0, 0, 0,
		0, 0,
		hInstance,
		0,
	)
	if hwnd == 0 {
		fmt.Println("CreateWindowExW failed:", err)
		return
	}

	// register clipboard listener
	r, _, err := procAddClipboardFormatListener.Call(hwnd)
	if r == 0 {
		fmt.Println("AddClipboardFormatListener failed:", err)
		return
	}
	defer procRemoveClipboardFormatListener.Call(hwnd)

	fmt.Println("Listing Clipboard...")

	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}
