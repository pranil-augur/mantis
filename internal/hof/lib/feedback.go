package lib

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/opentofu/opentofu/internal/hof/lib/yagu"
	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/cmd/hof/verinfo"
)

var info = `
Add more information here

---

Hof Metadata:

<pre>
Version:     v%s
Commit:      %s

BuildDate:   %s
GoVersion:   %s
OS / Arch:   %s %s
</pre>
`

func SendFeedback(args []string, rflags flags.RootPflagpole, cflags flags.FeedbackPflagpole) error {
	title := url.QueryEscape(strings.Join(args, " "))

	body := fmt.Sprintf(
		info,
		verinfo.Version,
		verinfo.Commit,
		verinfo.BuildDate,
		verinfo.GoVersion,
		verinfo.BuildOS,
		verinfo.BuildArch,
	)
	body = url.QueryEscape(body)

	labels := cflags.Labels
	what := "discussions"
	catg := "category=general&"
	if cflags.Issue {
		what = "issues"
		catg = ""
	}
	
	url := fmt.Sprintf("https://github.com/opentofu/opentofu/internal/hof/%s/new?%slabels=%s&title=%s&body=%s", what, catg, labels, title, body)
	yagu.OpenBrowserCmdSafe(url)

	return nil
}
