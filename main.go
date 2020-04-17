package main

import (
	"flag"
	"os"
	"sort"
	"time"

	"github.com/crazyfacka/remoteclean/handler"
	"github.com/crazyfacka/remoteclean/modules"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func main() {
	var dirs []string

	dryrun := flag.Bool("dry", false, "doesn't remove any file")
	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	log.Info().Msg("Starting remoteclean")

	viper.SetConfigName(".remoteclean")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Msg("Error reading config file")
		os.Exit(-1)
	}

	log.Debug().Interface(".remoteclean", viper.AllSettings()).Msg("Loaded configuration")

	remote, err := modules.GetSSHConn(viper.GetStringMap("remote"))
	if err != nil {
		log.Error().Err(err).Msg("Unable to setup remote session")
		return
	}

	mount := viper.GetString("mount")

	space, err := handler.GetFreeSpace(remote, mount)
	if err != nil {
		log.Error().Err(err).Msg("Unable to check free space")
		return
	}

	log.Info().Float64("amount", space).Msg("Free space in GB")
	threshold := viper.GetFloat64("space_threshold")
	if space > threshold {
		log.Info().Float64("threshold", threshold).Msg("There is enough free space on remote")
		os.Exit(0)
	}

	for _, dir := range viper.GetStringMap("remote")["dirs"].([]interface{}) {
		dirs = append(dirs, dir.(string))
	}

	items, err := handler.GetContents(remote, dirs)
	if err != nil {
		log.Error().Err(err).Msg("Unable to get contents")
		return
	}

	sort.Sort(items)
	handler.DeleteUntil(remote, items, space, threshold*1.1, *dryrun)

	if !*dryrun {
		err = handler.RefreshLibrary(viper.GetStringMap("remote")["host"].(string))
	}

	log.Info().Msg("Done")
}
