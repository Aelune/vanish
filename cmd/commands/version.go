package command

import (
	"fmt"
)

// Version information
const (
	VERSION = "0.9.2"
	// NAME BUILD_DATE = "2024-01-15"
	NAME = "ares-cinnamon"
	//AUTHOR = "Aelune"
)

// func PrintVersion() {
// 	fmt.Printf("vx %s (commit: %s, built: %s)\n", Version)
// }

// ShowVersion displays version information of vanish with alias name too
func ShowVersion() {
	// fmt.Printf("Vanish (vx) - Safe File Removal Tool\n")
	fmt.Printf("%s %s\n", VERSION, NAME)
	// fmt.Printf("Build Date: %s\n", BUILD_DATE)
	// fmt.Printf("Build Date: %s\n", ALIAS)
	// fmt.Printf("Author: %s\n", AUTHOR)
	// fmt.Printf("Go Version: %s\n", runtime.Version())
	// fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	// Show config path
	// homeDir, _ := os.UserHomeDir()
	// configPath := filepath.Join(homeDir, ".config", "vanish-test", "vanish.toml")
	// fmt.Printf("Config Path: %s\n", configPath)

	// Show cache path
	// config, err := loadConfig()
	// if err == nil {
	// 	fmt.Printf("Cache Path: %s\n", expandPath(config.Cache.Directory))
	// }
}

// Possible future Names, added so that i wont forget proper name

// 0.1   zeus-honey ,
// v0.2: thor-caramel ,
// v0.3: athena-vanilla ,
// v0.4: apollo-chocolate ,
// v0.5: artemis-strawberry ,
// v0.6: hades-licorice ,
// v0.7: poseidon-blueberry ,
// v0.8: hermes-mint ,
// v0.9: ares-cinnamon ,
// v1.0: hera-raspberry ,
// v1.1: demeter-maple ,
// v1.2: dionysus-grape ,
// v1.3: hephaestus-toffee ,
// v1.4: aphrodite-rose ,
// v1.5: persephone-cherry ,
// v2.0: odin-butterscotch ,
// v2.1: freya-lavender ,
// v2.2: loki-peppermint ,
// v2.3: baldur-coconut ,
// v2.4: tyr-almond ,
// v3.0: ra-mango ,
// v3.1: isis-orange ,
// v3.2: anubis-coffee ,
// v3.3: thoth-pistachio ,
// v4.0: brahma-cardamom ,
// v4.1: vishnu-saffron ,
// v4.2: shiva-ginger ,
// v5.0: jupiter-tiramisu ,
// v5.1: venus-bubblegum ,
// v5.2: mars-chili ,
// v6.0: prometheus-ambrosia ,
