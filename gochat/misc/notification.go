package misc

import (
	"runtime"

	"github.com/andybrewer/mack"
)

func Notify(message string, args ...string) error {
	if runtime.GOOS == "darwin" {
		var title, subtitle, sound string
		if len(args) > 0 && args[0] != "" {
			title = args[0]
		}
		if len(args) > 1 && args[1] != "" {
			subtitle = args[1]
		}
		if len(args) > 2 && args[2] != "" {
			sound = args[2]
		}
		
		
		return mack.Notify(message, title, subtitle, sound)
	}
	//TODO: implement different OS notifications
	return nil
}