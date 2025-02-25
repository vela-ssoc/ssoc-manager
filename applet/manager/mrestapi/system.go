package mrestapi

import (
	"bytes"
	"html/template"
	"sync"

	"github.com/vela-ssoc/vela-manager/banner"
	"github.com/xgfone/ship/v5"
)

func NewSystem() *System {
	sys := new(System)
	sys.onceBannerHTML = sync.OnceValues(sys.parseBannerHTML)
	return sys
}

type System struct {
	onceBannerHTML func() (*template.Template, error)
}

func (sys *System) Route(r *ship.RouteGroupBuilder) error {
	r.Route("/system/banner").GET(sys.banner)
	return nil
}

func (sys *System) banner(c *ship.Context) error {
	if banner.SupportedANSI(c.UserAgent()) {
		c.SetContentType(ship.MIMETextPlainCharsetUTF8)
		banner.ANSI(c)
		return nil
	}

	tpl, err := sys.onceBannerHTML()
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	banner.ANSI(buf)

	return tpl.Execute(c, buf)
}

func (sys *System) parseBannerHTML() (*template.Template, error) {
	const bannerHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta http-equiv="X-UA-Compatible" content="IE=edge" />
	<meta name="color-scheme" content="light dark">
    <title>BANNER</title>
</head>
<body>
<pre id="app" style="font-family: monospace, serif; font-size: smaller;"></pre>

<script type="module">
    import {AnsiUp} from 'https://cdn.jsdelivr.net/npm/ansi_up/ansi_up.min.js'

    var txt = '{{ . }}'
    var ansiup = new AnsiUp();
    var html = ansiup.ansi_to_html(txt);
    document.getElementById("app").innerHTML = html;
</script>
</body>
</html>
`
	return template.New("banner").Parse(bannerHTML)
}
