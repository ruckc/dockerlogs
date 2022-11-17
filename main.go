package main

import (
	"bufio"
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"
	"sync"

	"github.com/PerformLine/go-stockutil/colorutil"
	"github.com/jwalton/gchalk"

	"github.com/ruckc/dockerlogs/pkg"
)

func main() {
	ctx := context.Background()
	dp := pkg.NewLocalDockerProvider()
	dp.Refresh(ctx)
	sources := dp.GetSources()

	streams := make([]StreamPair, len(sources))
	prefixLength := 0
	for _, v := range sources {
		prefixLength = Max(prefixLength, len(v.Name()))
	}

	for i, v := range sources {
		out, err := v.Tail(ctx, false, "10")

		r, g, b := nameToRgb(v.Name())

		streams[i] = StreamPair{
			prefix: fmt.Sprintf("%-*s", prefixLength, v.Name()),
			color:  gchalk.RGB(r, g, b),
			bold:   gchalk.WithRGB(r, g, b).Bold,
			stdout: bufio.NewScanner(out),
			stderr: bufio.NewScanner(err),
		}
	}

	out := make(chan string)
	wg := sync.WaitGroup{}
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
}

func (sp *StreamPair) Tail(out chan string) {
	go tail(out, sp.stdout, sp.prefix, sp.color)
	go tail(out, sp.stderr, sp.prefix, sp.bold)
}

func tail(out chan string, in *bufio.Scanner, prefix string, color gchalk.ColorFn) {
	for in.Scan() {
		bytes := in.Bytes()
		out <- color(prefix + " " + string(bytes[8:]))
	}
}

func nameToRgb(name string) (r uint8, g uint8, b uint8) {
	h := float64(hash(name))
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
