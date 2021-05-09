package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/blang/mpv"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

var mpvCmd *exec.Cmd
var mpvClient *mpv.Client

func playMpv(wid uint32, mpvStack *gtk.Stack, url string) {
	if mpvCmd == nil {
		widParam := fmt.Sprintf("--wid=%d", wid)
		mpvCmd = exec.Command("mpv", widParam, "--idle", "--input-ipc-server=/tmp/mpvsocket")
		err := mpvCmd.Start()
		if err != nil {
			log.Println(err)
		}
	}
	if mpvClient == nil {
		ipcc := getIPCC("/tmp/mpvsocket")
		for ipcc == nil {
			time.Sleep(time.Millisecond * 250)
			ipcc = getIPCC("/tmp/mpvsocket")
		}
		mpvClient = mpv.NewClient(ipcc) // Highlevel client, can also use RPCClient
	} else {
		mpvClient.SetPause(false)
	}
	err := mpvClient.Loadfile(url, mpv.LoadListModeReplace)
	if err != nil {
		log.Println(err)
	}
	coreIdle, _ := mpvClient.GetBoolProperty("core-idle")
	for coreIdle {
		time.Sleep(time.Millisecond * 250)
		coreIdle, _ = mpvClient.GetBoolProperty("core-idle")
	}
	mpvStack.SetVisibleChildName(playerPage)
}

func getIPCC(socket string) *mpv.IPCClient {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	return mpv.NewIPCClient(socket)
}

func startSpinner() {
	obj, _ := builder.GetObject("spinner")
	spinner := obj.(*gtk.Spinner)
	spinner.Start()
}

func stopSpinner() {
	obj, _ := builder.GetObject("spinner")
	spinner := obj.(*gtk.Spinner)
	spinner.Stop()
}

var fscrn bool

func toggleFullscreen() {
	obj, _ := builder.GetObject("stack")
	stack := obj.(*gtk.Stack)
	if stack.GetVisibleChildName() != "channels_page" {
		return
	}
	fscrn = !fscrn
	obj, _ = builder.GetObject("window")
	wnd := obj.(*gtk.Window)
	obj, _ = builder.GetObject("sidebar")
	sidebar := obj.(*gtk.ScrolledWindow)
	obj, _ = builder.GetObject("status_bar")
	statusbar := obj.(*gtk.Box)
	if fscrn {
		wnd.Fullscreen()
		sidebar.Hide()
		statusbar.Hide()
	} else {
		wnd.Unfullscreen()
		sidebar.Show()
		statusbar.Show()
	}
}

func keypress(wnd *gtk.Window, event *gdk.Event) bool {
	evk := gdk.EventKeyNewFromEvent(event)
	key := evk.KeyVal()
	/*ctrl := evk.State() & gdk.CONTROL_MASK
	if ctrl != 0 && key == gdk.KEY_q {
		appExit()
		app.Quit()
	}*/
	if (fscrn && key == gdk.KEY_Escape) || key == gdk.KEY_F11 {
		toggleFullscreen()
	}
	return false
}
