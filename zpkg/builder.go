/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package zpkg

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/fezz-io/zps/phase"

	"github.com/fezz-io/zps/zps"

	"context"

	"github.com/chuckpreslar/emission"
	"github.com/fezz-io/zps/action"
	"github.com/fezz-io/zps/provider"
	"github.com/fezz-io/zps/zpkg/payload"
)

type Builder struct {
	*emission.Emitter

	options *provider.Options

	zpfPath    string
	workPath   string
	outputPath string

	version uint8

	manifest *action.Manifest

	filename string

	header  *Header
	payload *payload.Writer

	writer *Writer
}

func NewBuilder() *Builder {
	builder := &Builder{Emitter: emission.NewEmitter()}

	builder.version = Version

	builder.options = &provider.Options{}

	builder.manifest = action.NewManifest()

	builder.header = NewHeader(Version, Compression)
	builder.payload = payload.NewWriter("", 0)
	builder.writer = NewWriter()

	return builder
}

func (b *Builder) ZpfPath(zp string) *Builder {
	b.zpfPath = zp
	return b
}

func (b *Builder) TargetPath(tp string) *Builder {
	b.options.TargetPath = tp
	return b
}

func (b *Builder) Restrict(r bool) *Builder {
	b.options.Restrict = r
	return b
}

func (b *Builder) Secure(s bool) *Builder {
	b.options.Secure = s
	return b
}

func (b *Builder) WorkPath(wp string) *Builder {
	b.workPath = wp
	b.payload.WorkPath = wp
	return b
}

func (b *Builder) OutputPath(op string) *Builder {
	b.outputPath = op
	return b
}

func (b *Builder) Version(version uint8) *Builder {
	b.version = version
	b.header.Version = version
	return b
}

// Set default paths
func (b *Builder) setPaths() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Can be avoided if we error on create
	// but it would break the builder pattern
	if b.zpfPath == "" {
		b.zpfPath = path.Join(wd, DefaultZpfPath)
	}

	// If the path is a directory append the default
	// ZpfPath
	stat, err := os.Stat(b.zpfPath)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		b.zpfPath = path.Join(b.zpfPath, DefaultZpfPath)
	}

	if b.options.TargetPath == "" {
		dir, _ := path.Split(b.zpfPath)
		b.options.TargetPath = path.Join(dir, DefaultTargetDir)
	}
	if b.workPath == "" {
		b.workPath = wd
		b.payload.WorkPath = wd
	}
	if b.outputPath == "" {
		b.outputPath = wd
	}

	return err
}

// This isn't efficient, I don't expect these files to be terribly large however
func (b *Builder) loadZpkgfile() error {
	zpkgFile, err := (&ZpkgFile{}).Load(b.zpfPath)
	if err != nil {
		return err
	}

	b.manifest, err = zpkgFile.Eval()
	if err != nil {
		return err
	}

	return err
}

// Process options deal with any special cases here
func (b *Builder) processOptions() error {

	return nil
}

// Add FS objects
func (b *Builder) resolve() error {
	// If restrict is set don't walk the target path
	// this will result in only defined file system objects being added
	// to the package
	if b.options.Restrict == true {
		return nil
	}

	err := filepath.Walk(b.options.TargetPath, func(path string, f os.FileInfo, err error) error {
		objectPath := strings.Replace(path, b.options.TargetPath+string(os.PathSeparator), "", 1)

		if objectPath != b.options.TargetPath {
			if f.IsDir() {
				var dir = action.NewDir()
				dir.Path = objectPath

				b.manifest.Add(dir)
			}

			if f.Mode().IsRegular() {
				var file = action.NewFile()
				file.Path = objectPath

				b.manifest.Add(file)
			}

			if f.Mode()&os.ModeSymlink == os.ModeSymlink {
				var symlink = action.NewSymLink()
				symlink.Path = objectPath

				b.manifest.Add(symlink)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return b.manifest.Validate()
}

// Set file name and zpkg timestamp
func (b *Builder) set() error {
	pkg, err := zps.NewPkgFromManifest(b.manifest)
	if err != nil {
		return err
	}

	pkg.Version().Timestamp = time.Now()
	b.manifest.Zpkg.Version = pkg.Version().String()

	b.filename = pkg.FileName()

	return nil
}

// Completes manifest, builds payload
func (b *Builder) realize() error {
	var err error

	// Setup context
	ctx := context.WithValue(context.Background(), "options", b.options)
	ctx = context.WithValue(ctx, "phase", phase.PACKAGE)
	ctx = context.WithValue(ctx, "payload", b.payload)

	factory := provider.DefaultFactory(b.Emitter)

	for _, act := range b.manifest.Actions() {
		err = factory.Get(act).Realize(ctx)
		if err != nil {
			return err
		}
	}

	return err
}

func (b *Builder) Build() (string, error) {
	err := b.setPaths()
	if err != nil {
		return "", err
	}

	err = b.loadZpkgfile()
	if err != nil {
		return "", err
	}

	err = b.processOptions()
	if err != nil {
		return "", err
	}

	err = b.resolve()
	if err != nil {
		return "", err
	}

	err = b.set()
	if err != nil {
		return "", err
	}

	err = b.realize()
	if err != nil {
		return "", err
	}

	// Write the file
	err = b.writer.Write(b.filename, b.header, b.manifest, b.payload)
	if err != nil {
		return "", err
	}

	b.Emit("builder.complete", b.filename)

	return b.filename, err
}
