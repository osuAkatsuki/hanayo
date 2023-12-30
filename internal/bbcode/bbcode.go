package bbcode

import (
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/osuAkatsuki/hanayo/app/states/settings"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// Google implementation of clamp function
func Clamp(x int, min int, max int) int {
	switch {
	case x < min:
		return min
	case x > max:
		return max
	default:
		return x
	}
}

func ClampFloat(x float64, min float64, max float64) float64 {
	switch {
	case x < min:
		return min
	case x > max:
		return max
	default:
		return x
	}
}

// Reference: https://gist.github.com/elliotchance/d419395aa776d632d897
func ReplaceAllStringSubmatchFunc(re *regexp.Regexp, str string, repl func([]string, string) string) string {
	result := ""
	lastIndex := 0

	for _, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			if v[i] == -1 {
				groups = append(groups, "")
				continue
			}
			groups = append(groups, str[v[i]:v[i+1]])
		}

		identifier := RandStringBytes(6)
		result += str[lastIndex:v[0]] + repl(groups, identifier)
		lastIndex = v[1]
	}

	return result + str[lastIndex:]
}

func parseBold(text string) string {
	text = strings.Replace(text, "[b]", "<strong>", -1)
	text = strings.Replace(text, "[/b]", "</strong>", -1)

	text = strings.Replace(text, "[bold]", "<strong>", -1)
	text = strings.Replace(text, "[/bold]", "</strong>", -1)
	return text
}

func parseCentre(text string) string {
	text = strings.Replace(text, "[centre]", "<center>", -1)
	text = strings.Replace(text, "[/centre]", "</center>", -1)

	text = strings.Replace(text, "[center]", "<center>", -1)
	text = strings.Replace(text, "[/center]", "</center>", -1)
	return text
}

func parseHeading(text string) string {
	text = strings.Replace(text, "[heading]", "<h2>", -1)

	regex := regexp.MustCompile(`\[\/heading\]\n?`)
	text = regex.ReplaceAllString(text, "</h2>")
	return text
}

func parseItalic(text string) string {
	text = strings.Replace(text, "[i]", "<em>", -1)
	text = strings.Replace(text, "[/i]", "</em>", -1)

	text = strings.Replace(text, "[italic]", "<em>", -1)
	text = strings.Replace(text, "[/italic]", "</em>", -1)
	return text
}

func parseStrike(text string) string {
	text = strings.Replace(text, "[s]", "<strike>", -1)
	text = strings.Replace(text, "[/s]", "</strike>", -1)

	text = strings.Replace(text, "[strike]", "<strike>", -1)
	text = strings.Replace(text, "[/strike]", "</strike>", -1)

	return text
}

func parseUnderline(text string) string {
	text = strings.Replace(text, "[u]", "<u>", -1)
	text = strings.Replace(text, "[/u]", "</u>", -1)

	text = strings.Replace(text, "[underline]", "<u>", -1)
	text = strings.Replace(text, "[/underline]", "</u>", -1)
	return text
}

func parseSpoiler(text string) string {
	text = strings.Replace(text, "[spoiler]", "<span class='bbcode-spoiler'>", -1)
	text = strings.Replace(text, "[/spoiler]", "</span>", -1)
	return text
}

func parseNotice(text string) string {
	regex := regexp.MustCompile(`(?s)\[notice\]\n?(.*?)\n?\[\/notice\]\n?`)
	text = regex.ReplaceAllString(text, "<div class='bbcode-notice'>$1</div>")

	return text
}

func parseColour(text string) string { // support both colour and color
	regex := regexp.MustCompile(`\[(color|colour)=([^]:]+)\]`)
	text = regex.ReplaceAllString(text, "<span style='color: $2'>")

	regex2 := regexp.MustCompile(`\[\/(color|colour)\]`)
	text = regex2.ReplaceAllString(text, "</span>")
	return text
}

func parseAudio(text string) string {
	regex := regexp.MustCompile(`\[audio\]([^[]+)\[\/audio\]\n?`)

	text = regex.ReplaceAllString(text, "<audio controls='controls' preload='none' src='$1'></audio>")
	return text
}

func parseUrl(text string) string {
	regex := regexp.MustCompile(`\[url\](.+?)\[\/url\]`)
	text = regex.ReplaceAllString(text, "<a rel='nofollow' href='$1'>$1</a>")

	regex2 := regexp.MustCompile(`\[url=([^\]]+)\]`)
	text = regex2.ReplaceAllString(text, "<a rel='nofollow' href='$1'>")
	text = strings.Replace(text, "[/url]", "</a>", -1)

	return text
}

func parseQuote(text string) string {
	regex := regexp.MustCompile(`\[quote=\"([^:]+)\"\]\s*`)
	text = regex.ReplaceAllString(text, "<blockquote class='bbcode-blockquote'><h4>$1 wrote:</h4>")

	regex2 := regexp.MustCompile(`\[quote\]\s*`)
	text = regex2.ReplaceAllString(text, "<blockquote class='bbcode-blockquote'>")

	regex3 := regexp.MustCompile(`\s*\[\/quote\]\n?`)
	text = regex3.ReplaceAllString(text, "</blockquote>")

	return text
}

func parseSize(text string) string {
	regex := regexp.MustCompile(`\[size=(\d+)\]`)
	text = ReplaceAllStringSubmatchFunc(regex, text, func(groups []string, _ string) string {
		size, _ := strconv.Atoi(groups[1])
		size = Clamp(size, 30, 200)

		return fmt.Sprintf("<span style='font-size: %d%%'>", size)
	})

	text = strings.Replace(text, "[/size]", "</span>", -1)
	return text
}

func parseEmail(text string) string {
	regex := regexp.MustCompile(`\[email\](([^[]+)@([^[]+))\[\/email\]`)
	text = regex.ReplaceAllString(text, "<a rel='nofollow' href='mailto:$1'>$1</a>")

	regex2 := regexp.MustCompile(`\[email=(([^[]+)@([^[]+))\]`)
	text = regex2.ReplaceAllString(text, "<a rel='nofollow' href='mailto:$1'>")
	text = strings.Replace(text, "[/email]", "</a>", -1)

	return text
}

// TODO: this needs to be redone into profile cards.
func parseProfile(text string) string {

	regex := regexp.MustCompile(`\[profile(?:=([0-9]+))?\](.*?)\[\/profile\]`)
	text = ReplaceAllStringSubmatchFunc(regex, text, func(groups []string, _ string) string {
		if groups[1] != "" {
			return fmt.Sprintf("<a href='/u/%s'>%s</a>", groups[1], groups[2])
		}

		return fmt.Sprintf("<a href='/u/%s'>/u/%s</a>", groups[2], groups[2])
	})

	return text
}

func parseImage(text string) string {
	regex := regexp.MustCompile(`\[img\]([^[]+)\[\/img\]`)
	text = ReplaceAllStringSubmatchFunc(regex, text, func(groups []string, _ string) string {
		decoded, _ := url.QueryUnescape(groups[1])
		return fmt.Sprintf("<img src='%s' loading='lazy'/>", decoded)
	})

	// there is also a case of our old bbcode parsing [img=url][/img]
	regex2 := regexp.MustCompile(`\[img=([^[]+)\]\[\/img\]`)
	text = ReplaceAllStringSubmatchFunc(regex2, text, func(groups []string, _ string) string {
		decoded, _ := url.QueryUnescape(groups[1])
		return fmt.Sprintf("<img src='%s' loading='lazy'/>", decoded)
	})

	return text
}

func parseList(text string) string {
	regex := regexp.MustCompile(`\[list=[^]]+\]\s*\[\*\]`) // numered list.
	text = regex.ReplaceAllString(text, "<ol><li>")

	regex2 := regexp.MustCompile(`\[list\]\s*\[\*\]`) // bullet list.
	text = regex2.ReplaceAllString(text, "<ol style='list-style-type: disc;'><li>")

	regex3 := regexp.MustCompile(`\[\/\*(:m)?\]\n?\n?`)
	text = regex3.ReplaceAllString(text, "</li>")

	regex4 := regexp.MustCompile(`\s*\[\*\]`)
	text = regex4.ReplaceAllString(text, "<li>")

	regex5 := regexp.MustCompile(`\s*\[\/list\]\n?\n?`)
	text = regex5.ReplaceAllString(text, "</ol>")

	regex6 := regexp.MustCompile(`\[list=[^]]+\](.+?)(<li>|</ol>)`)
	text = regex6.ReplaceAllString(text, "<ul class='bbcode-list-title'><li>$1</li></ul><ol>$2")

	regex7 := regexp.MustCompile(`\[list\](.+?)(<li>|</ol>)`)
	text = regex7.ReplaceAllString(text, "<ul class='bbcode-list-title'><li>$1</li></ul><ol style='list-style-type: disc;'>$2")

	return text
}

func parseImagemap(text string) string {
	regex := regexp.MustCompile(`(?s)\[imagemap\]\s+(?P<image_url>.+?)(?P<lines>(?:\s+.+?)\s+)+\[\/imagemap\]\n?`)

	text = ReplaceAllStringSubmatchFunc(regex, text, func(matches []string, _ string) string {
		if matches == nil {
			return ""
		}

		pseudoHtml := fmt.Sprintf(
			"<div class='bbcode-imagemap'><img src='%s' class='bbcode-imagemap-image' loading='lazy'>",
			matches[regex.SubexpIndex("image_url")],
		)

		lineRegex := regexp.MustCompile(`(?m)^\s*(?P<x>\S+)\s+(?P<y>\S+)\s+(?P<width>\S+)\s+(?P<height>\S+)\s+(?P<redirect>\S+)\s+(?P<title>.+?)\s*$`)
		lines := lineRegex.FindAllStringSubmatch(matches[regex.SubexpIndex("lines")], -1)

		for _, line := range lines {

			redirect := line[lineRegex.SubexpIndex("redirect")]
			xAxis, err := strconv.ParseFloat(line[lineRegex.SubexpIndex("x")], 64)
			if err != nil {
				xAxis = 0.0
			}
			yAxis, err := strconv.ParseFloat(line[lineRegex.SubexpIndex("y")], 64)
			if err != nil {
				yAxis = 0.0
			}
			width, err := strconv.ParseFloat(line[lineRegex.SubexpIndex("width")], 64)
			if err != nil {
				width = 0.0
			}
			height, err := strconv.ParseFloat(line[lineRegex.SubexpIndex("height")], 64)
			if err != nil {
				height = 0.0
			}
			title := line[lineRegex.SubexpIndex("title")]

			tag := "a"
			if redirect == "#" {
				tag = "span"
			}

			xAxisClamped := ClampFloat(xAxis, float64(0.0), float64(100.0))
			yAxisClamped := ClampFloat(yAxis, float64(0.0), float64(100.0))
			widthClamped := ClampFloat(width, float64(0.0), float64(100.0))
			heightClamped := ClampFloat(height, float64(0.0), float64(100.0))

			tooltipPos := "top center"
			if yAxisClamped < 13.0 {
				tooltipPos = "bottom center"
			}

			pseudoHtml += fmt.Sprintf(
				"<%s class='bbcode-imagemap-tooltip' href='%s' style='left: %f%%; top: %f%%; width: %f%%; height: %f%%;' data-tooltip='%s' data-position='%s'></%s>",
				tag,
				redirect,
				xAxisClamped,
				yAxisClamped,
				widthClamped,
				heightClamped,
				title,
				tooltipPos,
				tag,
			)
		}

		pseudoHtml += "</div>"
		pseudoHtml = strings.Replace(pseudoHtml, "\n", "", -1)

		return pseudoHtml
	})

	return text
}

func parseBox(text string) string {
	regex := regexp.MustCompile(`\[box=((\\\[\[\]]|[^][]|\[(\\\[\[\]]|[^][]|(.*?))*\])*?)\]\n*`)

	text = ReplaceAllStringSubmatchFunc(regex, text, func(groups []string, identifier string) string {

		return fmt.Sprintf(
			"<div class='bbcode-box'><button class='bbcode-box-btn' id='btn-%s' type='button' onclick='toggleBBCodeBox(this)'><i id='icon-%s' class='bbcode-box-icon fa-solid fa-angle-right'></i><span>%s</span></button><div class='bbcode-box-content bbcode-hidden' id='content-%s'>",
			identifier,
			identifier,
			groups[1],
			identifier,
		)
	})

	regex2 := regexp.MustCompile(`\n*\[\/box\]\n?`)
	text = regex2.ReplaceAllString(text, "</div></div>")

	regex3 := regexp.MustCompile(`\[spoilerbox\]\n*`)
	text = ReplaceAllStringSubmatchFunc(regex3, text, func(_ []string, identifier string) string {
		return fmt.Sprintf(
			"<div class='bbcode-box'><button class='bbcode-box-btn' id='btn-%s' type='button' onclick='toggleBBCodeBox(this)'><i id='icon-%s' class='bbcode-box-icon fa-solid fa-angle-right'></i><span>SPOILER</span></button><div class='bbcode-box-content bbcode-hidden' id='content-%s'>",
			identifier,
			identifier,
			identifier,
		)
	})

	regex4 := regexp.MustCompile(`\n*\[\/spoilerbox\]\n?`)
	text = regex4.ReplaceAllString(text, "</div></div>")

	return text
}

func parseYoutube(text string) string {

	regex := regexp.MustCompile(`\[youtube\]https:\/\/(.*)youtube\.com\/watch\?v=([^&]+)`)
	text = regex.ReplaceAllString(text, "<div class='bbcode-video-box'><div class='bbcode-video'><iframe src='https://www.youtube.com/embed/$2")

	regex2 := regexp.MustCompile(`\[youtube\]https:\/\/(.*)youtu\.be\/([^?]+)`)
	text = regex2.ReplaceAllString(text, "<div class='bbcode-video-box'><div class='bbcode-video'><iframe src='https://www.youtube.com/embed/$2")

	regex3 := regexp.MustCompile(`\[youtube\]https:\/\/(.*)youtube\.com\/embed\/([^?]+)`)
	text = regex3.ReplaceAllString(text, "<div class='bbcode-video-box'><div class='bbcode-video'><iframe src='https://www.youtube.com/embed/$2")

	regex4 := regexp.MustCompile(`\[youtube\](.*)`)
	text = regex4.ReplaceAllString(text, "<div class='bbcode-video-box'><div class='bbcode-video'><iframe src='https://www.youtube.com/embed/$1")

	regex5 := regexp.MustCompile(`\[\/youtube\]\n?`)
	text = regex5.ReplaceAllString(text, "?rel=0' frameborder='0' allowfullscreen></iframe></div></div>")

	return text
}

func parseTwitch(text string) string {

	regex := regexp.MustCompile(`\[twitch\]https:\/\/(.*)\.twitch\.tv\/(.*)\/clip\/([^?]+)`)
	text = regex.ReplaceAllString(text, "<div class='bbcode-video-box'><div class='bbcode-video'><iframe src='https://clips.twitch.tv/embed?clip=$3")

	regex2 := regexp.MustCompile(`\[twitch\](.*)`)
	text = regex2.ReplaceAllString(text, "<div class='bbcode-video-box'><div class='bbcode-video'><iframe src='https://clips.twitch.tv/embed?clip=$1")

	regex3 := regexp.MustCompile(`\[\/twitch\]\n?`)
	text = regex3.ReplaceAllString(text, fmt.Sprintf(
		"&parent=%s' frameborder='0' allowfullscreen></iframe></div></div>",
		strings.Split(settings.GetSettings().APP_BASE_URL, "://")[1],
	))

	return text
}

func parseCode(text string) string {
	regex := regexp.MustCompile(`(?s)\[(code|c)\]\n?(.*?)\n?\[\/(code|c)\]\n?`)

	text = regex.ReplaceAllString(text, "<pre><code class='bbcode-code'>$2</code></pre>")
	return text
}

// Ripple specific.
func parseSeparator(text string) string {
	text = strings.Replace(text, "[hr]", "<div class='ui divider'></div>", -1)
	return text
}

var policy = func() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()

	p.AllowElements(
		"div", "center", "strong", "h2", "em", "strike", "u", "span", "audio", "a", "blockquote", "img", "ol", "li", "ul", "button", "i", "pre", "code", "h4", "iframe", "br",
	)

	p.AllowStandardURLs()
	p.AllowStyling()
	p.AllowAttrs("style", "class", "id").Globally()
	p.AllowAttrs("href", "rel").OnElements("a")
	p.AllowAttrs("loading", "src").OnElements("img")
	p.AllowAttrs("controls", "src", "preload").OnElements("audio")
	p.AllowAttrs("src", "frameborder", "allowfullscreen").OnElements("iframe")
	p.AllowAttrs("onclick", "type").OnElements("button")

	p.AllowAttrs("data-tooltip", "data-position").OnElements("a", "span")

	return p
}()

func ConvertBBCodeToHTML(bbcode string) string {

	// block
	bbcode = parseImagemap(bbcode)
	bbcode = parseBox(bbcode)
	bbcode = parseCode(bbcode)
	bbcode = parseList(bbcode)
	bbcode = parseNotice(bbcode)
	bbcode = parseQuote(bbcode)
	bbcode = parseHeading(bbcode)

	// inline
	bbcode = parseAudio(bbcode)
	bbcode = parseBold(bbcode)
	bbcode = parseCentre(bbcode)
	bbcode = parseColour(bbcode)
	bbcode = parseEmail(bbcode)
	bbcode = parseImage(bbcode)
	bbcode = parseItalic(bbcode)
	bbcode = parseSize(bbcode)
	bbcode = parseSpoiler(bbcode)
	bbcode = parseStrike(bbcode)
	bbcode = parseUnderline(bbcode)
	bbcode = parseUrl(bbcode)
	bbcode = parseSeparator(bbcode)
	bbcode = parseYoutube(bbcode)
	bbcode = parseTwitch(bbcode)
	bbcode = parseProfile(bbcode)

	bbcode = strings.Replace(bbcode, "\n", "<br>", -1)

	bbcodeFinal := fmt.Sprintf("<div class='bbcode-container'>%s</div>", bbcode)
	// Sanitize HTML
	return policy.Sanitize(bbcodeFinal)
}
