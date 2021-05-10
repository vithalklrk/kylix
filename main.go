package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const landingPage = "landing_page"
const favTextPage = "fav_text_page"
const favoritesPage = "favorites_page"
const channelsPage = "channels_page"
const categoriesPage = "categories_page"
const spinnerPage = "spinner_page"
const playerPage = "player_page"

func showButton(name string, show bool) {
	btn := getButton(name)
	// btn.SetTooltipText(prevPage)
	btn.SetVisible(show)
}

func showBox(name string, show bool) {
	obj, err := builder.GetObject(name)
	if err != nil {
		log.Fatalln("Couldn't get box ", name, err)
	}
	box, ok := obj.(*gtk.Box)
	if !ok {
		log.Fatalln("not box")
	}
	box.SetVisible(show)
}

func setLabelText(name string, text string) {
	obj, err := builder.GetObject(name)
	if err != nil {
		log.Fatalln("Couldn't get label ", err)
	}
	lbl, _ := obj.(*gtk.Label)
	lbl.SetLabel(text)
}

func setPicture(logo string, fileName string, dim int) {
	obj, err := builder.GetObject(logo)
	if err != nil {
		log.Fatalln("Logo not found:", logo)
	}
	img, ok := obj.(*gtk.Image)
	if !ok {
		log.Fatalln("not image:", logo)
	}
	pixBuf, err := gdk.PixbufNewFromFileAtSize(fileName, dim, dim)
	img.SetFromPixbuf(pixBuf)
}

func getButton(s string) *gtk.Button {
	obj, err := builder.GetObject(s)
	if err != nil {
		log.Fatalln("Couldn't get full_screen_button ", s, err)
	}
	btn, ok := obj.(*gtk.Button)
	if !ok {
		log.Fatalln("not button:", s)
	}
	return btn
}

func getExtnFromURL(url string) string {
	i := strings.LastIndex(url, ".")
	j := strings.LastIndex(url, "/")
	if i > j {
		return url[i:]
	}
	return ""
}

func sanitize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "/", "_")
	return s
}

func deleteChildren(c gtk.Container) {
	list := c.GetChildren()
	var j uint
	for j = 0; j < list.Length(); j++ {
		item := list.Nth(j)
		w := item.Data().(*gtk.Widget)
		c.Remove(w)
	}
}

var defaultPixBuf *gdk.Pixbuf

func makeChannelBtn(ch1 *channel, imgMap map[string]*gtk.Image) *gtk.Button {
	lbl, _ := gtk.LabelNew(ch1.title)
	btn, _ := gtk.ButtonNew()
	if isFavChannel(ch1) {
		setFavColor(btn)
	}
	img, _ := gtk.ImageNew()
	if ch1.logo == "" {
		ch1.icon = kylixIcon
		if defaultPixBuf == nil {
			defaultPixBuf, _ = gdk.PixbufNewFromFileAtSize(ch1.icon, -1, 12)
		}
		img.SetFromPixbuf(defaultPixBuf)
	} else {
		groupDir := filepath.Join(getConfigPath(), ch1.group)
		if ch1.icon == "" {
			ch1.icon = filepath.Join(groupDir, ch1.title+getExtnFromURL(ch1.logo))
		}
		if _, err := os.Stat(ch1.icon); err == nil {
			pixBuf, _ := gdk.PixbufNewFromFileAtSize(ch1.icon, -1, 12)
			img.SetFromPixbuf(pixBuf)
		} else if imgMap != nil {
			imgMap[ch1.title] = img
		}
	}
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 2)
	box.PackStart(img, false, false, 0)
	box.PackStart(lbl, false, false, 0)
	box.SetSpacing(6)
	btn.Add(box)
	return btn
}

func groupClicked(g *group, clearSubtitle bool) {
	curPage = channelsPage
	curGroup = g
	curChannel = nil
	setTitle()
	names := make([]string, len(g.channels))
	i := 0
	for name := range g.channels {
		names[i] = name
		i++
	}
	sort.Strings(names)
	obj, _ := builder.GetObject("channels_flowbox")
	fb, _ := obj.(*gtk.FlowBox)
	deleteChildren(fb.Container)
	groupDir := filepath.Join(getConfigPath(), g.title)
	var imgMap = make(map[string]*gtk.Image)
	for _, name := range names {
		ch1 := g.channels[name]
		btn := makeChannelBtn(ch1, imgMap)

		btn.Connect("button-press-event", func(b *gtk.Button, ev *gdk.Event) bool {
			mb := gdk.EventButtonNewFromEvent(ev)
			switch mb.Button() {
			case gdk.BUTTON_PRIMARY:
				channelClicked(ch1)
				setTitle()
				return true
			case gdk.BUTTON_MIDDLE:
				return true
			case gdk.BUTTON_SECONDARY:
				if isFavChannel(ch1) {
					favPopupDel(b, ev, ch1, g)
				} else {
					favPopupAdd(b, ev, ch1, g)
				}
				return true
			default:
				return false
			}
		})
		fb.Add(btn)
	}
	fb.ShowAll()
	fb.SetVisible(true)
	showPage(channelsPage)
	prevPage = categoriesPage
	showButton("go_back_button", true)
	showButton("full_screen_button", true)
	if len(imgMap) != 0 {
		go setChannelImages(imgMap, g, groupDir)
	}
}

func setTitle() {
	obj, _ := builder.GetObject("headerbar")
	titleBar, _ := obj.(*gtk.HeaderBar)
	var title, subTitle string
	switch curPage {
	case landingPage:
		title = "Kylix"
		subTitle = curProvider
	case categoriesPage:
		title = curProvider
		subTitle = tvTypes[curTvType]
	case channelsPage:
		title = curProvider
		if dispFavs {
			subTitle = "Favorites"
		} else {
			subTitle = fmt.Sprintf("%s - %s", tvTypes[curTvType], curGroup.title)
		}
		if curChannel != nil {
			subTitle += " - " + curChannel.title
		}
	}
	titleBar.SetTitle(title)
	titleBar.SetSubtitle(subTitle)
}

func channelClicked(ch1 *channel) {
	curChannel = ch1
	dispFavs = false
	obj, _ := builder.GetObject("mpv_stack")
	mpvStack, _ := obj.(*gtk.Stack)
	mpvStack.SetVisibleChildName(spinnerPage)
	go startSpinner()
	obj, _ = builder.GetObject("mpv_drawing_area")
	drawArea, _ := obj.(*gtk.DrawingArea)
	gdkw, _ := drawArea.GetWindow()
	wid := gdkw.GetXID()
	go playMpv(wid, mpvStack, ch1.url)
}

/*
func setImages(imgMap map[string]*gtk.Image) {
	for title := range imgMap {
		g := groups[title]
		img := imgMap[title]
		pixBuf, _ := gdk.PixbufNewFromFileAtSize(g.icon, -1, 12)
		img.SetFromPixbuf(pixBuf)
	}
}
*/

func setChannelImages(imgMap map[string]*gtk.Image, g *group, groupDir string) {
	if _, err := os.Stat(groupDir); err != nil {
		os.Mkdir(groupDir, 0755)
	}
	for title := range imgMap {
		ch1 := g.channels[title]
		img := imgMap[title]
		if ch1.logo != "" {
			resp, err := http.Get(ch1.logo)
			if err == nil && resp.StatusCode == 200 {
				defer resp.Body.Close()
				file, _ := os.Create(ch1.icon)
				defer file.Close()
				io.Copy(file, resp.Body)
			}
		} else {
			ch1.icon = kylixIcon
		}
		pixBuf, _ := gdk.PixbufNewFromFileAtSize(ch1.icon, -1, 12)
		img.SetFromPixbuf(pixBuf)
	}
}

func showPage(page string) {
	obj, _ := builder.GetObject("stack")
	stack, _ := obj.(*gtk.Stack)
	stack.SetVisibleChildName(page)
	curPage = page
}

func showGroups() {
	groups = tv[curTvType]
	curPage = categoriesPage
	setTitle()
	obj, _ := builder.GetObject("categories_flowbox")
	fb, _ := obj.(*gtk.FlowBox)
	deleteChildren(fb.Container)
	titles := make([]string, len(groups))
	i := 0
	for title := range groups {
		titles[i] = title
		i++
	}
	sort.Strings(titles)
	for _, title := range titles {
		g := groups[title]
		s := fmt.Sprintf("%s(%d)", g.title, len(g.channels))
		lbl, _ := gtk.LabelNew(s)
		btn, _ := gtk.ButtonNew()
		btn.Connect("clicked", func() { groupClicked(g, true) })
		img, _ := gtk.ImageNew()
		icon := "/usr/share/kylix/pictures/flags/" + strings.ToLower(ctryIds[title]) + ".png"
		if _, ok := os.Stat(icon); ok != nil {
			icon = "/usr/share/kylix/pictures/kylix.jpg"
		}
		pixBuf, _ := gdk.PixbufNewFromFileAtSize(icon, -1, 12)
		img.SetFromPixbuf(pixBuf)
		box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
		box.PackStart(img, false, false, 0)
		box.PackStart(lbl, false, false, 0)
		box.SetSpacing(6)
		btn.Add(box)
		fb.Add(btn)
	}
	fb.ShowAll()
	fb.SetVisible(true)
	showPage(categoriesPage)
	prevPage = landingPage
	showButton("go_back_button", true)
	showButton("full_screen_button", false)
}

func getNChannels(tvType int) int {
	nCh := 0
	for _, g := range tv[tvType] {
		nCh += len(g.channels)
	}
	return nCh
}

func main() {

	cfgPath := getConfigPath()
	if _, err := os.Stat(cfgPath); err != nil {
		os.MkdirAll(cfgPath, 0755)
	}
	loadFromJson("https://iptv-org.github.io/iptv/channels.json")
	curProvider = "iptv-org"

	// loadFromURL("https://raw.githubusercontent.com/Free-IPTV/Countries/master/ZZ_PLAYLIST_ALL_TV.m3u")
	loadConfig()
	if config.wndHt == 0 {
		config.wndHt = 600
		config.wndWd = 800
	}
	const appdef = "com.vithal.kylix"
	app, err := gtk.ApplicationNew(appdef, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatalln("Failed to create gtk app", err)
	}

	_, err = app.Connect("activate", mainWnd)
	if err != nil {
		log.Fatalln(err)
	}
	glib.IdleAdd(setChannelImages)
	// glib.IdleAdd(setImages)
	app.Run(os.Args)
}
