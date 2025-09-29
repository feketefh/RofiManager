package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

type RofiManager struct {
	baseDir    string
	themesDir  string
	scriptsDir string
	configPath string
	config     *ini.File
	allModes   []string
}

func NewRofiManager() *RofiManager {
	home, _ := os.UserHomeDir()
	baseDir := filepath.Join(home, ".config", "rofi-manager")
	themesDir := filepath.Join(baseDir, "themes")
	scriptsDir := filepath.Join(baseDir, "scripts")
	configPath := filepath.Join(baseDir, "config.conf")
	manager := &RofiManager{
		baseDir:    baseDir,
		themesDir:  themesDir,
		scriptsDir: scriptsDir,
		configPath: configPath,
		allModes:   []string{"run", "drun", "window", "ssh", "filebrowser", "key"},
	}
	manager.ensureConfigDirs()
	manager.loadConfig()
	return manager
}

func (rm *RofiManager) ensureConfigDirs() {
	os.MkdirAll(rm.themesDir, 0755)
	os.MkdirAll(rm.scriptsDir, 0755)
	if _, err := os.Stat(rm.configPath); os.IsNotExist(err) {
		cfg := ini.Empty()
		cfg.Section("modes").Key("enabled").SetValue("run,drun,window")
		cfg.Section("scripts").Key("enabled").SetValue("")
		cfg.Section("theme").Key("enabled").SetValue("")
		cfg.SaveTo(rm.configPath)
	}
}

func (rm *RofiManager) loadConfig() {
	cfg, err := ini.Load(rm.configPath)
	if err != nil {
		cfg = ini.Empty()
		cfg.Section("modes").Key("enabled").SetValue("run,drun,window")
		cfg.Section("scripts").Key("enabled").SetValue("")
		cfg.Section("theme").Key("enabled").SetValue("")
		cfg.SaveTo(rm.configPath)
	}
	rm.config = cfg
}

func (rm *RofiManager) saveConfig() {
	rm.config.SaveTo(rm.configPath)
}

func (rm *RofiManager) getEnabledModes() []string {
	val := rm.config.Section("modes").Key("enabled").String()
	modes := []string{}
	for _, m := range strings.Split(val, ",") {
		m = strings.TrimSpace(m)
		if m != "" {
			modes = append(modes, m)
		}
	}
	return modes
}

func (rm *RofiManager) setEnabledModes(modes []string) {
	rm.config.Section("modes").Key("enabled").SetValue(strings.Join(modes, ","))
	rm.saveConfig()
}

func (rm *RofiManager) getEnabledScripts() []string {
	val := rm.config.Section("scripts").Key("enabled").String()
	scripts := []string{}
	for _, s := range strings.Split(val, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			scripts = append(scripts, s)
		}
	}
	return scripts
}

func (rm *RofiManager) setEnabledScripts(scripts []string) {
	rm.config.Section("scripts").Key("enabled").SetValue(strings.Join(scripts, ","))
	rm.saveConfig()
}

func (rm *RofiManager) getEnabledTheme() string {
	return rm.config.Section("theme").Key("enabled").String()
}

func (rm *RofiManager) setEnabledTheme(theme string) {
	rm.config.Section("theme").Key("enabled").SetValue(theme)
	rm.saveConfig()
}

func (rm *RofiManager) loadScripts() []string {
	files, _ := ioutil.ReadDir(rm.scriptsDir)
	scripts := []string{}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".sh") {
			scripts = append(scripts, f.Name())
		}
	}
	return scripts
}

func (rm *RofiManager) loadThemes() []string {
	files, _ := ioutil.ReadDir(rm.themesDir)
	themes := []string{}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".rasi") {
			themes = append(themes, f.Name())
		}
	}
	return themes
}

func (rm *RofiManager) rofiMenu(prompt string, options []string, theme string) string {
	rofiCmd := []string{"-dmenu", "-p", prompt}
	if theme != "" {
		themePath := filepath.Join(rm.themesDir, theme)
		if _, err := os.Stat(themePath); err == nil {
			rofiCmd = append(rofiCmd, "-theme", themePath)
		}
	}
	cmd := exec.Command("rofi", rofiCmd...)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	io.WriteString(stdin, strings.Join(options, "\n"))
	stdin.Close()
	out, _ := ioutil.ReadAll(stdout)
	cmd.Wait()
	return strings.TrimSpace(string(out))
}

func (rm *RofiManager) showInfo(message string) {
	rm.rofiMenu(message, []string{"OK"}, rm.getEnabledTheme())
}

func (rm *RofiManager) addScript() {
	path := rm.rofiMenu("Path to script (.sh)", []string{""}, rm.getEnabledTheme())
	if path == "" {
		return
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		rm.showInfo("File does not exist.")
		return
	}
	if !strings.HasSuffix(path, ".sh") {
		rm.showInfo("File must end with .sh")
		return
	}
	dest := filepath.Join(rm.scriptsDir, filepath.Base(path))
	if _, err := os.Stat(dest); err == nil {
		overwrite := rm.rofiMenu(fmt.Sprintf("Overwrite %s?", filepath.Base(path)), []string{"No", "Yes"}, rm.getEnabledTheme())
		if overwrite != "Yes" {
			return
		}
	}
	input, err := ioutil.ReadFile(path)
	if err != nil {
		rm.showInfo(fmt.Sprintf("Error copying file:\n%s", err))
		return
	}
	err = ioutil.WriteFile(dest, input, 0755)
	if err != nil {
		rm.showInfo(fmt.Sprintf("Error copying file:\n%s", err))
	}
}

func (rm *RofiManager) addTheme() {
	path := rm.rofiMenu("Path to theme (.rasi)", []string{""}, rm.getEnabledTheme())
	if path == "" {
		return
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		rm.showInfo("File does not exist.")
		return
	}
	if !strings.HasSuffix(path, ".rasi") {
		rm.showInfo("File must end with .rasi")
		return
	}
	dest := filepath.Join(rm.themesDir, filepath.Base(path))
	if _, err := os.Stat(dest); err == nil {
		overwrite := rm.rofiMenu(fmt.Sprintf("Overwrite %s?", filepath.Base(path)), []string{"No", "Yes"}, rm.getEnabledTheme())
		if overwrite != "Yes" {
			return
		}
	}
	input, err := ioutil.ReadFile(path)
	if err != nil {
		rm.showInfo(fmt.Sprintf("Error copying file:\n%s", err))
		return
	}
	err = ioutil.WriteFile(dest, input, 0644)
	if err != nil {
		rm.showInfo(fmt.Sprintf("Error copying file:\n%s", err))
	}
}

func (rm *RofiManager) enableScript() {
	scripts := rm.loadScripts()
	if len(scripts) == 0 {
		rm.showInfo("No scripts found.")
		return
	}
	enabled := rm.getEnabledScripts()
	enabledSet := make(map[string]bool)
	for _, s := range enabled {
		enabledSet[s] = true
	}
	for {
		options := []string{}
		for _, script := range scripts {
			mark := "[ ]"
			if enabledSet[script] {
				mark = "[x]"
			}
			options = append(options, fmt.Sprintf("%s %s", mark, script))
		}
		choice := rm.rofiMenu("Toggle scripts (Enter to finish)", options, rm.getEnabledTheme())
		if choice == "" {
			break
		}
		script := strings.TrimSpace(choice[4:])
		if enabledSet[script] {
			delete(enabledSet, script)
		} else {
			enabledSet[script] = true
		}
		newEnabled := []string{}
		for s := range enabledSet {
			newEnabled = append(newEnabled, s)
		}
		rm.setEnabledScripts(newEnabled)
	}
}

func (rm *RofiManager) enableTheme() {
	themes := rm.loadThemes()
	if len(themes) == 0 {
		rm.showInfo("No themes found.")
		return
	}
	enabledTheme := rm.getEnabledTheme()
	for {
		options := []string{}
		for _, theme := range themes {
			mark := "[ ]"
			if theme == enabledTheme {
				mark = "[x]"
			}
			options = append(options, fmt.Sprintf("%s %s", mark, theme))
		}
		choice := rm.rofiMenu("Select theme (Enter to escape)", options, enabledTheme)
		if choice == "" {
			break
		}
		theme := strings.TrimSpace(choice[4:])
		if theme == enabledTheme {
			enabledTheme = ""
		} else {
			enabledTheme = theme
		}
		rm.setEnabledTheme(enabledTheme)
		enabledTheme = rm.getEnabledTheme()
	}
}

func (rm *RofiManager) toggleModes() {
	enabled := rm.getEnabledModes()
	enabledSet := make(map[string]bool)
	for _, m := range enabled {
		enabledSet[m] = true
	}
	for {
		options := []string{}
		for _, mode := range rm.allModes {
			mark := "[ ]"
			if enabledSet[mode] {
				mark = "[x]"
			}
			options = append(options, fmt.Sprintf("%s %s", mark, mode))
		}
		choice := rm.rofiMenu("Toggle modes (Enter to finish)", options, rm.getEnabledTheme())
		if choice == "" {
			break
		}
		mode := strings.TrimSpace(choice[4:])
		if enabledSet[mode] {
			delete(enabledSet, mode)
		} else {
			enabledSet[mode] = true
		}
		newEnabled := []string{}
		for m := range enabledSet {
			newEnabled = append(newEnabled, m)
		}
		rm.setEnabledModes(newEnabled)
	}
}

func (rm *RofiManager) selectMode() {
	enabledModes := rm.getEnabledModes()
	if len(enabledModes) == 0 {
		rm.showInfo("No modes enabled.")
		return
	}
	choice := rm.rofiMenu("Select mode", enabledModes, rm.getEnabledTheme())
	for _, mode := range enabledModes {
		if choice == mode {
			theme := rm.getEnabledTheme()
			rofiCmd := []string{"-show", mode}
			if theme != "" {
				themePath := filepath.Join(rm.themesDir, theme)
				if _, err := os.Stat(themePath); err == nil {
					rofiCmd = append(rofiCmd, "-theme", themePath)
				}
			}
			cmd := exec.Command("rofi", rofiCmd...)
			cmd.Run()
			os.Exit(0)
		}
	}
}

func (rm *RofiManager) run() {
	for {
		mainOptions := []string{
			"Select Mode",
			"Enable/Disable Modes",
			"Enable Script",
			"Enable Theme",
			"Add Script",
			"Add Theme",
			"Exit",
		}
		choice := rm.rofiMenu("Rofi Manager", mainOptions, rm.getEnabledTheme())
		switch choice {
		case "Select Mode":
			rm.selectMode()
		case "Enable/Disable Modes":
			rm.toggleModes()
		case "Enable Script":
			rm.enableScript()
		case "Enable Theme":
			rm.enableTheme()
		case "Add Script":
			rm.addScript()
		case "Add Theme":
			rm.addTheme()
		case "Exit", "":
			return
		}
	}
}

func main() {
	manager := NewRofiManager()
	manager.run()
}
