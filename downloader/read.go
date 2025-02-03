package downloader

import (
	"fmt"
	"github.com/belphemur/mangal/color"
	"github.com/belphemur/mangal/constant"
	"github.com/belphemur/mangal/history"
	"github.com/belphemur/mangal/key"
	"github.com/belphemur/mangal/log"
	"github.com/belphemur/mangal/open"
	"github.com/belphemur/mangal/packer"
	"github.com/belphemur/mangal/source"
	"github.com/belphemur/mangal/style"
	"github.com/spf13/viper"
)

// Read the chapter by downloading it with the given source
// and opening it with the configured reader.
func Read(chapter *source.Chapter, progress func(string)) error {
	if viper.GetBool(key.ReaderReadInBrowser) {
		return open.StartWith(
			chapter.URL,
			viper.GetString(key.ReaderBrowser),
		)
	}

	if viper.GetBool(key.DownloaderReadDownloaded) && chapter.IsDownloaded() {
		path, err := chapter.Path(false)
		if err == nil {
			return openRead(path, chapter, progress)
		}
	}

	log.Infof("downloading %s for reading. Provider is %s", chapter.Name, chapter.Source().ID())
	log.Infof("getting pages of %s", chapter.Name)
	progress("Getting pages")
	err := chapter.Source().LoadPagesOf(chapter)
	if err != nil {
		log.Error(err)
		return err
	}

	err = chapter.DownloadPages(true, progress)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Info("getting " + viper.GetString(key.FormatsUse) + " converter")
	conv, err := packer.Get(viper.GetString(key.FormatsUse))
	if err != nil {
		log.Error(err)
		return err
	}

	log.Info("converting " + viper.GetString(key.FormatsUse))
	progress(fmt.Sprintf(
		"Converting %d pages to %s %s",
		len(chapter.Pages),
		style.Fg(color.Yellow)(viper.GetString(key.FormatsUse)),
		style.Faint(chapter.SizeHuman())),
	)
	path, err := conv.SaveTemp(chapter)
	if err != nil {
		log.Error(err)
		return err
	}

	err = openRead(path, chapter, progress)
	if err != nil {
		log.Error(err)
		return err
	}

	progress("Done")
	return nil
}

func openRead(path string, chapter *source.Chapter, progress func(string)) error {
	if viper.GetBool(key.HistorySaveOnRead) {
		go func() {
			err := history.Save(chapter)
			if err != nil {
				log.Warn(err)
			} else {
				log.Info("history saved")
			}
		}()
	}

	var (
		reader string
		err    error
	)

	switch viper.GetString(key.FormatsUse) {
	case constant.FormatPDF:
		reader = viper.GetString(key.ReaderPDF)
	case constant.FormatCBZ:
		reader = viper.GetString(key.ReaderCBZ)
	case constant.FormatZIP:
		reader = viper.GetString(key.ReaderZIP)
	case constant.FormatPlain:
		reader = viper.GetString(key.RaderPlain)
	}

	if reader != "" {
		log.Info("opening with " + reader)
		progress(fmt.Sprintf("Opening %s", reader))
	} else {
		log.Info("no reader specified. opening with default")
		progress("Opening")
	}

	err = open.RunWith(path, reader)
	if err != nil {
		log.Error(err)
		return fmt.Errorf("could not open %s with %s: %s", path, reader, err.Error())
	}

	log.Info("opened without errors")

	return nil
}
