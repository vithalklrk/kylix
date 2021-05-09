package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

// Configuration data is stored in ~/.config/kylix

var config struct {
	wndX      int
	wndY      int
	wndHt     int
	wndWd     int
	favorites []*channel
}

func getConfigPath() string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	return filepath.Join(dir, ".config/kylix")
}

const settingsFile = "settings.txt"

func saveConfig() {
	path := filepath.Join(getConfigPath(), settingsFile)
	file, err := os.Create(path)
	if err != nil {
		log.Fatalln("Failed to create", path, err)
	}
	writer := bufio.NewWriter(file)
	s := fmt.Sprintf("%d\t%d\t%d\t%d\n", config.wndX, config.wndY, config.wndHt, config.wndWd)
	writer.WriteString(s)
	for i := range config.favorites {
		ch1 := config.favorites[i]
		s = fmt.Sprintf("%s\t%d\t%s\t%s\n", curProvider, ch1.tvType, ch1.group, ch1.title)
		writer.WriteString(s)
	}
	writer.Flush()
	file.Close()
}

func loadConfig() {
	path := filepath.Join(getConfigPath(), settingsFile)
	if _, err := os.Stat(path); err == nil {
		if file, err := os.Open(path); err == nil {
			scanner := bufio.NewScanner(file)
			scanner.Scan()
			line := scanner.Text()
			fmt.Sscanf(line, "%d\t%d\t%d\t%d", &config.wndX, &config.wndY, &config.wndHt, &config.wndWd)
			for scanner.Scan() {
				line = scanner.Text()
				strs := strings.Split(line, "\t")
				prov := strs[0]
				if prov != curProvider {
					continue
				}
				typ, _ := strconv.Atoi(strs[1])
				g := strs[2]
				t := strs[3]
				gr := tv[typ][g]
				ch1 := gr.channels[t]
				if ch1 == nil {
					continue
				}
				config.favorites = append(config.favorites, ch1)
			}
			file.Close()
		}
	}
}

func sizeChange(wnd *gtk.Window, event *gdk.Event) bool {
	config.wndWd, config.wndHt = wnd.GetSize()
	config.wndX, config.wndY = wnd.GetPosition()
	return false
}

func appExit() {
	if mpvCmd != nil {
		mpvCmd.Process.Kill()
	}
	saveConfig()
}
