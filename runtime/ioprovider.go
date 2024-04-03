package runtime

import (
	"io"
	"os"
	"strings"

	"github.com/Frank-Mayer/ohmygosh/iohelper"
)

type IoProvider struct {
	DefaultOut io.WriteCloser
	DefaultErr io.WriteCloser
	DefaultIn  io.Reader
	Closer     *iohelper.Closer
}

func DefaultIoProvider() *IoProvider {
	return &IoProvider{
		DefaultOut: iohelper.WrapWriteFakeCloser(os.Stdout),
		DefaultErr: iohelper.WrapWriteFakeCloser(os.Stderr),
		DefaultIn:  os.Stdin,
		Closer:     iohelper.NewCloser(),
	}
}

func TestIoProvider(stdin string) (*IoProvider, *strings.Builder, *strings.Builder) {
	outSB := &strings.Builder{}
	outW := iohelper.WrapWriteFakeCloser(outSB)
	errSB := &strings.Builder{}
	errW := iohelper.WrapWriteFakeCloser(errSB)
	var inR io.Reader
	if stdin == "" {
		inR = os.Stdin
	} else {
		inR = strings.NewReader(stdin)
	}
	return &IoProvider{
		DefaultOut: outW,
		DefaultErr: errW,
		DefaultIn:  inR,
		Closer:     iohelper.NewCloser(),
	}, outSB, errSB
}

func SubshellIoProvider(parent *IoProvider) (*IoProvider, *strings.Builder) {
	sb := &strings.Builder{}
	w := iohelper.WrapWriteFakeCloser(sb)
	return &IoProvider{
		DefaultOut: w,
		DefaultErr: parent.DefaultErr,
		DefaultIn:  parent.DefaultIn,
		Closer:     iohelper.NewCloser(),
	}, sb
}

func (i *IoProvider) Close() {
	i.Closer.Close()
}
