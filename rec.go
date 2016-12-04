package radigo

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/mitchellh/cli"
	"github.com/yyoshiki41/go-radiko"
)

type recCommand struct {
	ui cli.Ui
}

func (c *recCommand) Run(args []string) int {
	var stationID, start, areaID string
	var flagForce bool

	f := flag.NewFlagSet("rec", flag.ContinueOnError)
	f.StringVar(&stationID, "id", "", "id")
	f.StringVar(&start, "start", "", "start")
	f.StringVar(&start, "s", "", "start")
	f.StringVar(&areaID, "area", "", "area")
	f.StringVar(&areaID, "a", "", "area")
	f.BoolVar(&flagForce, "force", false, "force")
	f.BoolVar(&flagForce, "f", false, "force")
	f.Usage = func() { c.ui.Error(c.Help()) }
	if err := f.Parse(args); err != nil {
		return 1
	}

	if stationID == "" {
		c.ui.Error("StationID is empty.")
		return 1
	}

	startTime, err := time.ParseInLocation(datetimeLayout, start, location)
	if err != nil {
		c.ui.Error(fmt.Sprintf(
			"invalid start time format: %s", start))
		return 1
	}

	err = downloadSwfPlayer(flagForce)
	if err != nil {
		c.ui.Error(fmt.Sprintf(
			"Failed to download player.swf: %s", err))
		return 1
	}

	pngPath, err := extractPngFile()
	if err != nil {
		c.ui.Error(fmt.Sprintf(
			"Failed to execute swfextract: %s", err))
		return 1
	}

	client, err := getClient("", areaID)
	if err != nil {
		c.ui.Error(fmt.Sprintf(
			"Failed to construct a radiko Client: %s", err))
		return 1
	}
	_, err = client.AuthorizeToken(context.Background(), pngPath)
	if err != nil {
		c.ui.Error(fmt.Sprintf(
			"Failed to get auth_token: %s", err))
		return 1
	}

	uri, err := client.TimeshiftPlaylistM3U8(context.Background(), stationID, startTime)
	if err != nil {
		c.ui.Error(fmt.Sprintf(
			"Failed to get playlist.m3u8: %s", err))
		return 1
	}

	chunklist, err := radiko.GetChunklistFromM3U8(uri)
	if err != nil {
		c.ui.Error(fmt.Sprintf(
			"Failed to get chunklist: %s", err))
		return 1
	}

	aacDir, err := ioutil.TempDir(radigoPath, "aac")
	if err != nil {
		c.ui.Error(fmt.Sprintf(
			"Failed to create temp dir: %s", err))
		return 1
	}
	defer os.RemoveAll(aacDir)

	err = bulkDownload(aacDir, chunklist)
	if err != nil {
		c.ui.Error(fmt.Sprintf(
			"Failed to download aac files: %s", err))
		return 1
	}

	err = createConcatedAACFile(aacDir)
	if err != nil {
		c.ui.Error(fmt.Sprintf(
			"Failed to create concat aac file: %s", err))
		return 1
	}

	err = convertAACToMP3()
	if err != nil {
		c.ui.Error(fmt.Sprintf(
			"Failed to convert aac to mp3: %s", err))
		return 1
	}

	return 0
}

func (c *recCommand) Synopsis() string {
	return "Record a radiko program"
}

func (c *recCommand) Help() string {
	return strings.TrimSpace(`
Usage: radigo rec [options]
  Record a radiko program.
Options:
  -id=name                 Station id
  -start,s=201610101000    Start time
  -area,a=name             Area id
  -force,-f                Ignore cache and force refresh
`)
}
