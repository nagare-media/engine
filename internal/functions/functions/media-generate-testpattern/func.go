/*
Copyright 2022-2025 The nagare media authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package function

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/nagare-media/engine/internal/functions"
	"github.com/nagare-media/engine/internal/functions/executils"
	"github.com/nagare-media/engine/internal/functions/io"
	"github.com/nagare-media/engine/pkg/nbmp"
	nbmputils "github.com/nagare-media/engine/pkg/nbmp/utils"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

// Function description
const (
	Name = "media-generate-testpattern"

	MainOutputPortName = "out"

	DurationParameterKey = "media-generate-testpattern.engine.nagare.media/duration"
)

var ffmpegScriptTmpl = template.Must(template.New(Name + "/ffmpeg-script").Parse(`
	exec ffmpeg -hide_banner -nostats \
		-re -f lavfi -i "
			testsrc2=size=1920x1080:rate=25,
			drawbox=x=0:y=0:w=1400:h=100:c=black@.6:t=fill,
			drawtext=x= 10:y=10:fontsize=108:fontcolor=white:text='%{pts\:gmtime\:{{ .SecondsSinceEpoch }}\:%Y-%m-%d}',
			drawtext=x=690:y=10:fontsize=108:fontcolor=white:timecode='{{ .Time }}\:00':rate=25:tc24hmax=1,
			setparams=field_mode=prog:range=tv:color_primaries=bt709:color_trc=bt709:colorspace=bt709,
			format=yuv420p
		" \
		{{ .DurationFlag }} \
		-fflags genpts \
		-c:v libx264 \
			-preset:v veryfast \
			-tune zerolatency \
			-profile:v main \
			-crf:v 23 -bufsize:v:0 4500k -maxrate:v 5000k \
			-g:v 100000 -keyint_min:v 50000 -force_key_frames:v "expr:gte(t,n_forced*2)" \
			-x264opts no-open-gop=1 \
			-bf 2 -b_strategy 2 -refs 1 \
			-rc-lookahead 24 \
			-export_side_data prft \
			-field_order progressive -colorspace bt709 -color_primaries bt709 -color_trc bt709 -color_range tv \
			-pix_fmt yuv420p \
		-f mpegts \
		-
`))

// function generates a test pattern input stream.
type function struct {
	tsk *nbmpv2.Task

	duration *time.Duration
}

var _ nbmp.Function = &function{}

// Exec media-generate-testpattern function.
func (f *function) Exec(ctx context.Context) error {
	l := log.FromContext(ctx).WithName(Name)
	ctx = log.IntoContext(ctx, l)

	mgr := io.NewManager()

	// ┌─────function─────┐
	// │┌──────┐   ┌──────┴──┐
	// ││ main │──▶│ out.{n} │──▶
	// │└──────┘   └──────┬──┘
	// └──────────────────┘

	// streams
	mainStream := io.NewStream()
	mainStreamProcessor := f.mainStreamProcessor(mainStream)
	if err := mgr.ManageStreamProcessor(mainStreamProcessor); err != nil {
		return err
	}

	// output ports
	if len(f.tsk.General.OutputPorts) == 0 {
		return io.PortMissing
	}
	for _, p := range f.tsk.General.OutputPorts {
		port, err := io.NewOutputPortForOutputs(p, &f.tsk.Output)
		if err != nil {
			return err
		}

		switch port.BaseName() {
		case MainOutputPortName:
			mainStream.Connect(port)
			if err := mgr.ManagePort(port); err != nil {
				return err
			}
		default:
			return io.PortNameUnknown
		}
	}

	// TODO: should we validate that everything is connected or is this simply a user error?

	return mgr.Start(ctx)
}

func (f *function) mainStreamProcessor(out io.Stream) io.StreamProcessor {
	return io.StreamProcessorFunc(func(ctx context.Context) error {
		defer out.Close()

		l := log.FromContext(ctx).WithName("main-stream")
		ctx = log.IntoContext(ctx, l)
		logWriter := executils.NewLogWriter(executils.LoggerLogFunc(l.WithName("ffmpeg")), true)
		defer logWriter.Close()

		// build command
		now := time.Now()
		durationFlag := ""
		if f.duration != nil {
			durationFlag = "-t " + strconv.FormatInt(f.duration.Microseconds(), 10) + "us"
		}
		cmd := &strings.Builder{}
		err := ffmpegScriptTmpl.Execute(cmd, struct {
			SecondsSinceEpoch int64
			Time              string
			DurationFlag      string
		}{
			SecondsSinceEpoch: now.Unix(),
			Time:              now.Format("15\\:04\\:05"),
			DurationFlag:      durationFlag,
		})
		if err != nil {
			return err
		}

		// define FFmpeg process
		ffmpegCmd := exec.CommandContext(ctx, "sh", "-c", cmd.String())
		ffmpegCmd.Stdin = nil
		ffmpegCmd.Stderr = logWriter
		ffmpegCmd.Stdout = out
		ffmpegCmd.Cancel = executils.GracefulShutdown(ffmpegCmd)

		// start FFmpeg process
		return ffmpegCmd.Run()
	})
}

// BuildTask from media-generate-testpattern function.
func BuildTask(ctx context.Context, t *nbmpv2.Task) (nbmp.Function, error) {
	f := &function{
		tsk: t,
	}

	// config: duration
	if d, ok := nbmputils.GetStringParameterValue(t.Configuration.Parameters, DurationParameterKey); ok {
		dur, err := time.ParseDuration(d)
		if err != nil {
			return nil, err
		}
		f.duration = &dur
	}

	return f, nil
}

func init() {
	functions.TaskBuilders.Register(Name, BuildTask)
}
