package main

import (
	"fmt"
	"log"
	"sort"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func showFavPage(page string) {
	obj, _ := builder.GetObject("fav_stack")
	stack, _ := obj.(*gtk.Stack)
	stack.SetVisibleChildName(page)
}

func updateFavBox() {
	favs := config.favorites
	obj, _ := builder.GetObject("favorites_flowbox")
	fb, _ := obj.(*gtk.FlowBox)
	if len(favs) == 0 {
		obj, _ = builder.GetObject("fav_label")
		lbl, _ := obj.(*gtk.Label)
		mRefProvider, _ := gtk.CssProviderNew()
		css := `
		label {
			background-color: rgb(0,128,128);
			color:            rgb(0,0,0);
		  }`
		if err := mRefProvider.LoadFromData(css); err != nil {
			log.Println(err)
		}
		ctx, _ := lbl.GetStyleContext()
		ctx.AddProvider(mRefProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
		lbl.ShowAll()
		lbl.SetVisible(true)
		showFavPage(favTextPage)
		return
	} else {
		showFavPage(favoritesPage)
	}
	if (int)(fb.GetChildren().Length()) == len(favs) {
		return
	}
	deleteChildren(fb.Container)
	for i := 0; i < len(favs); i++ {
		ch1 := favs[i]
		btn := makeChannelBtn(ch1, nil)
		btn.Connect("button-press-event", func(b *gtk.Button, ev *gdk.Event) bool {
			mb := gdk.EventButtonNewFromEvent(ev)
			switch mb.Button() {
			case gdk.BUTTON_PRIMARY:
				favClicked(ch1)
				return true
			case gdk.BUTTON_MIDDLE:
				return true
			case gdk.BUTTON_SECONDARY:
				favPopupDel(b, ev, ch1, nil)
				return true
			default:
				return false
			}
		})
		fb.Add(btn)
	}
	fb.ShowAll()
	fb.SetVisible(true)
}

func favClicked(ch2 *channel) {
	curPage = channelsPage
	obj, _ := builder.GetObject("channels_flowbox")
	fb, _ := obj.(*gtk.FlowBox)
	deleteChildren(fb.Container)
	var imgMap = make(map[string]*gtk.Image)
	for i := range config.favorites {
		ch1 := config.favorites[i]
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
				favPopupDel(b, ev, ch1, nil)
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
	prevPage = landingPage
	showButton("go_back_button", true)
	showButton("full_screen_button", true)
	channelClicked(ch2)
	dispFavs = true
	setTitle()
	// if len(imgMap) != 0 {
	// go setChannelImages(imgMap, g, groupDir)
	// }
}

func setFavColor(btn *gtk.Button) {

	mRefProvider, _ := gtk.CssProviderNew()
	css := `
	button {
		background-color: rgb(191,191,191);
		color:            rgb(0, 0, 0);
	  }`
	if err := mRefProvider.LoadFromData(css); err != nil {
		log.Println(err)
	}
	ctx, _ := btn.GetStyleContext()
	ctx.AddProvider(mRefProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
}

func isFavChannel(ch1 *channel) bool {
	for i := range config.favorites {
		if config.favorites[i] == ch1 {
			return true
		}
	}
	return false
}

func addToFavorites(ch1 *channel, g *group) {
	if config.favorites == nil {
		config.favorites = make([]*channel, 0)
	}
	config.favorites = append(config.favorites, ch1)
	sort.Slice(config.favorites[:], func(i, j int) bool {
		return config.favorites[i].title < config.favorites[j].title
	})
	groupClicked(g, false)
	msg := fmt.Sprintf("Channel '%s' added to Favorites", ch1.title)
	setLabelText("status_label", msg)
}

func removeFromFav(ch1 *channel) {
	if len(config.favorites) == 1 {
		config.favorites = nil
	} else {
		i := 0
		for config.favorites[i] != ch1 {
			i++
		}
		config.favorites = append(config.favorites[:i], config.favorites[i+1:]...)
	}
}

func favPopupAdd(b *gtk.Button, ev *gdk.Event, ch1 *channel, g *group) {
	menu, _ := gtk.MenuNew()
	item, _ := gtk.MenuItemNewWithLabel("Add to Favorites")
	item.Show()
	item.Connect("activate", func() { addToFavorites(ch1, g) })
	menu.Add(item)
	menu.PopupAtPointer(ev)
}

func favPopupDel(b *gtk.Button, ev *gdk.Event, ch1 *channel, g *group) {
	menu, _ := gtk.MenuNew()
	item, _ := gtk.MenuItemNewWithLabel("Delete from Favorites")
	item.Show()
	item.Connect("activate", func() {
		removeFromFav(ch1)
		switch {
		case curPage == landingPage:
			updateFavBox()
		case g == nil:
			if config.favorites == nil {
				updateFavBox()
				if mpvClient != nil {
					mpvClient.SetPause(true)
				}
				showPage(landingPage)
			} else {
				favClicked(ch1)
			}
		default:
			groupClicked(g, false)
		}
		msg := fmt.Sprintf("Channel '%s' deleted from Favorites", ch1.title)
		setLabelText("status_label", msg)
	})
	menu.Add(item)
	menu.PopupAtPointer(ev)
}
