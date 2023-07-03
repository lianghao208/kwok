/*
Copyright 2023 The Kubernetes Authors.

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

package exec

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"sigs.k8s.io/kwok/pkg/utils/path"
)

// IOStreams contains the standard streams.
type IOStreams struct {
	// In think, os.Stdin
	In io.Reader
	// Out think, os.Stdout
	Out io.Writer
	// ErrOut think, os.Stderr
	ErrOut io.Writer
}

type optCtx int

type execOptions struct {
	// Dir is the working directory of the command.
	Dir string
	// Env is the environment variables of the command.
	Env []string
	// IOStreams contains the standard streams.
	IOStreams
	// PipeStdin is true if the command's stdin should be piped.
	PipeStdin bool
}

func (e *execOptions) DeepCopy() *execOptions {
	return &execOptions{
		Dir:       e.Dir,
		Env:       append([]string(nil), e.Env...),
		IOStreams: e.IOStreams,
		PipeStdin: e.PipeStdin,
	}
}

// WithPipeStdin returns a context with the given pipeStdin option.
func WithPipeStdin(ctx context.Context, pipeStdin bool) context.Context {
	ctx, opt := withExecOptions(ctx)
	opt.PipeStdin = pipeStdin
	return ctx
}

// WithEnv returns a context with the given environment variables.
func WithEnv(ctx context.Context, env []string) context.Context {
	ctx, opt := withExecOptions(ctx)
	opt.Env = append(opt.Env, env...)
	return ctx
}

// WithDir returns a context with the given working directory.
func WithDir(ctx context.Context, dir string) context.Context {
	ctx, opt := withExecOptions(ctx)
	opt.Dir = dir
	return ctx
}

// WithIOStreams returns a context with the given IOStreams.
func WithIOStreams(ctx context.Context, streams IOStreams) context.Context {
	ctx, opt := withExecOptions(ctx)
	opt.IOStreams = streams
	return ctx
}

// WithStdIO returns a context with the standard IOStreams.
func WithStdIO(ctx context.Context) context.Context {
	return WithIOStreams(ctx, IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	})
}

// WithReadWriter returns a context with the given io.ReadWriter as the In and Out streams.
func WithReadWriter(ctx context.Context, rw io.ReadWriter) context.Context {
	return WithIOStreams(ctx, IOStreams{
		In:  rw,
		Out: rw,
	})
}

// WithAllWriteTo returns a context with the given io.Writer as the In, Out, and ErrOut streams.
func WithAllWriteTo(ctx context.Context, w io.Writer) context.Context {
	return WithIOStreams(ctx, IOStreams{
		ErrOut: w,
		Out:    w,
	})
}

// WithWriteTo returns a context with the given io.Writer as the Out stream.
func WithWriteTo(ctx context.Context, w io.Writer) context.Context {
	return WithIOStreams(ctx, IOStreams{
		Out: w,
	})
}

// WithAllWriteToErrOut returns a context with the given io.Writer as the ErrOut stream.
func WithAllWriteToErrOut(ctx context.Context) context.Context {
	return WithAllWriteTo(ctx, os.Stderr)
}

func withExecOptions(ctx context.Context) (context.Context, *execOptions) {
	v := ctx.Value(optCtx(0))
	if v == nil {
		opt := &execOptions{}
		return context.WithValue(ctx, optCtx(0), opt), opt
	}
	opt := v.(*execOptions).DeepCopy()
	return context.WithValue(ctx, optCtx(0), opt), opt
}

func fromExecOptions(ctx context.Context) *execOptions {
	v := ctx.Value(optCtx(0))
	if v == nil {
		return &execOptions{}
	}
	return v.(*execOptions)
}

// Exec executes the given command and returns the output.
func Exec(ctx context.Context, name string, args ...string) error {
	cmd := command(ctx, name, args...)
	opt := fromExecOptions(ctx)
	if opt.Env != nil {
		cmd.Env = opt.Env
		cmd.Env = append(os.Environ(), cmd.Env...)
	}
	cmd.Dir = opt.Dir

	if opt.In != nil {
		if opt.PipeStdin {
			inPipe, err := cmd.StdinPipe()
			if err != nil {
				return err
			}
			go func() {
				_, _ = io.Copy(inPipe, opt.In)
			}()
		} else {
			cmd.Stdin = opt.In
		}
	}

	cmd.Stdout = opt.Out
	cmd.Stderr = opt.ErrOut

	if cmd.Stderr == nil {
		buf := bytes.NewBuffer(nil)
		cmd.Stderr = buf
	}

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("cmd start: %s %s: %w", name, strings.Join(args, " "), err)
	}

	err = cmd.Wait()
	if err != nil {
		if buf, ok := cmd.Stderr.(*bytes.Buffer); ok {
			return fmt.Errorf("cmd wait: %s %s: %w\n%s", name, strings.Join(args, " "), err, buf.String())
		}
		return fmt.Errorf("cmd wait: %s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

// FormatExec prints the command to be executed to the output stream.
func FormatExec(ctx context.Context, name string, args ...string) string {
	opt := fromExecOptions(ctx)
	out := bytes.NewBuffer(nil)
	if opt.Dir != "" {
		_, _ = fmt.Fprintf(out, "cd %s && ", opt.Dir)
	}

	if len(opt.Env) != 0 {
		_, _ = fmt.Fprintf(out, "%s ", strings.Join(opt.Env, " "))
	}

	_, _ = fmt.Fprintf(out, "%s", path.Base(name))

	for _, arg := range args {
		_, _ = fmt.Fprintf(out, " %s", arg)
	}
	return out.String()
}
