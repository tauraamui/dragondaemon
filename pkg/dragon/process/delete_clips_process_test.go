package process

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

type DeleteOldClipsTestSuite struct {
	suite.Suite
	fs               afero.Fs
	timeMinuteOffset int
}

func (suite *DeleteOldClipsTestSuite) SetupSuite() {
	logging.CurrentLoggingLevel = logging.SilentLevel
	suite.fs = afero.NewMemMapFs()
	fs = suite.fs
}

func (suite *DeleteOldClipsTestSuite) TearDownSuite() {
	logging.CurrentLoggingLevel = logging.WarnLevel
	suite.fs = afero.NewOsFs()
}

func (suite *DeleteOldClipsTestSuite) SetupTest() {
	suite.timeMinuteOffset = 0
	suite.fs.MkdirAll("/testroot/clips/FakeCamera", os.ModePerm|os.ModeDir)
}

func (suite *DeleteOldClipsTestSuite) TearDownTest() {
	suite.fs.RemoveAll("/")
}

func TestDeleteOldClipsTestSuite(t *testing.T) {
	suite.Run(t, &DeleteOldClipsTestSuite{})
}

func (suite *DeleteOldClipsTestSuite) TestDeleteOldClips() {
	TimeNow = suite.timeNowQuery

	err := suite.fs.MkdirAll("/testroot/clips/FakeCamera/2010-03-11", os.ModePerm|os.ModeDir)
	require.NoError(suite.T(), err)

	conn, err := camera.ConnectWithCancel(context.TODO(), "FakeCamera", "fakeaddr", camera.Settings{
		FPS:             22,
		PersistLocation: "/testroot/clips",
		SecondsPerClip:  3,
	}, testVideoBackend{})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), conn)

	deleteProcess := Settings{
		WaitForShutdownMsg: fmt.Sprintf("Stopping test deleting old saved clips for [%s]", conn.Title()),
		Process:            DeleteOldClips(conn),
	}

	deleteClips := New(deleteProcess)
	deleteClips.Start()

	timeout := time.After(3 * time.Second)
fileExistanceProcLoop:
	for {
		time.Sleep(1 * time.Microsecond)
		exists, err := afero.Exists(suite.fs, "/testroot/clips/FakeCamera/2010-03-11")
		select {
		case <-timeout:
			suite.T().Fatal("Timeout exceeded. Delete clip process took too long...")
			break fileExistanceProcLoop
		default:
			if err != nil {
				suite.T().Fatal("Unable to query existance of clip: %w", err)
				break fileExistanceProcLoop
			}
			if exists {
				continue
			}
			if !exists {
				break fileExistanceProcLoop
			}
		}
	}

	deleteClips.Stop()
	deleteClips.Wait()

	exists, err := afero.Exists(suite.fs, "/testroot/clips/FakeCamera/2010-03-11")
	require.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}

func (suite *DeleteOldClipsTestSuite) timeNowQuery() time.Time {
	suite.timeMinuteOffset++
	return time.Now().Add(time.Minute * time.Duration(suite.timeMinuteOffset))
}

type testVideoBackend struct {
	onConnectError        error
	onConnectionReadError error
}

func (tvb testVideoBackend) Connect(context context.Context, address string) (video.Connection, error) {
	if tvb.onConnectError != nil {
		return nil, tvb.onConnectError
	}
	return testVideoConnection{
		onReadError: tvb.onConnectionReadError,
	}, nil
}

func (tvb testVideoBackend) NewFrame() video.Frame {
	return testVideoFrame{}
}

type testVideoFrame struct {
}

func (tvf testVideoFrame) DataRef() interface{} {
	return nil
}

func (tvf testVideoFrame) Close() {}

type testVideoConnection struct {
	onReadError error
}

func (tvc testVideoConnection) Read(frame video.Frame) error {
	return tvc.onReadError
}

func (tvc testVideoConnection) IsOpen() bool {
	return true
}

func (tvc testVideoConnection) Close() error {
	return nil
}
