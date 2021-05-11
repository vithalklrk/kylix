package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

var curPage string = landingPage
var prevPage string
var curProvider string
var tvTypes = []string{"Channels", "Series", "Movies"}
var curTvType int
var curGroup *group
var curChannel *channel
var dispFavs bool

const shareDir = "/usr/share/kylix"
const picDir = shareDir + "/pictures"
const kylixIcon = picDir + "/kylix.jpg"

func backButtonClicked() {
	if mpvClient != nil {
		mpvClient.SetPause(true)
	}
	showPage(prevPage)
	switch prevPage {
	case landingPage:
		showButton("full_screen_button", false)
		dispFavs = false
		setTitle()
		updateFavBox()
		showButton("go_back_button", false)
		showButton("full_screen_button", false)
	case categoriesPage:
		showButton("full_screen_button", false)
		dispFavs = false
		setTitle()
		stopSpinner()
		prevPage = landingPage
		showButton("full_screen_button", false)
		setLabelText("status_label", "")
	case channelsPage:
		showButton("full_screen_button", true)
		setTitle()
		prevPage = categoriesPage
	}
}

func channelsClicked() {
	curTvType = channels
	showGroups()
}

func moviesClicked() {
	curTvType = movies
	showGroups()
}

func seriesClicked() {
	curTvType = series
	showGroups()
}

var builder *gtk.Builder

func mainWnd(app *gtk.Application) {
	const gladeFile = shareDir + "/kylix.ui"
	var err error
	builder, err = gtk.BuilderNewFromFile(gladeFile)
	if err != nil {
		log.Fatalf("Failed to create gtk builder from %s -- %s\n", gladeFile, err)
	}

	obj, err := builder.GetObject("window")
	wnd := obj.(*gtk.Window)
	wnd.Move(config.wndX, config.wndY)
	wnd.SetDefaultSize(config.wndWd, config.wndHt)
	wnd.SetIconFromFile(kylixIcon)
	setTitle()
	wnd.ShowAll()
	app.AddWindow(wnd)
	showButton("full_screen_button", false)
	getButton("full_screen_button").Connect("clicked", toggleFullscreen)
	makeMainMenu()
	showButton("provider_button", false)
	showButton("settings_button", false)
	showButton("go_back_button", false)
	showBox("playback_bar", false)

	setPicture("channels_logo", picDir+"/Channels.png", 96)
	s := fmt.Sprintf("Channels(%d)", getNChannels(channels))
	setLabelText("channels_label", s)
	getButton("tv_button").Connect("clicked", channelsClicked)

	setPicture("series_logo", picDir+"/Series.jpeg", 96)
	s = fmt.Sprintf("Series(%d)", getNChannels(series))
	setLabelText("series_label", s)
	getButton("series_button").Connect("clicked", seriesClicked)

	setPicture("movies_logo", picDir+"/Movies.png", 96)
	s = fmt.Sprintf("Movies(%d)", getNChannels(movies))
	setLabelText("movies_label", s)
	getButton("movies_button").Connect("clicked", moviesClicked)

	getButton("go_back_button").Connect("clicked", backButtonClicked)
	obj, _ = builder.GetObject("favorites_box")
	box, _ := obj.(*gtk.Box)
	mRefProvider, _ := gtk.CssProviderNew()
	css := `
	box {
		background-color: rgb(0,128,128);
		color:            rgb(255,255,255);
	  }`
	if err = mRefProvider.LoadFromData(css); err != nil {
		log.Println(err)
	}
	ctx, _ := box.GetStyleContext()
	ctx.AddProvider(mRefProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
	updateFavBox()

	obj, _ = builder.GetObject("spinner")
	spinner := obj.(*gtk.Spinner)
	mRefProvider, _ = gtk.CssProviderNew()
	css = `
	spinner {
		background-color: rgb(0,0,0);
		color:            rgb(128,128,128);
	  }`
	if err = mRefProvider.LoadFromData(css); err != nil {
		log.Println(err)
	}
	ctx, _ = spinner.GetStyleContext()
	ctx.AddProvider(mRefProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
	wnd.Connect("configure-event", sizeChange)
	wnd.Connect("key-press-event", keypress)
	wnd.Connect("destroy", appExit)
}

func makeMainMenu() {
	acc, _ := gtk.AccelGroupNew()
	obj, _ := builder.GetObject("window")
	wnd := obj.(*gtk.Window)
	obj, _ = builder.GetObject("main_menu")
	menu := obj.(*gtk.Menu)
	wnd.AddAccelGroup(acc)
	item, _ := gtk.MenuItemNewWithLabel("About")
	item.Connect("activate", about)
	item.AddAccelerator("activate", acc, gdk.KEY_F1, 0, gtk.ACCEL_VISIBLE)
	menu.Append(item)
	item, _ = gtk.MenuItemNewWithLabel("Quit")
	item.Connect("activate", quit)
	item.AddAccelerator("activate", acc, gdk.KEY_q, gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE)
	menu.Append(item)
	menu.ShowAll()
}

func about() {
	dlg, _ := gtk.AboutDialogNew()
	dlg.SetTitle("About")
	dlg.SetProgramName("Kylix")
	dlg.SetComments("Watch ipTV")
	dlg.SetLicense(getGPL())
	dlg.SetVersion("1.0")
	dlg.SetWebsite("https://github.com/vithalklrk/kylix")
	dlg.SetAuthors([]string{"Vithal Kuchibhotla", "", "Channel source from", "https://iptv-org.github.io/iptv/channels.json"})
	pixBuf, _ := gdk.PixbufNewFromFileAtSize(picDir+"/kylix.jpg", 128, 128)
	dlg.SetLogo(pixBuf)
	dlg.Connect("response", func(dlg *gtk.AboutDialog, resp int) {
		dlg.Destroy()
	})
	dlg.Show()
}

func getGPL() string {
	licFile := "/usr/share/common-licenses/GPL"
	if stat, err := os.Stat(licFile); err == nil {
		inp := make([]byte, stat.Size())
		f, _ := os.Open(licFile)
		f.Read(inp)
		f.Close()
		return string(inp)
	}
	return ""
}

func quit() {
	obj, _ := builder.GetObject("window")
	wnd := obj.(*gtk.Window)
	app, _ := wnd.GetApplication()
	appExit()
	app.Quit()
}
