package embedded

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	tor "github.com/clementauger/tor-prebuilt/embedded/tor_latest"
	"github.com/cretz/bine/process"
)

func NewCreator() process.Creator {
	return &exeProcessCreator{}
}

type exeProcessCreator struct{}

type exeProcess struct {
	*exec.Cmd
}

func (e *exeProcessCreator) New(ctx context.Context, args ...string) (process.Process, error) {

	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}

	assets := tor.AssetNames()
	for _, name := range assets {
		dst := filepath.Join(tmp, name)
		err := ioutil.WriteFile(dst, tor.MustAsset(name), os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	exe := filepath.Join(tmp, "tor")
	if strings.Contains(runtime.GOOS, "windows") {
		exe += ".exe"
	}
	log.Println(exe, args)
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Dir = tmp
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return &exeProcess{cmd}, nil
}

func (e *exeProcess) EmbeddedControlConn() (net.Conn, error) {
	return nil, process.ErrControlConnUnsupported
}
