package bbcode

import (
	"fmt"
	"html"
	"html/template"
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
	regex := regexp.MustCompile(`\[(color|colour)=([^\[\]<>;:]+)\]`)
	text = regex.ReplaceAllString(text, "<span style='color: $2'>")

	regex2 := regexp.MustCompile(`\[\/(color|colour)\]`)
	text = regex2.ReplaceAllString(text, "</span>")
	return text
}

func parseAudio(text string) string {
	regex := regexp.MustCompile(`\[audio\]([^\[<>
]+)\[\/audio\]\n?`)

	text = regex.ReplaceAllString(text, "<audio controls='controls' preload='none' src='$1'></audio>")
	return text
}

func parseUrl(text string) string {
	regex := regexp.MustCompile(`\[url\]([^\[<>\r\n]+)\[\/url\]`)
	text = regex.ReplaceAllString(text, "<a rel='nofollow' href='$1'>$1</a>")

	regex2 := regexp.MustCompile(`\[url=([^\[\]<>\r\n]+)\]`)
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
	regex := regexp.MustCompile(`\[email\](([^\[<>
]+)@([^\[<>
]+))\[\/email\]`)
	text = regex.ReplaceAllString(text, "<a rel='nofollow' href='mailto:$1'>$1</a>")

	regex2 := regexp.MustCompile(`\[email=(([^\[<>
]+)@([^\[<>
]+))\]`)
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

		if strings.ContainsAny(groups[2], "[]<>") {
			return groups[0]
		}
		return fmt.Sprintf("<a href='/u/%s'>/u/%s</a>", groups[2], groups[2])
	})

	return text
}

func parseImage(text string) string {
	regex := regexp.MustCompile(`\[img\]([^\[<>
]+)\[\/img\]`)
	text = ReplaceAllStringSubmatchFunc(regex, text, func(groups []string, _ string) string {
		decoded, _ := url.QueryUnescape(groups[1])
		return fmt.Sprintf("<img src='%s' loading='lazy'/>", template.HTMLEscapeString(decoded))
	})

	// there is also a case of our old bbcode parsing [img=url][/img]
	regex2 := regexp.MustCompile(`\[img=([^\[<>
]+)\]\[\/img\]`)
	text = ReplaceAllStringSubmatchFunc(regex2, text, func(groups []string, _ string) string {
		decoded, _ := url.QueryUnescape(groups[1])
		return fmt.Sprintf("<img src='%s' loading='lazy'/>", template.HTMLEscapeString(decoded))
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
			strings.ReplaceAll(matches[regex.SubexpIndex("image_url")], "[", "&#91;"),
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
				strings.ReplaceAll(redirect, "[", "&#91;"),
				xAxisClamped,
				yAxisClamped,
				widthClamped,
				heightClamped,
				strings.ReplaceAll(title, "[", "&#91;"),
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

	text = ReplaceAllStringSubmatchFunc(regex, text, func(groups []string, _ string) string {

		return fmt.Sprintf(
			"<div class='bbcode-box'><button class='bbcode-box-btn' type='button'><i class='bbcode-box-icon fa-solid fa-angle-right'></i><span>%s</span></button><div class='bbcode-box-content bbcode-hidden'>",
			groups[1],
		)
	})

	regex2 := regexp.MustCompile(`\n*\[\/box\]\n?`)
	text = regex2.ReplaceAllString(text, "</div></div>")

	regex3 := regexp.MustCompile(`\[spoilerbox\]\n*`)
	text = ReplaceAllStringSubmatchFunc(regex3, text, func(_ []string, _ string) string {
		return fmt.Sprintf(
			"<div class='bbcode-box'><button class='bbcode-box-btn' type='button'><i class='bbcode-box-icon fa-solid fa-angle-right'></i><span>SPOILER</span></button><div class='bbcode-box-content bbcode-hidden'>",
		)
	})

	regex4 := regexp.MustCompile(`\n*\[\/spoilerbox\]\n?`)
	text = regex4.ReplaceAllString(text, "</div></div>")

	return text
}

var (
	youtubeTag = regexp.MustCompile(`\[youtube\]([^\[<>
]+)\[/youtube\]
?`)
	youtubeVideo = regexp.MustCompile(`^(?:https://(?:(?:www\.|m\.)?youtube\.com/(?:watch\?v=|embed/)|youtu\.be/))?([A-Za-z0-9_-]{11})(?:[?&][^[:space:]]*)?$`)
	twitchTag    = regexp.MustCompile(`\[twitch\]([^\[<>
]+)\[/twitch\]
?`)
	twitchClip = regexp.MustCompile(`^(?:https://(?:clips\.twitch\.tv/|(?:www\.)?twitch\.tv/[^/]+/clip/))?([A-Za-z0-9_-]+)(?:\?[^[:space:]]*)?$`)
)

func videoFrame(src string) string {
	return "<div class='bbcode-video-box'><div class='bbcode-video'><iframe src='" +
		html.EscapeString(src) + "' frameborder='0' allowfullscreen></iframe></div></div>"
}

func parseYoutube(text string) string {
	return youtubeTag.ReplaceAllStringFunc(text, func(tag string) string {
		video := youtubeVideo.FindStringSubmatch(html.UnescapeString(youtubeTag.FindStringSubmatch(tag)[1]))
		if video == nil {
			return tag
		}
		return videoFrame("https://www.youtube.com/embed/" + video[1] + "?rel=0")
	})
}

func parseTwitch(text string) string {
	if !twitchTag.MatchString(text) {
		return text
	}
	baseURL, err := url.Parse(settings.GetSettings().APP_BASE_URL)
	if err != nil {
		return text
	}
	return twitchTag.ReplaceAllStringFunc(text, func(tag string) string {
		clip := twitchClip.FindStringSubmatch(html.UnescapeString(twitchTag.FindStringSubmatch(tag)[1]))
		if clip == nil {
			return tag
		}
		return videoFrame("https://clips.twitch.tv/embed?clip=" + clip[1] + "&parent=" + url.QueryEscape(baseURL.Hostname()))
	})
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

func parseContainer(text string) string {
	regex := regexp.MustCompile(`\[container(.*?)\]`)

	text = ReplaceAllStringSubmatchFunc(regex, text, func(matches []string, _ string) string {
		style := ""
		class := ""

		if matches != nil {
			args := strings.Split(strings.TrimSpace(matches[1]), " ")
			for _, arg := range args {
				argVars := strings.SplitN(arg, "=", 2)
				switch argVars[0] {
				case "compact":
					class += "compact-container"
					break
				case "center", "centre":
					style += "margin: 0 auto;"
					break
				case "width":
					if len(argVars) != 2 {
						continue
					}
					if parsedInt, err := strconv.Atoi(argVars[1]); err == nil {
						fixedInt := Clamp(parsedInt, 0, 100)
						style += fmt.Sprintf("width: %d%%;", fixedInt)
					}
					break
				default:
					break
				}
			}
		}

		pseudoHtml := fmt.Sprintf("<div class='%s' style='%s'>", class, style)
		return pseudoHtml
	})

	text = strings.Replace(text, "[/container]", "</div>", -1)

	return text
}

func parseLeft(text string) string {
	text = strings.Replace(text, "[left]", "<div style='text-align: left;'>", -1)
	text = strings.Replace(text, "[/left]", "</div>", -1)
	return text
}

func parseRight(text string) string {
	text = strings.Replace(text, "[right]", "<div style='text-align: right;'>", -1)
	text = strings.Replace(text, "[/right]", "</div>", -1)
	return text
}

var policy = func() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()

	p.AllowElements(
		"div", "center", "strong", "h2", "em", "strike", "u", "span", "audio", "a", "blockquote", "img", "ol", "li", "ul", "button", "i", "pre", "code", "h4", "iframe", "br",
	)

	p.AllowStandardURLs()
	p.AllowAttrs("class").Matching(regexp.MustCompile(`^(?:bbcode-[a-z-]+|compact-container|ui divider|bbcode-box-content bbcode-hidden|bbcode-box-icon fa-solid fa-angle-right)$`)).Globally()
	p.AllowStyles("color", "font-size").OnElements("span")
	p.AllowStyles("text-align").MatchingEnum("left", "right", "center").OnElements("div")
	p.AllowStyles("margin").MatchingEnum("0 auto").OnElements("div")
	p.AllowStyles("width").Matching(regexp.MustCompile(`^(?:100|[0-9]{1,2})%$`)).OnElements("div")
	p.AllowStyles("list-style-type").MatchingEnum("disc").OnElements("ol")
	p.AllowStyles("left", "top", "width", "height").Matching(regexp.MustCompile(`^[0-9]{1,3}(?:\.[0-9]+)?%$`)).OnElements("a", "span")
	p.AllowAttrs("href", "rel").OnElements("a")
	p.AllowAttrs("loading", "src").OnElements("img")
	p.AllowAttrs("controls", "src", "preload").OnElements("audio")
	p.AllowAttrs("src").Matching(regexp.MustCompile(`^https://(?:www\.youtube\.com/embed/[A-Za-z0-9_-]{11}\?rel=0|clips\.twitch\.tv/embed\?clip=[A-Za-z0-9_-]+&parent=[a-zA-Z0-9.-]+)$`)).OnElements("iframe")
	p.AllowAttrs("frameborder", "allowfullscreen").OnElements("iframe")
	p.AllowAttrs("type").Matching(regexp.MustCompile(`^button$`)).OnElements("button")

	p.AllowAttrs("data-tooltip", "data-position").OnElements("a", "span")

	return p
}()

func ConvertBBCodeToHTML(bbcode string) string {
	bbcode = template.HTMLEscapeString(bbcode)

	// block
	bbcode = parseImagemap(bbcode)
	bbcode = parseBox(bbcode)
	bbcode = parseCode(bbcode)
	bbcode = parseList(bbcode)
	bbcode = parseNotice(bbcode)
	bbcode = parseQuote(bbcode)
	bbcode = parseHeading(bbcode)
	bbcode = parseContainer(bbcode)

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
	bbcode = parseLeft(bbcode)
	bbcode = parseRight(bbcode)

	bbcode = strings.Replace(bbcode, "\n", "<br>", -1)

	bbcodeFinal := fmt.Sprintf("<div class='bbcode-container'>%s</div>", bbcode)
	// Sanitize HTML
	return policy.Sanitize(bbcodeFinal)
}
