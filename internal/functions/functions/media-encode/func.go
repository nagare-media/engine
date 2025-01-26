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
	"fmt"
	goio "io"
	"os/exec"
	"strconv"
	"strings"

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
	Name = "media-encode"

	MainInputPortName  = "in"
	MainOutputPortName = "out"

	BitrateParameterKey       = Name + ".engine.nagare.media/bitrate"
	ResolutionParameterKey    = Name + ".engine.nagare.media/resolution"
	RawInputArgsParameterKey  = Name + ".engine.nagare.media/raw-input-args"
	RawOutputArgsParameterKey = Name + ".engine.nagare.media/raw-output-args"
)

// function encodes input stream.
type function struct {
	tsk *nbmpv2.Task

	bitrate       string
	resolution    string
	rawInputArgs  []string
	rawOutputArgs []string
}

var _ nbmp.Function = &function{}

// Exec media-encode function.
func (f *function) Exec(ctx context.Context) error {
	l := log.FromContext(ctx).WithName(Name)
	ctx = log.IntoContext(ctx, l)

	mgr := io.NewManager()

	//       ┌───────function────────┐
	//    ┌──┴─┐   ┌──────┐   ┌──────┴──┐
	// ──▶│ in │──▶│ main │──▶│ out.{n} │──▶
	//    └──┬─┘   └──────┘   └──────┬──┘
	//       └───────────────────────┘

	// streams
	mainStreamInReader, mainStreamInWriter := goio.Pipe()
	mainStreamOut := io.NewStream()

	// stream processor
	mainStreamProcessor := f.mainStreamProcessor(mainStreamInReader, mainStreamOut)
	if err := mgr.ManageStreamProcessor(mainStreamProcessor); err != nil {
		return err
	}

	// input port
	if len(f.tsk.General.InputPorts) == 0 {
		return io.PortMissing
	}
	for _, p := range f.tsk.General.InputPorts {
		port, err := io.NewInputPortForInputs(p, &f.tsk.Input)
		if err != nil {
			return err
		}

		if port.Name() != MainInputPortName {
			return io.PortNameUnknown
		}

		port.Connect(io.NewConnection(mainStreamInWriter))
		if err := mgr.ManagePort(port, true); err != nil {
			return err
		}

		break
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
			mainStreamOut.Connect(port)
			if err := mgr.ManagePort(port, true); err != nil {
				return err
			}
		default:
			return io.PortNameUnknown
		}
	}

	return mgr.Start(ctx)
}

func (f *function) mainStreamProcessor(in goio.ReadCloser, out io.Stream) io.StreamProcessor {
	return io.StreamProcessorFunc(func(ctx context.Context) error {
		defer in.Close()
		defer out.Close()

		l := log.FromContext(ctx).WithName("main-stream")
		ctx = log.IntoContext(ctx, l)
		logWriter := executils.NewLogWriter(executils.LoggerLogFunc(l.WithName("ffmpeg")), true)
		defer logWriter.Close()

		// FFmpeg args
		ffmpegOutputArgs := f.rawOutputArgs
		if ffmpegOutputArgs == nil {
			ffmpegOutputArgs = []string{
				"-s", f.resolution,
				"-c:v", "libx264",
				"-preset:v", "veryfast",
				"-tune", "zerolatency",
				"-profile:v", "main",
				"-b:v", f.bitrate, "-minrate:v", f.bitrate, "-maxrate:v", f.bitrate, "-bufsize:v", f.bitrate,
			}
		}
		ffmpegArgs := make([]string, 0, len(f.rawInputArgs)+len(ffmpegOutputArgs)+9)
		ffmpegArgs = append(ffmpegArgs, "-hide_banner", "-nostats") // global (2 elements)
		if f.rawInputArgs != nil {
			ffmpegArgs = append(ffmpegArgs, f.rawInputArgs...) // input args
		}
		ffmpegArgs = append(ffmpegArgs, "-f", "mpegts", "-i", "-") // input  (4 elements)
		ffmpegArgs = append(ffmpegArgs, ffmpegOutputArgs...)       // output
		ffmpegArgs = append(ffmpegArgs, "-f", "mpegts", "-")       // format (3 elements)

		// define FFmpeg process
		ffmpegCmd := exec.CommandContext(ctx, "ffmpeg", ffmpegArgs...)
		ffmpegCmd.Stdin = in
		ffmpegCmd.Stderr = logWriter
		ffmpegCmd.Stdout = out
		ffmpegCmd.Cancel = executils.GracefulShutdown(ffmpegCmd)

		// start FFmpeg process
		return ffmpegCmd.Run()
	})
}

// BuildTask from media-encode function.
func BuildTask(ctx context.Context, t *nbmpv2.Task) (nbmp.Function, error) {
	f := &function{
		tsk: t,
	}

	// config: bitrate
	if b, ok := nbmputils.GetStringParameterValue(t.Configuration.Parameters, BitrateParameterKey); ok {
		bitrate, err := strconv.ParseUint(b, 10, 64)
		if err != nil {
			return nil, err
		}
		f.bitrate = fmt.Sprintf("%dk", bitrate)
	}

	// config: resolution
	if res, ok := nbmputils.GetStringParameterValue(t.Configuration.Parameters, ResolutionParameterKey); ok {
		f.resolution = res
	}

	// config: raw-input-args
	if argstr, ok := nbmputils.GetStringParameterValue(t.Configuration.Parameters, RawInputArgsParameterKey); ok {
		f.rawInputArgs = strings.Split(argstr, " ")
	}

	// config: raw-output-args
	if argstr, ok := nbmputils.GetStringParameterValue(t.Configuration.Parameters, RawOutputArgsParameterKey); ok {
		f.rawOutputArgs = strings.Split(argstr, " ")
	}

	return f, nil
}

func init() {
	functions.TaskBuilders.Register(Name, BuildTask)
}
