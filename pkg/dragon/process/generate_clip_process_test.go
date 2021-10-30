package process_test

import (
	"context"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/broadcast"
	"github.com/tauraamui/dragondaemon/pkg/dragon/process"
	"github.com/tauraamui/dragondaemon/pkg/video"
	"github.com/tauraamui/xerror"
)

// 30 fps * 2 seconds per clip
const framesPerClip = 60
const persistLoc = "/testroot/clips"

func TestNewGenerateClipProcess(t *testing.T) {
	b := broadcast.New(0)
	frames := make(chan video.Frame)
	generatedClips := make(chan video.Clip)

	is := is.New(t)
	proc := process.NewGenerateClipProcess(b.Listen(), frames, generatedClips, framesPerClip, persistLoc)
	is.True(proc != nil)
}

func TestGenerateClipProcessCreatesClipWithBroadcastEventForEarlyPause(t *testing.T) {
	b := broadcast.New(0)
	framesChan := make(chan video.Frame)
	generatedClipsChan := make(chan video.Clip)

	is := is.New(t)
	proc := process.NewGenerateClipProcess(b.Listen(), framesChan, generatedClipsChan, framesPerClip, persistLoc)
	proc.Start()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error)
	onLastClip := make(chan bool)
	clipsToGenerate := 30
	go func(ctx context.Context, timeout <-chan time.Time, onLastClip chan bool, done chan error, frames chan video.Frame) {
		defer close(done)
		for {
			time.Sleep(1 * time.Microsecond)
			select {
			case <-timeout:
				done <- xerror.New("test timeout 3s limit exceeded")
				return
			case <-ctx.Done():
				return
			case yes := <-onLastClip:
				if yes {
					for i := 0; i < framesPerClip; i++ {
						if i == (framesPerClip/2)+1 {
							b.Send(process.CAM_SWITCHED_OFF_EVT)
							return
						}
						frames <- &mockFrame{}
					}
				}
			default:
				frames <- &mockFrame{}
			}
		}
	}(ctx, time.After(3*time.Second), onLastClip, done, framesChan)

	for i := 0; i < clipsToGenerate; i++ {
		clip := <-generatedClipsChan
		if i < clipsToGenerate-1 {
			is.Equal(len(clip.GetFrames()), framesPerClip)
		} else {
			frameCount := len(clip.GetFrames())
			is.True(frameCount < framesPerClip || frameCount <= framesPerClip/2)
		}

		if i == clipsToGenerate-2 {
			onLastClip <- true
		}
	}

	cancel()
	is.NoErr(<-done)
}

func TestGenerateClipProcessCreatesClipsWithSpecifiedFrameAmount(t *testing.T) {
	b := broadcast.New(0)
	framesChan := make(chan video.Frame)
	generatedClipsChan := make(chan video.Clip)
	numClipsToGen := 3

	is := is.New(t)
	proc := process.NewGenerateClipProcess(b.Listen(), framesChan, generatedClipsChan, framesPerClip, persistLoc)
	proc.Start()

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error)
	sentFramesCount := 0
	go func(ctx context.Context, timeout <-chan time.Time, errs chan error, sentFramesCount *int) {
	sendFramesProcLoop:
		for {
			select {
			case <-timeout:
				errs <- xerror.New("test timeout 3s limit exceeded")
				break sendFramesProcLoop
			case <-ctx.Done():
				errs <- nil
				break sendFramesProcLoop
			default:
				framesChan <- &mockFrame{
					data: []byte{0x0A << *sentFramesCount},
				}
				*sentFramesCount++
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
		clipsFrames := generatedClips[i].GetFrames()
		is.Equal(len(clipsFrames), framesPerClip)

		for j := 0; j < len(clipsFrames); j++ {
			var value byte = 0x0A << (j + (framesPerClip * i))
			is.Equal([]byte{value}, clipsFrames[j].DataRef())
		}
	}
}
