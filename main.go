package main

import (
	"bufio"
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"
	"sync"

	"github.com/PerformLine/go-stockutil/colorutil"
	flags "github.com/jessevdk/go-flags"
	"github.com/jwalton/gchalk"

	"github.com/ruckc/dockerlogs/pkg"
)

type Options struct {
	Tail       string `short:"t" long:"tail" description:"Number of lines to show from the end of the logs" default:"all"`
	Containers struct {
		Names []string `positional-arg-name:"containers"`
	} `positional-args:"yes"`
}

func main() {
	opts := Options{}
	parser := flags.NewParser(&opts, flags.Default)
	parser.ShortDescription = "tails multiple containers concurrently"
	parser.LongDescription = "tails multiple containers concurrently"

	_, err := parser.Parse()
	if err != nil {
		return
	}

	ctx := context.Background()
	dp := pkg.NewLocalDockerProvider()
	dp.Refresh(ctx)
	sources := dp.GetSources()

	prefixLength := 0
	for _, v := range sources {
		prefixLength = Max(prefixLength, len(v.Name()))
	}

	wg := sync.WaitGroup{}

	filteredSources := make([]pkg.Source, 0)
	for _, v := range sources {
		if len(opts.Containers.Names) == 0 || contains(opts.Containers.Names, v.Name()) {
			filteredSources = append(filteredSources, v)
		}
	}

	streams := make([]StreamPair, len(filteredSources))
	for i, v := range filteredSources {
		name := v.Name()
		out, err := v.Tail(ctx, true, opts.Tail)

		r, g, b := generateRgb(i, len(streams))

		streams[i] = StreamPair{
			prefix: fmt.Sprintf("%-*s", prefixLength, name),
			color:  gchalk.RGB(r, g, b),
			bold:   gchalk.WithRGB(r, g, b).Bold,
			stdout: bufio.NewScanner(out),
			stderr: bufio.NewScanner(err),
			wg:     &wg,
		}
	}

	out := make(chan string)
	wg.Add(len(streams) * 2)
	go func() {
		wg.Wait()
		close(out)
	}()

	for _, stream := range streams {
		stream.Tail(out)
	}

	for val := range out {
		fmt.Println(val)
	}
}

type StreamPair struct {
	prefix string
	color  gchalk.ColorFn
	bold   gchalk.ColorFn
	stdout *bufio.Scanner
	stderr *bufio.Scanner
	wg     *sync.WaitGroup
}

func (sp *StreamPair) Tail(out chan string) {
	go tail(out, sp.stdout, sp.prefix, sp.color, sp.wg)
	go tail(out, sp.stderr, sp.prefix, sp.bold, sp.wg)
}

func tail(out chan string, in *bufio.Scanner, prefix string, color gchalk.ColorFn, wg *sync.WaitGroup) {
	defer wg.Done()
	for in.Scan() {
		bytes := in.Bytes()
		if len(bytes) > 8 {
			out <- color(prefix + " " + string(bytes[8:]))
		}
	}
}

func nameToRgb(name string) (r uint8, g uint8, b uint8) {
	h := float64(hash(name))
	s := rand.Float64()
	l := 0.5
	return colorutil.HslToRgb(h, s, l)
}

func generateRgb(i int, size int) (r uint8, g uint8, b uint8) {
	h := (float64(i) / float64(size)) * 360
	s := rand.Float64()
	l := 0.5
	return colorutil.HslToRgb(h, s, l)
}

func hash(text string) uint32 {
	h := fnv.New32()
	h.Write([]byte(text))
	return h.Sum32()
}

func Max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
