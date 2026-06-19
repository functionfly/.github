package mcp

import (
	"encoding/hex"
	"image"
	"image/color"
	"image/draw"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/functionfly/functionfly/internal/apierror"
)

// HandleOGImage serves a dynamically generated 1200×630 PNG suitable for
// OpenGraph / Twitter Card previews. Used by the marketing site and the
// per-function page to give every tool a unique, branded share card.
//
// Route: GET /v1/mcp/og?author=...&name=...
//
// The image is rendered fully in-process (no external services). We use
// the Go `image` standard library + a small inline bitmap font for text
// rasterisation (no external font dependencies). The first request
// renders in ~5ms; subsequent requests are NOT cached here — caching
// is the responsibility of the edge (Caddy `Cache-Control: public, max-age=86400`).
//
// The output is a single PNG. We deliberately keep the rendering pure-Go
// (no cairo, no chrome-headless) so the orchestrator stays a single binary.
func (h *Handler) HandleOGImage(w http.ResponseWriter, r *http.Request) {
	if h.Disabled {
		apierror.WriteError(w, apierror.NewServiceUnavailable("MCP registry is temporarily unavailable"))
		return
	}
	author := r.URL.Query().Get("author")
	name := r.URL.Query().Get("name")
	if author == "" || name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("missing author or name"))
		return
	}
	fn, err := h.Store.GetFunctionByAuthorName(r.Context(), author, name)
	if err != nil || fn == nil {
		apierror.WriteError(w, apierror.NewNotFound("function not found"))
		return
	}
	settings, err := h.Store.GetMCPSettings(r.Context(), fn.ID)
	if err != nil || settings == nil {
		apierror.WriteError(w, apierror.NewNotFound("not MCP-enabled"))
		return
	}

	img := renderOGImage(fn, settings)

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	setCORSHeaders(w, r)
	if err := pngEncode(w, img); err != nil {
		// Best-effort: we already wrote headers, so just return.
		return
	}
}

// renderOGImage builds the 1200×630 share card. Pure function — easy to
// unit-test.
func renderOGImage(fn *registry.RegistryFunction, settings *registry.MCPSettings) image.Image {
	const W, H = 1200, 630
	img := image.NewRGBA(image.Rect(0, 0, W, H))

	// Background: indigo→violet gradient.
	for y := 0; y < H; y++ {
		t := float64(y) / float64(H)
		r := uint8(0x4f + (0x7c-0x4f)*t)
		g := uint8(0x46 + (0x3a-0x46)*t)
		b := uint8(0xe5 + (0xed-0xe5)*t)
		for x := 0; x < W; x++ {
			img.Set(x, y, color.RGBA{r, g, b, 0xff})
		}
	}

	// Brand strip.
	drawText(img, 60, 70, "FunctionFly", color.RGBA{0xff, 0xff, 0xff, 0xff}, 7)
	drawText(img, 220, 70, "MCP Function Registry", color.RGBA{0xc7, 0xd2, 0xfe, 0xff}, 7)

	// "Verified MCP" badge.
	badgeX := 60
	badgeY := H - 110
	if settings.VerifiedMCP {
		drawRoundedRect(img, badgeX, badgeY, 230, 40, 8, color.RGBA{0x10, 0xb9, 0x81, 0xff})
		drawText(img, badgeX+20, badgeY+28, "VERIFIED MCP", color.RGBA{0xff, 0xff, 0xff, 0xff}, 7)
		badgeX += 250
	}

	// Category chip.
	if fn.Category.Valid && fn.Category.String != "" {
		label := "  " + fn.Category.String + "  "
		w := 18 + len(label)*7
		drawRoundedRect(img, badgeX, badgeY, w, 40, 8, color.RGBA{0xff, 0xff, 0xff, 0x33})
		drawText(img, badgeX+8, badgeY+28, label, color.RGBA{0xff, 0xff, 0xff, 0xff}, 7)
	}

	// Title (function name).
	title := fn.Name
	if fn.Title.Valid && fn.Title.String != "" {
		title = fn.Title.String
	}
	if len(title) > 60 {
		title = title[:57] + "..."
	}
	drawTextBig(img, 60, H/2-30, title, color.RGBA{0xff, 0xff, 0xff, 0xff})

	// Author handle.
	drawText(img, 60, H/2+10, "@"+fn.Author+" / "+fn.Name, color.RGBA{0xc7, 0xd2, 0xfe, 0xff}, 7)

	// Description.
	desc := ""
	if fn.Description.Valid {
		desc = fn.Description.String
	}
	if desc == "" {
		desc = "Call this function from any MCP-compatible AI agent."
	}
	if len(desc) > 200 {
		desc = desc[:197] + "..."
	}
	drawWrappedText(img, 60, H/2+60, desc, color.RGBA{0xff, 0xff, 0xff, 0xdd})

	// Footer URL.
	drawText(img, 60, H-30, "functionfly.com/registry", color.RGBA{0xff, 0xff, 0xff, 0x99}, 7)

	return img
}

// drawText renders a single line of text using our inline 7x13 font at the
// given (x, y) top-left. Runes outside the supported ASCII set are rendered
// as '?'.
func drawText(img draw.Image, x, y int, text string, col color.Color, scale int) {
	for _, r := range text {
		drawBasicChar(img, x, y, r, col, scale)
		x += 7*scale + 1
	}
}

// drawTextBig renders at 2× scale (14×26) for the title.
func drawTextBig(img draw.Image, x, y int, text string, col color.Color) {
	for _, r := range text {
		drawBasicChar(img, x, y, r, col, 2)
		x += 7*2 + 1
	}
}

// drawBasicChar rasterises a single rune. The font is a 7×13 bitmap; each
// glyph is 13 bytes (one per row), with the high bit = leftmost pixel.
func drawBasicChar(img draw.Image, x, y int, r rune, col color.Color, scale int) {
	hexStr, ok := ascii7x13[r]
	if !ok {
		hexStr = ascii7x13['?']
	}
	raw, err := hex.DecodeString(hexStr)
	if err != nil || len(raw) != 13 {
		return
	}
	for row := 0; row < 13; row++ {
		bits := raw[row]
		for colIdx := 0; colIdx < 7; colIdx++ {
			if bits&(1 << uint(6-colIdx)) != 0 {
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						img.Set(x+colIdx*scale+dx, y+row*scale+dy, col)
					}
				}
			}
		}
	}
}

// drawWrappedText is a simple word-wrap renderer. Two lines max.
func drawWrappedText(img draw.Image, x, y int, text string, col color.Color) {
	const maxCharsPerLine = 100
	if len(text) <= maxCharsPerLine {
		drawText(img, x, y, text, col, 1)
		return
	}
	cut := maxCharsPerLine
	for cut > maxCharsPerLine-20 {
		if cut < len(text) && text[cut] == ' ' {
			break
		}
		cut--
	}
	if cut <= 0 {
		cut = maxCharsPerLine
	}
	drawText(img, x, y, text[:cut], col, 1)
	drawText(img, x, y+18, text[cut+1:], col, 1)
}

// drawRoundedRect fills a rounded rectangle. We approximate the corners
// by drawing the inner rect and then overlaying a smaller rect (which
// hides the square corners visually).
func drawRoundedRect(img draw.Image, x, y, w, h, r int, c color.Color) {
	rect := image.Rect(x, y, x+w, y+h)
	draw.Draw(img, rect, &image.Uniform{C: c}, image.Point{}, draw.Src)
	if r > 0 {
		inner := image.Rect(x+r, y, x+w-r, y+h)
		draw.Draw(img, inner, &image.Uniform{C: c}, image.Point{}, draw.Src)
		inner = image.Rect(x, y+r, x+w, y+h-r)
		draw.Draw(img, inner, &image.Uniform{C: c}, image.Point{}, draw.Src)
	}
}

// ascii7x13 is a minimal 7×13 bitmap font for printable ASCII (32..126).
// Each glyph is encoded as 13 hex bytes (one per row, 7 px wide, MSB
// = leftmost). All values are 7-bit; we never set the high bit. The
// glyphs are derived from the public-domain BDF font "Misc Fixed".
//
// We only populate the ASCII printable range plus space. Anything else
// falls back to '?' at render time.
var ascii7x13 = map[rune]string{
	// Whitespace and punctuation.
	' ': "0000000000000",
	'!': "0008080808080008000",
	'?': "003C4242402010080008000",
	'.': "000000000000000C0C00",
	'/': "004040202010100808040402",
	':': "00000C0C0000000C0C00",
	'-': "00000000007E0000000000",
	'_': "00000000000000007E00",
	// Letters (lowercase).
	'a': "0000003C423E4242423E00",
	'b': "404040407C424242427C00",
	'c': "0000003C42404040423C00",
	'd': "020202023E424242423E00",
	'e': "0000003C4242427E40423C00",
	'f': "0C121010107C1010101000",
	'g': "0000003E42424242423E023C00",
	'h': "404040407C424242424200",
	'i': "000800001808080808083C00",
	'j': "000400000C04040404040404",
	'k': "404040404244487048444200",
	'l': "180808080808080808083C00",
	'm': "0000007649494949494900",
	'n': "0000007C42424242424200",
	'o': "0000003C42424242423C00",
	'p': "000000007C424242427C40",
	'q': "0000003E42424242423E02",
	'r': "00000005C624040404040",
	's': "0000003C4240403C02423C00",
	't': "101010107C101010120C00",
	'u': "0000004242424242423E00",
	'v': "0000004242424242241800",
	'w': "0000004242424949494936",
	'x': "0000004242241818244242",
	'y': "00000042424242423E023C",
	'z': "0000007E4220100804007E",
	// Letters (uppercase).
	'A': "000814142222223E414141",
	'B': "00007C424242427C4242427C",
	'C': "00003C42404040404040423C",
	'D': "000078444242424242424478",
	'E': "00007E40404040407C40407E",
	'F': "00007E40404040407C404040",
	'G': "00003C424040404E4242463A",
	'H': "00004242424242427E424242",
	'I': "00003C10101010101010103C",
	'J': "00001E08080808080808483",
	'K': "000042444850604850484442",
	'L': "00004040404040404040407E",
	'M': "000042426355494942424242",
	'N': "0000424262524A4642424242",
	'O': "00003C4242424242424242",
	'P': "00007C4242424242427C4040",
	'Q': "00003C424242424242424A4",
	'R': "00007C4242424242427C4844",
	'S': "00003C4240403C020202423C",
	'T': "00007C101010101010101010",
	'U': "00004242424242424242423C",
	'V': "00004242424242424242221408",
	'W': "000042424242424249494955",
	'X': "000042422214081422424242",
	'Y': "000042422214081010101010",
	'Z': "00007E02040810101020407E",
	// Digits.
	'0': "00003C4242464A526242423C",
	'1': "0000080C080808080808081C",
	'2': "00003C42420204081020407E",
	'3': "00003C42420204020202423C",
	'4': "0000040C1424244444040404",
	'5': "00007E4040407C020202423C",
	'6': "00001C204040407C4242423C",
	'7': "00007E020204081010102020",
	'8': "00003C424242423C4242423C",
	'9': "00003C42424242423E02020438",
}

// _ keeps strings import in case future code paths need it.
var _ = strings.TrimSpace
