package main

import (
	"bufio"
	"flag"
	"fmt"
	"html"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

type benchSample struct {
	label string
	value float64
}

type benchGroup struct {
	samples  []benchSample
	axisUnit string
}

func main() {
	inputPath := flag.String("in", "", "path to benchmark log (stdin when empty)")
	outDir := flag.String("out-dir", ".", "output directory for generated plots")
	dataDir := flag.String("data-dir", "", "directory to write textual benchmark summaries (disabled when empty)")
	width := flag.Int("width", 1200, "SVG width in pixels")
	height := flag.Int("height", 640, "SVG width in pixels")
	flag.Parse()

	if *width <= 0 || *height <= 0 {
		log.Fatal("width and height must be positive values")
	}

	var scanner *bufio.Scanner
	if *inputPath == "" {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		f, err := os.Open(*inputPath)
		if err != nil {
			log.Fatalf("opening input file: %v", err)
		}
		defer f.Close()
		scanner = bufio.NewScanner(f)
	}

	groups, order, err := parseBench(scanner)
	if err != nil {
		log.Fatalf("parsing benchmarks: %v", err)
	}

	if len(order) == 0 {
		log.Println("no benchmark data detected")
		return
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatalf("creating output directory: %v", err)
	}
	if *dataDir != "" {
		if err := os.MkdirAll(*dataDir, 0o755); err != nil {
			log.Fatalf("creating data directory: %v", err)
		}
	}

	for _, name := range order {
		group := groups[name]
		if len(group.samples) == 0 {
			continue
		}
		outPath := filepath.Join(*outDir, fmt.Sprintf("benchplot_%s.svg", sanitizeFilename(name)))
		if err := renderGroup(name, group, *width, *height, outPath); err != nil {
			log.Printf("skipping %s: %v", name, err)
			continue
		}
		if *dataDir != "" {
			if err := writeGroupData(name, group, *dataDir); err != nil {
				log.Printf("writing data %s: %v", name, err)
				continue
			}
		}
		log.Printf("wrote %s", outPath)
	}
}

func writeGroupData(name string, group benchGroup, dir string) error {
	path := filepath.Join(dir, fmt.Sprintf("benchdata_%s.txt", sanitizeFilename(name)))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "# %s\n", name)
	fmt.Fprintf(f, "# axis unit: %s\n", group.axisUnit)
	fmt.Fprintf(f, "label,value\n")
	for _, sample := range group.samples {
		fmt.Fprintf(f, "%s,%g\n", sample.label, sample.value)
	}
	return nil
}

func parseBench(scanner *bufio.Scanner) (map[string]benchGroup, []string, error) {
	groups := make(map[string]benchGroup)
	order := make([]string, 0, 8)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "Benchmark") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		name := fields[0]
		rawValue := fields[2]
		unit := fields[3]

		parsed, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			continue
		}

		value, axisUnit := normalizeValue(parsed, unit)
		groupName, label := splitBenchName(name)

		group, exists := groups[groupName]
		if !exists {
			order = append(order, groupName)
		}
		if group.axisUnit == "" {
			group.axisUnit = axisUnit
		}
		group.samples = append(group.samples, benchSample{label: label, value: value})
		groups[groupName] = group
	}

	return groups, order, scanner.Err()
}

func renderGroup(name string, group benchGroup, width, height int, outPath string) error {
	marginX, marginY := 80, 60
	chartWidth := width - marginX*2
	chartHeight := height - marginY*2
	if chartWidth <= 0 || chartHeight <= 0 {
		return fmt.Errorf("width/height too small")
	}

	maxVal := 0.0
	for _, sample := range group.samples {
		if sample.value > maxVal {
			maxVal = sample.value
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	tickCount := 5
	rawStep := maxVal / float64(tickCount)
	step := niceStep(rawStep)
	maxTick := math.Ceil(maxVal/step) * step
	if maxTick == 0 {
		maxTick = step
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img">`, width, height, width, height)
	fmt.Fprintf(&b, `<style>text{font-family:Verdana,sans-serif;font-size:12px;fill:#1e1e1e;}</style>`)
	fmt.Fprintf(&b, `<rect width="100%%" height="100%%" fill="#ffffff"/>`)

	// Axes
	fmt.Fprintf(&b, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#444" stroke-width="1.2"/>`, marginX, marginY, marginX, marginY+chartHeight)
	fmt.Fprintf(&b, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#444" stroke-width="1.2"/>`, marginX, marginY+chartHeight, marginX+chartWidth, marginY+chartHeight)

	// Y ticks and labels
	for i := 0; i <= tickCount; i++ {
		value := float64(i) * step
		y := float64(marginY) + float64(chartHeight)*(1-value/maxTick)
		fmt.Fprintf(&b, `<line x1="%d" y1="%.1f" x2="%d" y2="%.1f" stroke="#ddd" stroke-width="1"/>`, marginX, y, marginX+chartWidth, y)
		fmt.Fprintf(&b, `<text x="%d" y="%.1f" text-anchor="end"> %s</text>`, marginX-8, y+4, html.EscapeString(formatValue(value)))
	}

	// Axis labels & title
	title := html.EscapeString(name)
	fmt.Fprintf(&b, `<text x="%d" y="%d" text-anchor="middle" font-size="16px" font-weight="600">%s</text>`, width/2, marginY/2, title)
	fmt.Fprintf(&b, `<text x="%d" y="%d" text-anchor="middle">%s</text>`, marginX+chartWidth/2, marginY+chartHeight+35, html.EscapeString(group.axisUnit))

	points := make([]string, 0, len(group.samples))
	for i, sample := range group.samples {
		var x float64
		if len(group.samples) == 1 {
			x = float64(marginX + chartWidth/2)
		} else {
			x = float64(marginX) + float64(i)*(float64(chartWidth)/float64(len(group.samples)-1))
		}
		y := float64(marginY) + float64(chartHeight)*(1-sample.value/maxTick)
		points = append(points, fmt.Sprintf("%.1f,%.1f", x, y))
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="5" fill="#1f77b4"/>`, x, y)
		fmt.Fprintf(&b, `<text x="%.1f" y="%d" text-anchor="middle" fill="#333">%s</text>`, x, marginY+chartHeight+55, html.EscapeString(sample.label))
	}

	fmt.Fprintf(&b, `<polyline points="%s" fill="none" stroke="#ff7f0e" stroke-width="3" stroke-linejoin="round" stroke-linecap="round"/>`, strings.Join(points, " "))

	fmt.Fprintf(&b, `<text x="%d" y="%d" text-anchor="middle" font-size="12px">%.0f %s max</text>`, marginX+chartWidth/2, marginY-12, maxVal, html.EscapeString(group.axisUnit))

	b.WriteString("</svg>")

	return os.WriteFile(outPath, []byte(b.String()), 0o644)
}

func niceStep(raw float64) float64 {
	if raw <= 0 {
		return 1
	}
	if raw < 1 {
		return 1
	}
	exp := math.Pow(10, math.Floor(math.Log10(raw)))
	frac := raw / exp

	switch {
	case frac <= 1:
		return exp
	case frac <= 2:
		return 2 * exp
	case frac <= 5:
		return 5 * exp
	default:
		return 10 * exp
	}
}

func formatValue(v float64) string {
	if v == 0 {
		return "0"
	}
	if v >= 1000 {
		return fmt.Sprintf("%.0f", v)
	}
	if v >= 10 {
		return fmt.Sprintf("%.1f", v)
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", v), "0"), ".")
}

func normalizeValue(value float64, unit string) (float64, string) {
	if idx := strings.Index(unit, "/"); idx != -1 {
		head := unit[:idx]
		if scale, ok := timeScale(head); ok {
			return value * scale, "ns/op"
		}
	}
	return value, unit
}

func timeScale(unit string) (float64, bool) {
	switch unit {
	case "ns":
		return 1, true
	case "us", "µs":
		return 1e3, true
	case "ms":
		return 1e6, true
	case "s":
		return 1e9, true
	default:
		return 0, false
	}
}

func splitBenchName(name string) (string, string) {
	if idx := strings.Index(name, "/"); idx != -1 {
		return name[:idx], name[idx+1:]
	}
	return name, name
}

func sanitizeFilename(name string) string {
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}
	if b.Len() == 0 {
		return "benchplot"
	}
	return b.String()
}
