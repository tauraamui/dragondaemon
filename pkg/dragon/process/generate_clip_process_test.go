package process_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/dragon/process"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

// 30 fps * 2 seconds per clip
const framesPerClip = 60
const persistLoc = "/testroot/clips"

func TestNewGenerateClipProcess(t *testing.T) {
	frames := make(chan video.Frame)
	generatedClips := make(chan video.Clip)

	is := is.New(t)
	proc := process.NewGenerateClipProcess(frames, generatedClips, framesPerClip, persistLoc)
	is.True(proc != nil)
}

func TestGenerateClipProcessCreatesClipsWithSpecifiedFrameAmount(t *testing.T) {
	framesChan := make(chan video.Frame)
	generatedClipsChan := make(chan video.Clip)
	numClipsToGen := 3

	is := is.New(t)
	proc := process.NewGenerateClipProcess(framesChan, generatedClipsChan, framesPerClip, persistLoc)
	proc.Start()

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error)
	sentFramesCount := 0
	go func(ctx context.Context, timeout <-chan time.Time, errs chan error, sentFramesCount *int) {
	sendFramesProcLoop:
		for {
			select {
			case <-timeout:
				errs <- errors.New("test timeout 3s limit exceeded")
				break sendFramesProcLoop
			case <-ctx.Done():
				errs <- nil
				break sendFramesProcLoop
			default:
				*sentFramesCount++
				framesChan <- &mockFrame{}
			}
		}
	}(ctx, time.After(3*time.Second), errChan, &sentFramesCount)

	generatedClips := []video.Clip{}
	timeout := time.After(3 * time.Second)
receiveClipsProcLoop:
	for {
		select {
		case err := <-errChan:
			t.Fatal(err)
		case <-timeout:
			t.Fatal("test timeout 3s limit exceeded")
		case c := <-generatedClipsChan:
			is.True(c != nil)
			generatedClips = append(generatedClips, c)
			if len(generatedClips) == numClipsToGen {
				cancel()
				break receiveClipsProcLoop
			}
		}
	}

	proc.Stop()
	proc.Wait()

	is.Equal(len(generatedClips), numClipsToGen)
	is.True(sentFramesCount >= numClipsToGen*framesPerClip)
	for i := 0; i < numClipsToGen; i++ {
		is.Equal(len(generatedClips[i].GetFrames()), framesPerClip)
	}
}
