package serverfull

import (
	"github.com/asecurityteam/runhttp"
	"github.com/asecurityteam/settings"
)

// HelpStatic generates the help output for static builds.
func HelpStatic() string {
	grp, _ := settings.GroupFromComponent(&runhttp.Component{})
	return settings.ExampleEnvGroups([]settings.Group{&settings.SettingGroup{
		NameValue:   "SERVERFULL",
		GroupValues: []settings.Group{grp},
	}})
}
