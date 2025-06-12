package media

import (
	"fmt"
	"path/filepath"
	"strings"

	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

type DashVariant struct {
	Name    string
	Width   int
	Height  int
	Bitrate string
	Map     int
}

var DashVariants []DashVariant = []DashVariant{
	{"1080p", 1920, 1080, "5000k", 0},
	{"720p", 1280, 720, "3000k", 1},
	{"480p", 854, 480, "1000k", 2},
}

func BuildFilterComplex(variants []DashVariant) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("[0:v]split=%d", len(variants)))
	for i := range variants {
		builder.WriteString(fmt.Sprintf("[v%d]", i+1))
	}
	builder.WriteString(";")

	for i, variant := range variants {
		builder.WriteString(fmt.Sprintf("[v%d]scale=%d:%d[v%do];", i+1, variant.Width, variant.Height, i))
	}

	return builder.String()
}

func BuildFFmpegArgs(variants []DashVariant) []ffmpeg_go.KwArgs {
	var args []ffmpeg_go.KwArgs

	for _, variant := range variants {
		args = append(args, ffmpeg_go.KwArgs{
			"c:v": "libx264",
			"b:v": fmt.Sprintf("%dk", variant.Bitrate),
			"s":   fmt.Sprintf("%dx%d", variant.Width, variant.Height),
		})
	}

	return args
}

func ConvertToDASH(inputPath string, outputDir string, variants []DashVariant) error {
	args := BuildFFmpegArgs(variants)

	return ffmpeg_go.Input(inputPath).
		Output(filepath.Join(outputDir, "manifest.mpd"), args...).
		OverWriteOutput().
		Run()
}
